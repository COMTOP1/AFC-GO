package views

import (
	"context"
	"encoding/gob"
	"encoding/hex"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/patrickmn/go-cache"

	"github.com/COMTOP1/AFC-GO/affiliation"
	"github.com/COMTOP1/AFC-GO/document"
	"github.com/COMTOP1/AFC-GO/image"
	"github.com/COMTOP1/AFC-GO/infrastructure/db"
	"github.com/COMTOP1/AFC-GO/infrastructure/mail"
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

type (
	Config struct {
		Address           string
		DatabaseURL       string
		DomainName        string
		SessionCookieName string
		FileDir           string
		Mail              SMTPConfig
		Security          SecurityConfig
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
	v.team = team.NewTeamRepo(dbStore)
	v.user = user.NewUserRepo(dbStore)
	v.whatsOn = whatson.NewWhatsOnRepo(dbStore)

	v.template = templates.NewTemplate(v.team)

	// Initialising cache
	v.cache = cache.New(1*time.Hour, 1*time.Hour)

	// Initialising session cookie
	authKey, err := hex.DecodeString(conf.Security.AuthenticationKey)
	if err != nil {
		log.Printf("failed to decode authentication key: %+v", err)
	}
	if len(authKey) == 0 {
		authKey = securecookie.GenerateRandomKey(64)
	}
	encryptionKey, err := hex.DecodeString(conf.Security.EncryptionKey)
	if err != nil {
		log.Printf("failed to decode encryption key: %+v", err)
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

	if countToFlush == 0 {
		return
	}

	v.cache.Set(visitorCount, countToFlush, cache.DefaultExpiration)

	ctx := context.Background()
	currentSetting, err := v.setting.GetSetting(ctx, visitorCount)
	if err != nil {
		_, err = v.setting.AddSetting(ctx, setting.Setting{
			ID:          visitorCount,
			SettingText: strconv.Itoa(countToFlush),
		})
		if err != nil {
			log.Printf("Error creating visitorCount: %v", err)
		}
		return
	}

	currentValue, _ := strconv.Atoi(currentSetting.SettingText)
	newValue := currentValue + countToFlush

	_, err = v.setting.EditSetting(ctx, setting.Setting{
		ID:          visitorCount,
		SettingText: strconv.Itoa(newValue),
	})
	if err != nil {
		log.Printf("Error updating visitorCount: %v", err)
	}

	v.cache.Set(visitorCount, newValue, cache.DefaultExpiration)
}

func (v *Views) Stop() {
	close(v.stopChan)
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
