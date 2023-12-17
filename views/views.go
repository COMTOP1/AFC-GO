package views

import (
	"encoding/gob"
	"encoding/hex"
	"log"
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
	"github.com/COMTOP1/AFC-GO/sponsor"
	"github.com/COMTOP1/AFC-GO/team"
	"github.com/COMTOP1/AFC-GO/templates"
	"github.com/COMTOP1/AFC-GO/user"
	"github.com/COMTOP1/AFC-GO/whatson"
)

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
		sponsor     *sponsor.Store
		team        *team.Store
		template    *templates.Templater
		user        *user.Store
		whatsOn     *whatson.Store
	}

	TemplateHelper struct {
		UserPermissions []role.Role
		ActivePage      string
		Assumed         bool
	}
)

func New(conf *Config, host string) *Views {
	v := &Views{}
	// Connecting to stores
	dbStore := db.NewStore(conf.DatabaseURL, host)
	v.affiliation = affiliation.NewAffiliationRepo(dbStore)
	v.document = document.NewDocumentRepo(dbStore)
	v.image = image.NewImageRepo(dbStore)
	v.news = news.NewNewsRepo(dbStore)
	v.player = player.NewPlayerRepo(dbStore)
	v.programme = programme.NewProgrammeRepo(dbStore)
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

	return v
}
