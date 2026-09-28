// Package app wires the stores, services, legacy views and API handlers into
// one Echo server.
package app

import (
	"context"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho" //nolint:staticcheck // still functional; github.com/labstack/echo-opentelemetry is its replacement but isn't part of opentelemetry-go-contrib

	"github.com/COMTOP1/AFC-GO/server/internal/affiliation"
	"github.com/COMTOP1/AFC-GO/server/internal/auth"
	_ "github.com/COMTOP1/AFC-GO/server/internal/docs" // registers the swagger spec
	"github.com/COMTOP1/AFC-GO/server/internal/document"
	"github.com/COMTOP1/AFC-GO/server/internal/image"
	infradb "github.com/COMTOP1/AFC-GO/server/internal/infrastructure/db"
	"github.com/COMTOP1/AFC-GO/server/internal/infrastructure/mail"
	"github.com/COMTOP1/AFC-GO/server/internal/infrastructure/storage"
	"github.com/COMTOP1/AFC-GO/server/internal/legacy"
	"github.com/COMTOP1/AFC-GO/server/internal/legacy/views"
	"github.com/COMTOP1/AFC-GO/server/internal/news"
	"github.com/COMTOP1/AFC-GO/server/internal/player"
	"github.com/COMTOP1/AFC-GO/server/internal/programme"
	"github.com/COMTOP1/AFC-GO/server/internal/setting"
	"github.com/COMTOP1/AFC-GO/server/internal/sponsor"
	"github.com/COMTOP1/AFC-GO/server/internal/team"
	"github.com/COMTOP1/AFC-GO/server/internal/upload"
	"github.com/COMTOP1/AFC-GO/server/internal/user"
	"github.com/COMTOP1/AFC-GO/server/internal/visitors"
	"github.com/COMTOP1/AFC-GO/server/internal/web"
	"github.com/COMTOP1/AFC-GO/server/internal/whatson"
)

// Config is everything the server needs, loaded from the environment by main.
type Config struct {
	Address      string
	DomainName   string
	DatabaseURL  string
	DatabaseHost string
	Version      string
	Session      auth.Config
	S3           storage.Config
	Mail         mail.Config
	Passwords    auth.PasswordConfig
	Redis        auth.RedisConfig
}

// Stores are the database repositories.
type Stores struct {
	Affiliation *affiliation.Store
	Document    *document.Store
	Image       *image.Store
	News        *news.Store
	Player      *player.Store
	Programme   *programme.Store
	Setting     *setting.Store
	Sponsor     *sponsor.Store
	Team        *team.Store
	User        *user.Store
	WhatsOn     *whatson.Store
}

// NewStores builds every repository over one connection pool.
func NewStores(db *sqlx.DB) Stores {
	return Stores{
		Affiliation: affiliation.NewAffiliationRepo(db),
		Document:    document.NewDocumentRepo(db),
		Image:       image.NewImageRepo(db),
		News:        news.NewNewsRepo(db),
		Player:      player.NewPlayerRepo(db),
		Programme:   programme.NewProgrammeRepo(db),
		Setting:     setting.NewSettingRepo(db),
		Sponsor:     sponsor.NewSponsorRepo(db),
		Team:        team.NewTeamRepo(db),
		User:        user.NewUserRepo(db),
		WhatsOn:     whatson.NewWhatsOnRepo(db),
	}
}

// App is a wired server.
type App struct {
	Echo     *echo.Echo
	address  string
	visitors *visitors.Counter
	tokens   *auth.Tokens
}

// New connects to Postgres and S3 and builds the server.
func New(conf Config) *App {
	db := infradb.NewStore(conf.DatabaseURL, conf.DatabaseHost)
	objects := storage.NewStore(context.Background(), conf.S3)
	return Build(conf, NewStores(db), objects, mail.NewMailer(conf.Mail))
}

// Build wires everything without connecting to anything, so tests can pass
// fakes. Background work starts in Start.
func Build(conf Config, s Stores, objects upload.Storage, mailer *mail.MailerInit) *App {
	uploads := upload.New(objects)
	newsSvc := news.NewService(s.News, uploads)
	whatsOnSvc := whatson.NewService(s.WhatsOn, uploads)
	sponsorSvc := sponsor.NewService(s.Sponsor, s.Team, uploads)
	affiliationSvc := affiliation.NewService(s.Affiliation, uploads)
	documentSvc := document.NewService(s.Document, uploads)
	gallerySvc := image.NewService(s.Image, uploads)
	programmeSvc := programme.NewService(s.Programme, uploads)
	sessions := auth.NewSessions(conf.Session, s.User)
	tokens := auth.NewTokens(conf.Redis)
	counter := visitors.New(s.Setting, 30*time.Second, "afcaldermaston.co.uk")

	legacyViews := views.New(views.Deps{
		Conf: &views.Config{
			DomainName:        conf.DomainName,
			SessionCookieName: sessions.Name(),
			Security:          conf.Passwords,
		},
		Sessions:           sessions.CookieStore(),
		Storage:            objects,
		Mailer:             mailer,
		Visitors:           counter,
		Tokens:             tokens,
		Affiliation:        s.Affiliation,
		AffiliationService: affiliationSvc,
		Document:           s.Document,
		DocumentService:    documentSvc,
		GalleryService:     gallerySvc,
		Image:              s.Image,
		News:               s.News,
		NewsService:        newsSvc,
		Player:             s.Player,
		Programme:          s.Programme,
		ProgrammeService:   programmeSvc,
		Setting:            s.Setting,
		Sponsor:            s.Sponsor,
		SponsorService:     sponsorSvc,
		Team:               s.Team,
		User:               s.User,
		WhatsOn:            s.WhatsOn,
		WhatsOnService:     whatsOnSvc,
	})

	e := echo.New()
	e.HideBanner = true
	e.Pre(middleware.RemoveTrailingSlash())
	e.Use(middleware.Recover())
	e.Use(otelecho.Middleware("afc-go", otelecho.WithSkipper(func(c echo.Context) bool {
		return c.Path() == "/api/health" || c.Path() == "/api/v1/health"
	})))
	e.Use(middleware.BodyLimit("15M"))
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Level: 5,
		Skipper: func(c echo.Context) bool {
			// File downloads are redirects to the CDN; don't gzip them.
			return strings.HasPrefix(c.Path(), "/download") || strings.HasPrefix(c.Path(), "/api/v1/files")
		},
	}))
	e.Use(counter.Middleware)
	e.HTTPErrorHandler = web.ErrorHandler(legacyViews.CustomHTTPErrorHandler)

	legacy.Mount(e, legacyViews)

	api := web.NewAPI(e, conf.Session.Secure)
	guards := sessions.Guards()
	auth.NewHandlers(sessions, uploads).Register(api, guards)
	news.NewHandlers(newsSvc).Register(api, guards)
	whatson.NewHandlers(whatsOnSvc).Register(api, guards)
	sponsor.NewHandlers(sponsorSvc).Register(api, guards)
	affiliation.NewHandlers(affiliationSvc).Register(api, guards)
	document.NewHandlers(documentSvc).Register(api, guards)
	image.NewHandlers(gallerySvc).Register(api, guards)
	programme.NewHandlers(programmeSvc).Register(api, guards)

	return &App{Echo: e, address: conf.Address, visitors: counter, tokens: tokens}
}

// Start begins background work and serves until the server stops.
func (a *App) Start() error {
	a.visitors.Start(context.Background())
	return a.Echo.Start(a.address)
}

// Stop ends background work.
func (a *App) Stop() {
	a.visitors.Stop()
	a.tokens.Close()
}
