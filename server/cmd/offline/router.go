package main

import (
	"embed"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

//go:embed public/*
var embeddedFiles embed.FS

type (
	Router struct {
		config *Config
		views  *Views
		router *echo.Echo
	}
	RouterConf struct {
		Config *Config
		Views  *Views
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

	assetHandler := http.FileServer(http.FS(echo.MustSubFS(embeddedFiles, "public")))

	r.router.GET("/public/*", echo.WrapHandler(http.StripPrefix("/public/", assetHandler)))

	r.router.GET("*", r.views.Error404)
}

func (v *Views) Error404(c echo.Context) error {
	return v.template.RenderTemplate(c.Response().Writer, struct {
		Year int
	}{
		Year: time.Now().Year(),
	}, OfflineTemplate, RegularType)
}
