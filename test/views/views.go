package views

//import (
//	"encoding/gob"
//	"encoding/hex"
//	"github.com/COMTOP1/AFC-GO/affiliation"
//	"github.com/COMTOP1/AFC-GO/document"
//	"github.com/COMTOP1/AFC-GO/image"
//	"github.com/COMTOP1/AFC-GO/news"
//	"github.com/COMTOP1/AFC-GO/player"
//	"github.com/COMTOP1/AFC-GO/sponsor"
//	"github.com/COMTOP1/AFC-GO/team"
//	"github.com/COMTOP1/AFC-GO/whatson"
//	"github.com/gorilla/securecookie"
//	"github.com/gorilla/sessions"
//	"time"
//)
//
//type (
//	Config struct {
//		Address           string
//		DatabaseURL       string
//		BaseDomainName    string
//		DomainName        string
//		LogoutEndpoint    string
//		SessionCookieName string
//		//Mail              SMTPConfig
//		Security SecurityConfig
//	}
//
//	//// SMTPConfig stores the SMTP Mailer configuration
//	//SMTPConfig struct {
//	//	Host       string
//	//	Username   string
//	//	Password   string
//	//	Port       int
//	//	DomainName string
//	//}
//
//	// SecurityConfig stores the security configuration
//	SecurityConfig struct {
//		EncryptionKey     string
//		AuthenticationKey string
//		SigningKey        string
//	}
//
//	// Views encapsulates our view dependencies
//	Views struct {
//		//affiliation *affiliation.Store
//		//cache       *cache.Cache
//		conf *Config
//		//cookie      *sessions.CookieStore
//		//document    *document.Store
//		//image       *image.Store
//		//news        *news.Store
//		//player      *player.Store
//		//sponsor     *sponsor.Store
//		//team        *team.Store
//		template *templates.Templater
//		//user        *user.Store
//		//whatsOn     *whatson.Store
//	}
//
//	TemplateHelper struct {
//		UserPermissions []role.Role
//		ActivePage      string
//		Assumed         bool
//	}
//)
//
//func New(conf *Config, host string) *Views {
//	v := &Views{}
//	// Connecting to stores
//	dbStore := db.NewStore(conf.DatabaseURL, host)
//	v.affiliation = affiliation.NewAffiliationRepo(dbStore)
//	v.document = document.NewDocumentRepo(dbStore)
//	v.image = image.NewImageRepo(dbStore)
//	v.news = news.NewNewsRepo(dbStore)
//	v.player = player.NewPlayerRepo(dbStore)
//	v.sponsor = sponsor.NewSponsorRepo(dbStore)
//	v.team = team.NewTeamRepo(dbStore)
//	v.user = user.NewUserRepo(dbStore)
//	v.whatsOn = whatson.NewWhatsOnRepo(dbStore)
//
//	v.template = templates.NewTemplate(nil)
//
//	// Initialising cache
//	v.cache = cache.New(1*time.Hour, 1*time.Hour)
//
//	// Initialise mailer
//	//v.mailer = mail.NewMailer(mail.Config{
//	//	Host:       conf.Mail.Host,
//	//	Port:       conf.Mail.Port,
//	//	Username:   conf.Mail.Username,
//	//	Password:   conf.Mail.Password,
//	//	DomainName: conf.Mail.DomainName,
//	//})
//
//	// Initialising session cookie
//	authKey, _ := hex.DecodeString(conf.Security.AuthenticationKey)
//	if len(authKey) == 0 {
//		authKey = securecookie.GenerateRandomKey(64)
//	}
//	encryptionKey, _ := hex.DecodeString(conf.Security.EncryptionKey)
//	if len(encryptionKey) == 0 {
//		encryptionKey = securecookie.GenerateRandomKey(32)
//	}
//	v.cookie = sessions.NewCookieStore(
//		authKey,
//		encryptionKey,
//	)
//	v.cookie.Options = &sessions.Options{
//		MaxAge:   60 * 60 * 24,
//		HttpOnly: true,
//		Path:     "/",
//	}
//
//	// So we can use our struct in the cookie
//	gob.Register(user.User{})
//	//gob.Register(InternalContext{})
//
//	v.conf = conf
//
//	// Struct validator
//	//v.validate = validator.New()
//
//	//go func() {
//	//	for {
//	//		err := v.api.DeleteOldToken(context.Background())
//	//		if err != nil {
//	//			log.Printf("failed to delete old token func: %+v", err)
//	//		}
//	//		time.Sleep(30 * time.Second)
//	//	}
//	//}()
//
//	return v
//}
