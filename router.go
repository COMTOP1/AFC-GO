package main

import (
	"embed"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/COMTOP1/AFC-GO/views"
)

//go:embed public/*
var embeddedFiles embed.FS

type (
	Router struct {
		config *views.Config
		views  *views.Views
		router *echo.Echo
	}
	RouterConf struct {
		Config *views.Config
		Views  *views.Views
	}
)

func NewRouter(conf *RouterConf) *Router {
	r := &Router{
		config: conf.Config,
		router: echo.New(),
		views:  conf.Views,
	}
	r.router.HideBanner = true

	r.middleware()

	r.loadRoutes()

	return r
}

func (r *Router) Start() error {
	r.router.Logger.Error(r.router.Start(r.config.Address))
	return fmt.Errorf("failed to start router on address %s", r.config.Address)
}

// middleware initialises web server middleware
func (r *Router) middleware() {
	r.router.Pre(middleware.RemoveTrailingSlash())
	r.router.Use(middleware.Recover())
	r.router.Use(middleware.BodyLimit("15M"))
	r.router.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Level: 5,
	}))
}

func (r *Router) loadRoutes() {
	r.router.RouteNotFound("/*", r.views.Error404)

	r.router.HTTPErrorHandler = r.views.CustomHTTPErrorHandler

	assetHandler := http.FileServer(http.FS(echo.MustSubFS(embeddedFiles, "public")))

	r.router.GET("/public/*", echo.WrapHandler(http.StripPrefix("/public/", assetHandler)))

	validMethods := []string{http.MethodGet, http.MethodPost}

	base := r.router.Group("/")

	affiliation := base.Group("affiliation", r.views.RequiresLogin)
	affiliation.Match(validMethods, "add", r.views.AffiliationAddFunc)
	affiliation.Match(validMethods, "delete/:id", r.views.AffiliationDeleteFunc)

	// base is the functions that don't require being logged in
	base.GET("", r.views.HomeFunc)
	base.Match(validMethods, "login", r.views.LoginFunc)
	base.Match(validMethods, "logout", r.views.LogoutFunc, r.views.RequiresLogin)
	base.Match(validMethods, "download", r.views.Download)
	base.Match(validMethods, "account", r.views.AccountFunc, r.views.RequiresLogin)
	base.Match(validMethods, "contact", r.views.ContactFunc)
	base.Match(validMethods, "info", r.views.InfoFunc)
	//base.Match(validMethods, "signup", r.views.SignUpFunc)
	//base.Match(validMethods, "forgot", r.views.ForgotFunc)
	//base.Match(validMethods, "reset/:url", r.views.ResetURLFunc)
}
