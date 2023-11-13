package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"log"
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

	base.GET("api/health", func(c echo.Context) error {
		marshal, err := json.Marshal(struct {
			Status int `json:"status"`
		}{
			Status: http.StatusOK,
		})
		if err != nil {
			log.Println(err)
			return &echo.HTTPError{
				Code:     http.StatusBadRequest,
				Message:  err.Error(),
				Internal: err,
			}
		}

		c.Response().Header().Set("Content-Type", "application/json")
		return c.JSON(http.StatusOK, marshal)
	})

	// base is the functions that don't require being logged in
	base.GET("", r.views.HomeFunc)

	affiliation := base.Group("affiliation", r.views.RequiresLogin, r.views.RequireNotManager)
	affiliation.Match(validMethods, "/add", r.views.AffiliationAddFunc)
	affiliation.Match(validMethods, "/:id/delete", r.views.AffiliationDeleteFunc)

	base.Match(validMethods, "account", r.views.AccountFunc, r.views.RequiresLogin)
	base.Match(validMethods, "contact", r.views.ContactFunc)
	base.Match(validMethods, "documents", r.views.DocumentsFunc)

	document := base.Group("document", r.views.RequiresLogin, r.views.RequireNotManager)
	document.Match(validMethods, "/add", r.views.DocumentAddFunc)
	document.Match(validMethods, "/:id/delete", r.views.DocumentDeleteFunc)

	base.Match(validMethods, "download", r.views.DownloadFunc)
	base.Match(validMethods, "gallery", r.views.GalleryFunc)

	image := base.Group("image", r.views.RequiresLogin, r.views.RequireNotManager)
	image.Match(validMethods, "/add", r.views.ImageAddFunc)
	image.Match(validMethods, "/:id/delete", r.views.ImageDeleteFunc)

	base.Match(validMethods, "info", r.views.InfoFunc)

	news := base.Group("news")
	news.Match(validMethods, "/add", r.views.NewsAddFunc, r.views.RequiresLogin, r.views.RequireNotManager)
	newsID := news.Group("/:id")
	newsID.Match(validMethods, "/edit", r.views.NewsEditFunc, r.views.RequiresLogin, r.views.RequireNotManager)
	newsID.Match(validMethods, "/delete", r.views.NewsDeleteFunc, r.views.RequiresLogin, r.views.RequireNotManager)
	newsID.Match(validMethods, "", r.views.NewsArticleFunc)
	news.Match(validMethods, "", r.views.NewsFunc)

	programmes := base.Group("programmes")
	programmes.Match(validMethods, "/:id", r.views.ProgrammesSeasonsFunc)
	programmes.Match(validMethods, "", r.views.ProgrammesFunc)
	programme := base.Group("programme")
	programme.Match(validMethods, "/add", r.views.ProgrammeAddFunc, r.views.RequiresLogin, r.views.RequireNotManager)
	programmeID := programme.Group("/:id")
	programmeID.Match(validMethods, "/delete", r.views.ProgrammeDeleteFunc, r.views.RequiresLogin, r.views.RequireNotManager)

	season := base.Group("season", r.views.RequiresLogin, r.views.RequireNotManager)
	season.Match(validMethods, "/add", r.views.ProgrammeSeasonAddFunc)
	seasonsID := season.Group("/:id")
	seasonsID.Match(validMethods, "/edit", r.views.ProgrammeSeasonEditFunc)
	seasonsID.Match(validMethods, "/delete", r.views.ProgrammeSeasonDeleteFunc)

	base.Match(validMethods, "sponsors", r.views.SponsorsFunc)

	sponsor := base.Group("sponsor", r.views.RequiresLogin, r.views.RequireNotManager)
	sponsor.Match(validMethods, "/add", r.views.SponsorAddFunc)
	sponsor.Match(validMethods, "/:id/delete", r.views.SponsorDeleteFunc)

	base.Match(validMethods, "teams", r.views.TeamsFunc)

	team := base.Group("team")
	team.Match(validMethods, "/add", r.views.TeamAddFunc, r.views.RequiresLogin, r.views.RequireNotManager)
	teamID := team.Group("/:id")
	teamID.Match(validMethods, "/edit", r.views.TeamEditFunc, r.views.RequiresLogin, r.views.RequireNotManager)
	teamID.Match(validMethods, "/delete", r.views.TeamDeleteFunc, r.views.RequiresLogin, r.views.RequireNotManager)
	teamID.Match(validMethods, "", r.views.TeamFunc)

	whatson := base.Group("whatson")
	whatson.Match(validMethods, "/add", r.views.WhatsOnAddFunc, r.views.RequiresLogin, r.views.RequireNotManager)
	whatsonID := whatson.Group("/:id")
	whatsonID.Match(validMethods, "/edit", r.views.WhatsOnEditFunc, r.views.RequiresLogin, r.views.RequireNotManager)
	whatsonID.Match(validMethods, "/delete", r.views.WhatsOnDeleteFunc, r.views.RequiresLogin, r.views.RequireNotManager)
	whatsonID.Match(validMethods, "", r.views.WhatsOnArticleFunc)
	whatson.Match(validMethods, "", r.views.WhatsOnFunc)

	base.Match(validMethods, "login", r.views.LoginFunc)

	base.Match(validMethods, "players", r.views.PlayersFunc, r.views.RequiresLogin)

	player := base.Group("player", r.views.RequiresLogin, r.views.RequireNotManager)
	player.Match(validMethods, "/add", r.views.PlayerAddFunc)
	playerID := player.Group("/:id")
	playerID.Match(validMethods, "/edit", r.views.PlayerEditFunc)
	playerID.Match(validMethods, "/delete", r.views.PlayerDeleteFunc)

	base.Match(validMethods, "users", r.views.UsersFunc, r.views.RequiresLogin, r.views.RequireClubSecretaryHigher)

	user := base.Group("user", r.views.RequiresLogin, r.views.RequireClubSecretaryHigher)
	user.Match(validMethods, "/add", r.views.UserAddFunc)
	userID := user.Group("/:id")
	userID.Match(validMethods, "/edit", r.views.UserEditFunc)
	userID.Match(validMethods, "/delete", r.views.UserDeleteFunc)

	base.Match(validMethods, "logout", r.views.LogoutFunc, r.views.RequiresLogin)
}
