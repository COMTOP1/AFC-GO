package views

import (
	"context"
	"crypto/tls"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log/slog"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/patrickmn/go-cache"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"

	"github.com/COMTOP1/AFC-GO/affiliation"
	"github.com/COMTOP1/AFC-GO/document"
	"github.com/COMTOP1/AFC-GO/image"
	"github.com/COMTOP1/AFC-GO/infrastructure/db"
	"github.com/COMTOP1/AFC-GO/infrastructure/mail"
	"github.com/COMTOP1/AFC-GO/infrastructure/storage"
	"github.com/COMTOP1/AFC-GO/news"
	"github.com/COMTOP1/AFC-GO/player"
	"github.com/COMTOP1/AFC-GO/programme"
	"github.com/COMTOP1/AFC-GO/role"
	"github.com/COMTOP1/AFC-GO/setting"
	"github.com/COMTOP1/AFC-GO/sponsor"
	"github.com/COMTOP1/AFC-GO/team"
	"github.com/COMTOP1/AFC-GO/templates"
	"github.com/COMTOP1/AFC-GO/user"
	"github.com/COMTOP1/AFC-GO/whatson"
)

const visitorCount = "visitorCount"

var tracer = otel.Tracer("github.com/COMTOP1/AFC-GO/views")

type (
	Config struct {
		Address           string
		DatabaseURL       string
		DomainName        string
		SessionCookieName string
		S3                storage.Config
		Mail              SMTPConfig
		Security          SecurityConfig
		Redis             RedisConfig
	}

	// RedisConfig stores the configuration for an optional Redis/Valkey
	// cache, used to share state (e.g. password reset tokens) across
	// multiple app instances. When Addresses is empty, an in-process cache
	// is used instead, which is only suitable for a single instance.
	RedisConfig struct {
		// Addresses is one or more "host:port" pairs. More than one
		// address puts the client into cluster mode, unless MasterName
		// is set, in which case it uses Sentinel-based failover.
		Addresses  []string
		MasterName string
		Username   string
		Password   string
		DB         int
		TLS        bool
		// KeyPrefix namespaces every key this app writes to Redis/Valkey,
		// e.g. "afc:dev:" or "afc:prod:". Set it differently per
		// environment so a Valkey ACL user can be scoped to only that
		// environment's keys (~afc:dev:* / ~afc:prod:*) on a shared
		// instance or cluster. Defaults to "afc:" if empty.
		KeyPrefix string
	}

	// SMTPConfig stores the SMTP Mailer configuration
	SMTPConfig struct {
		Host     string
		Username string
		Password string
		Port     int
	}

	// SecurityConfig stores the security configuration
	SecurityConfig struct {
		EncryptionKey           string
		AuthenticationKey       string
		Iterations              int
		ScryptWorkFactor        int
		ScryptBlockSize         int
		ScryptParallelismFactor int
		KeyLength               int
	}

	// Views encapsulates our view dependencies
	Views struct {
		affiliation *affiliation.Store
		cache       *cache.Cache
		redis       redis.UniversalClient
		redisPrefix string
		conf        *Config
		cookie      *sessions.CookieStore
		document    *document.Store
		image       *image.Store
		mailer      *mail.MailerInit
		news        *news.Store
		player      *player.Store
		programme   *programme.Store
		setting     *setting.Store
		sponsor     *sponsor.Store
		storage     *storage.Store
		team        *team.Store
		template    *templates.Templater
		user        *user.Store
		whatsOn     *whatson.Store

		// Visitor tracking
		count         int
		countMutex    sync.Mutex
		flushInterval time.Duration
		stopChan      chan struct{}
	}

	TemplateHelper struct {
		UserPermissions []role.Role
		ActivePage      string
		Assumed         bool
	}
)

func New(conf *Config, host string, interval time.Duration) *Views {
	v := &Views{}
	// Connecting to stores
	dbStore := db.NewStore(conf.DatabaseURL, host)
	v.affiliation = affiliation.NewAffiliationRepo(dbStore)
	v.document = document.NewDocumentRepo(dbStore)
	v.image = image.NewImageRepo(dbStore)
	v.news = news.NewNewsRepo(dbStore)
	v.player = player.NewPlayerRepo(dbStore)
	v.programme = programme.NewProgrammeRepo(dbStore)
	v.setting = setting.NewSettingRepo(dbStore)
	v.sponsor = sponsor.NewSponsorRepo(dbStore)
	v.storage = storage.NewStore(conf.S3)
	v.team = team.NewTeamRepo(dbStore)
	v.user = user.NewUserRepo(dbStore)
	v.whatsOn = whatson.NewWhatsOnRepo(dbStore)

	v.template = templates.NewTemplate(v.team, v.storage)

	// Initialising cache
	v.cache = cache.New(1*time.Hour, 1*time.Hour)

	// Initialising Redis/Valkey client, used to share state such as password
	// reset tokens across instances. Falls back to the in-process cache
	// above when no addresses are configured, e.g. for local development.
	if len(conf.Redis.Addresses) > 0 {
		opts := &redis.UniversalOptions{
			Addrs:      conf.Redis.Addresses,
			MasterName: conf.Redis.MasterName,
			Username:   conf.Redis.Username,
			Password:   conf.Redis.Password,
			DB:         conf.Redis.DB,
		}
		if conf.Redis.TLS {
			opts.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12}
		}
		v.redis = redis.NewUniversalClient(opts)
	}
	v.redisPrefix = conf.Redis.KeyPrefix
	if v.redisPrefix == "" {
		v.redisPrefix = "afc:"
	}

	// Initialising session cookie
	authKey, err := hex.DecodeString(conf.Security.AuthenticationKey)
	if err != nil {
		slog.Info(fmt.Sprintf("failed to decode authentication key: %+v", err))
	}
	if len(authKey) == 0 {
		authKey = securecookie.GenerateRandomKey(64)
	}
	encryptionKey, err := hex.DecodeString(conf.Security.EncryptionKey)
	if err != nil {
		slog.Info(fmt.Sprintf("failed to decode encryption key: %+v", err))
	}
	if len(encryptionKey) == 0 {
		encryptionKey = securecookie.GenerateRandomKey(32)
	}
	v.cookie = sessions.NewCookieStore(
		authKey,
		encryptionKey,
	)
	v.cookie.Options = &sessions.Options{
		MaxAge:   60 * 60 * 24,
		HttpOnly: true,
		Path:     "/",
	}

	// So we can use our struct in the cookie
	gob.Register(user.User{})
	gob.Register(InternalContext{})

	v.conf = conf

	// Initialise mailer
	v.mailer = mail.NewMailer(mail.Config{
		Host:     conf.Mail.Host,
		Port:     conf.Mail.Port,
		Username: conf.Mail.Username,
		Password: conf.Mail.Password,
	})

	v.flushInterval = interval
	v.stopChan = make(chan struct{})

	// Seed the local visitor count cache from the DB so this instance doesn't
	// report 0 visitors until its first flush.
	v.refreshVisitorCountCache(context.Background())

	go v.startFlusher()

	return v
}

func (v *Views) RecordVisit(visitorID string) {
	if _, found := v.cache.Get(visitorID); !found {
		v.cache.Set(visitorID, true, cache.DefaultExpiration)

		v.countMutex.Lock()
		v.count++
		v.countMutex.Unlock()
	}
}

func (v *Views) startFlusher() {
	ticker := time.NewTicker(v.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			v.flushToDB()
		case <-v.stopChan:
			return
		}
	}
}

func (v *Views) flushToDB() {
	v.countMutex.Lock()
	countToFlush := v.count
	v.count = 0
	v.countMutex.Unlock()

	ctx := context.Background()

	if countToFlush == 0 {
		// Nothing local to add, but another instance may have flushed visits
		// of its own since we last checked - keep our cached total in sync.
		v.refreshVisitorCountCache(ctx)
		return
	}

	newSetting, err := v.setting.IncrementSetting(ctx, visitorCount, countToFlush)
	if err != nil {
		slog.Info(fmt.Sprintf("Error incrementing visitorCount: %v", err))
		return
	}

	newValue, err := strconv.Atoi(newSetting.SettingText)
	if err != nil {
		slog.Info(fmt.Sprintf("Error parsing visitorCount: %v", err))
		return
	}

	v.cache.Set(visitorCount, newValue, cache.DefaultExpiration)
}

// refreshVisitorCountCache re-reads the visitor count from the DB and caches
// it locally, so this instance reflects visits recorded by other instances.
func (v *Views) refreshVisitorCountCache(ctx context.Context) {
	currentSetting, err := v.setting.GetSetting(ctx, visitorCount)
	if err != nil {
		// Not created yet - it will be on the first increment.
		return
	}

	currentValue, err := strconv.Atoi(currentSetting.SettingText)
	if err != nil {
		log.Printf("Error parsing visitorCount: %v", err)
		return
	}

	v.cache.Set(visitorCount, currentValue, cache.DefaultExpiration)
}

func (v *Views) Stop() {
	close(v.stopChan)
	if v.redis != nil {
		if err := v.redis.Close(); err != nil {
			log.Printf("failed to close redis client: %+v", err)
		}
	}
}

// resetTokenKey returns the fully-namespaced Redis/Valkey key for a reset
// token, scoped under this instance's configured environment prefix (see
// RedisConfig.KeyPrefix).
func (v *Views) resetTokenKey(token string) string {
	return v.redisPrefix + "reset-token:" + token
}

// SetResetToken stores a mapping of a one-time password reset token to a
// user ID, expiring after ttl. When Redis is configured this is shared
// across every app instance; otherwise it only lives in this instance's
// memory, which requires sticky sessions/single-instance deployment for the
// reset flow to work reliably.
func (v *Views) SetResetToken(ctx context.Context, token string, userID int, ttl time.Duration) error {
	if v.redis != nil {
		return v.redis.Set(ctx, v.resetTokenKey(token), userID, ttl).Err()
	}
	v.cache.Set(token, userID, ttl)
	return nil
}

// GetResetToken looks up the user ID a password reset token was issued for.
func (v *Views) GetResetToken(ctx context.Context, token string) (int, bool) {
	if v.redis != nil {
		userID, err := v.redis.Get(ctx, v.resetTokenKey(token)).Int()
		if err != nil {
			return 0, false
		}
		return userID, true
	}
	val, found := v.cache.Get(token)
	if !found {
		return 0, false
	}
	return val.(int), true
}

// DeleteResetToken invalidates a password reset token after use.
func (v *Views) DeleteResetToken(ctx context.Context, token string) {
	if v.redis != nil {
		if err := v.redis.Del(ctx, v.resetTokenKey(token)).Err(); err != nil {
			log.Printf("failed to delete reset token from redis: %+v", err)
		}
		return
	}
	v.cache.Delete(token)
}

func (v *Views) GetVisitorCount() int {
	val, ok := v.cache.Get(visitorCount)
	if !ok {
		return 0
	}

	count, ok := val.(int)
	if !ok {
		return 0
	}
	return count
}
