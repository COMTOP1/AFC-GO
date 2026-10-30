// Package app wires the stores, services and API handlers into one Echo
// server, and serves the web client.
package app

import (
	"context"
	"io/fs"
	"net"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/jmoiron/sqlx"
	otelecho "github.com/labstack/echo-otel/v5"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	// "go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho" //nolint:staticcheck // still functional; github.com/labstack/echo-opentelemetry is its replacement but isn't part of opentelemetry-go-contrib

	"github.com/COMTOP1/AFC-GO/server/internal/account"
	"github.com/COMTOP1/AFC-GO/server/internal/affiliation"
	"github.com/COMTOP1/AFC-GO/server/internal/auth"
	"github.com/COMTOP1/AFC-GO/server/internal/clienttelemetry"
	_ "github.com/COMTOP1/AFC-GO/server/internal/docs" // registers the swagger spec
	"github.com/COMTOP1/AFC-GO/server/internal/document"
	"github.com/COMTOP1/AFC-GO/server/internal/emails"
	"github.com/COMTOP1/AFC-GO/server/internal/filecheck"
	"github.com/COMTOP1/AFC-GO/server/internal/files"
	"github.com/COMTOP1/AFC-GO/server/internal/image"
	infradb "github.com/COMTOP1/AFC-GO/server/internal/infrastructure/db"
	"github.com/COMTOP1/AFC-GO/server/internal/infrastructure/mail"
	"github.com/COMTOP1/AFC-GO/server/internal/infrastructure/mtls"
	"github.com/COMTOP1/AFC-GO/server/internal/infrastructure/storage"
	"github.com/COMTOP1/AFC-GO/server/internal/news"
	"github.com/COMTOP1/AFC-GO/server/internal/player"
	"github.com/COMTOP1/AFC-GO/server/internal/programme"
	"github.com/COMTOP1/AFC-GO/server/internal/setting"
	"github.com/COMTOP1/AFC-GO/server/internal/site"
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
	// UI is the built React client (index.html + assets/); nil means not built.
	UI fs.FS
	// UIProxy, when set, proxies every non-API page to the Vite dev server instead of UI.
	UIProxy *url.URL
	// MTLS, when set, serves over TLS and only accepts requests from allowed client certificates.
	MTLS *mtls.Server
	// TrustedProxies are the reverse proxies whose X-Forwarded-For is
	// believed. Empty means the connecting address is the client's.
	TrustedProxies []*net.IPNet
	// FileRetention is how long a file may go unreferenced before the daily
	// file check deletes it; zero means filecheck.MaxRetention (30 days).
	FileRetention time.Duration
}

// The real file store must support the daily file check, or Build would
// quietly leave it switched off.
var _ filecheck.Objects = (*storage.Store)(nil)

// Stores are the database repositories.
type Stores struct {
	Affiliation *affiliation.Store
	Document    *document.Store
	FileCheck   *filecheck.Store
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
		FileCheck:   filecheck.NewStore(db),
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
	mtls     *mtls.Server
	// files is the daily file check; nil when the file store can't be listed
	// (tests).
	files *filecheck.Job
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
	teamSvc := team.NewService(s.Team, uploads, s.Player, s.Sponsor, s.User).WithSquadAges(s.Player)
	playerSvc := player.NewService(s.Player, s.Team, uploads)
	fileSvc := files.NewService(uploads, files.Resolvers(files.Sources{
		Affiliation: s.Affiliation, Document: s.Document, Image: s.Image, News: s.News,
		Programme: s.Programme, Sponsor: s.Sponsor, Team: s.Team, User: s.User, WhatsOn: s.WhatsOn,
		Players: playerSvc,
	}))
	settingSvc := setting.NewService(s.Setting)
	sessions := auth.NewSessions(conf.Session, s.User)
	tokens := auth.NewTokens(conf.Redis)
	authSvc := auth.NewService(s.User, tokens, conf.Passwords)
	accountSvc := account.NewService(s.User, uploads)
	counter := visitors.New(s.Setting, 30*time.Second, "afcaldermaston.co.uk")

	sender := emails.NewSMTP(mailer)
	userSvc := user.NewService(s.User, s.Team, uploads, sender, tokens,
		user.HashParams{
			WorkFactor:  conf.Passwords.ScryptWorkFactor,
			BlockSize:   conf.Passwords.ScryptBlockSize,
			Parallelism: conf.Passwords.ScryptParallelismFactor,
			KeyLength:   conf.Passwords.KeyLength,
		}, conf.DomainName)

	siteSvc := site.NewService(site.Deps{
		News: newsSvc, WhatsOn: whatsOnSvc, Sponsors: sponsorSvc, Affiliations: affiliationSvc,
		Teams: teamSvc, Players: playerSvc, Settings: settingSvc, Users: s.User, Visitors: counter,
		Files: uploads, Version: conf.Version,
	})

	e := echo.New()
	e.IPExtractor = web.IPExtractor(conf.TrustedProxies)
	e.Pre(middleware.RemoveTrailingSlash())
	if conf.MTLS != nil {
		e.Pre(conf.MTLS.Middleware("/api/health", "/api/v1/health"))
	}

	otelMiddleware := otelecho.NewMiddlewareWithConfig(otelecho.Config{
		Skipper: func(c *echo.Context) bool {
			return c.Path() == "/api/health" || c.Path() == "/api/v1/health"
		},
	})

	e.Use(middleware.Recover())
	e.Use(web.SecurityHeaders(conf.Session.Secure, conf.UIProxy == nil, web.Origin(conf.S3.Endpoint)))
	e.Use(otelMiddleware)
	e.Use(web.RedactSpanPath)
	e.Use(middleware.BodyLimit(15 * 1024 * 1024)) // 15MB
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Level: 5,
		Skipper: func(c *echo.Context) bool {
			// File downloads are redirects to the CDN; don't gzip them.
			return strings.HasPrefix(c.Path(), "/download") || strings.HasPrefix(c.Path(), "/api/v1/files")
		},
	}))
	e.Use(counter.Middleware)
	e.HTTPErrorHandler = web.ErrorHandler()

	web.MountRedirects(e)
	files.NewHandlers(fileSvc).RegisterDownload(e)
	web.MountSPA(e, conf.UI, conf.UIProxy)

	api := web.NewAPI(e, conf.Session.Secure)
	clienttelemetry.NewHandlers().Register(api)
	guards := sessions.Guards()
	auth.NewHandlers(sessions, authSvc, uploads).Register(api, guards)
	account.NewHandlers(accountSvc, uploads).Register(api, guards)
	news.NewHandlers(newsSvc).Register(api, guards)
	whatson.NewHandlers(whatsOnSvc).Register(api, guards)
	sponsor.NewHandlers(sponsorSvc).Register(api, guards)
	affiliation.NewHandlers(affiliationSvc).Register(api, guards)
	document.NewHandlers(documentSvc).Register(api, guards)
	image.NewHandlers(gallerySvc).Register(api, guards)
	programme.NewHandlers(programmeSvc).Register(api, guards)
	setting.NewHandlers(settingSvc).Register(api, guards)
	team.NewHandlers(teamSvc).Register(api, guards)
	player.NewHandlers(playerSvc).Register(api, guards)
	files.NewHandlers(fileSvc).Register(api, guards)
	site.NewHandlers(siteSvc).Register(api, guards)
	user.NewHandlers(userSvc).Register(api, guards)

	a := &App{Echo: e, address: conf.Address, visitors: counter, tokens: tokens, mtls: conf.MTLS}
	if lister, ok := objects.(filecheck.Objects); ok && s.FileCheck != nil {
		a.files = filecheck.NewJob(filecheck.NewChecker(s.FileCheck, lister).WithRetention(conf.FileRetention), s.FileCheck)
	}
	return a
}

// Start begins background work and serves until the server stops.
func (a *App) Start() error {
	a.visitors.Start(context.Background())
	if a.files != nil {
		a.files.Start()
	}
	// Like echo.Start, shut down gracefully on Ctrl+C or SIGTERM.
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	sc := echo.StartConfig{Address: a.address}
	if a.mtls != nil {
		stop := a.mtls.ReloadOnSIGHUP()
		defer stop()
		sc.TLSConfig = a.mtls.TLSConfig()
	}
	return sc.Start(ctx, a.Echo)
}

// Stop ends background work.
func (a *App) Stop() {
	a.visitors.Stop()
	if a.files != nil {
		a.files.Stop()
	}
	a.tokens.Close()
}
