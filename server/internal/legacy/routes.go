package legacy

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/COMTOP1/AFC-GO/server/internal/legacy/views"
)

// Mount registers the server-rendered template routes and /public assets.
func Mount(e *echo.Echo, v *views.Views) {
	e.RouteNotFound("/*", v.Error404)

	assetHandler := http.FileServer(http.FS(echo.MustSubFS(Public, "public")))

	e.GET("/public/*", echo.WrapHandler(http.StripPrefix("/public/", assetHandler)))

	validMethods := []string{http.MethodGet, http.MethodPost}

	base := e.Group("/")

	// base is the functions that don't require being logged in
	base.GET("", v.HomeFunc)

	affiliation := base.Group("affiliation", v.RequiresLogin, v.RequireNotManagerNotPhotographer)
	affiliation.Match(validMethods, "/add", v.AffiliationAddFunc)
	affiliation.Match(validMethods, "/:id/delete", v.AffiliationDeleteFunc)

	account := base.Group("account", v.RequiresLogin)
	account.Match(validMethods, "/uploadimage", v.UploadImageFunc)
	account.Match(validMethods, "/removeimage", v.RemoveImageFunc)
	account.Match(validMethods, "", v.AccountFunc)

	base.Match(validMethods, "contact", v.ContactFunc)
	base.Match(validMethods, "documents", v.DocumentsFunc)

	document := base.Group("document", v.RequiresLogin, v.RequireNotManagerNotPhotographer)
	document.Match(validMethods, "/add", v.DocumentAddFunc)
	document.Match(validMethods, "/:id/delete", v.DocumentDeleteFunc)

	base.Match(validMethods, "download", v.DownloadFunc)
	base.Match(validMethods, "gallery", v.GalleryFunc)

	image := base.Group("image", v.RequiresLogin, v.RequireNotManager)
	image.Match(validMethods, "/add", v.ImageAddFunc)
	image.Match(validMethods, "/:id/delete", v.ImageDeleteFunc)

	base.Match(validMethods, "info", v.InfoFunc)
	base.Match(validMethods, "info/edit", v.InfoEditFunc, v.RequiresLogin, v.RequireNotManagerNotPhotographer)

	news := base.Group("news")
	news.Match(validMethods, "/add", v.NewsAddFunc, v.RequiresLogin, v.RequireNotManagerNotPhotographer)
	newsID := news.Group("/:id")
	newsID.Match(validMethods, "/edit", v.NewsEditFunc, v.RequiresLogin, v.RequireNotManagerNotPhotographer)
	newsID.Match(validMethods, "/delete", v.NewsDeleteFunc, v.RequiresLogin, v.RequireNotManagerNotPhotographer)
	newsID.Match(validMethods, "", v.NewsArticleFunc)
	news.Match(validMethods, "", v.NewsFunc)

	programmes := base.Group("programmes")
	programmes.Match(validMethods, "/:id", v.ProgrammesSeasonsFunc)
	programmes.Match(validMethods, "", v.ProgrammesFunc)
	base.Match(validMethods, "programmeselect", v.ProgrammeSeasonSelectFunc)
	programme := base.Group("programme")
	programme.Match(validMethods, "/add", v.ProgrammeAddFunc, v.RequiresLogin, v.RequireNotManagerNotPhotographer)
	programmeID := programme.Group("/:id")
	programmeID.Match(validMethods, "/delete", v.ProgrammeDeleteFunc, v.RequiresLogin, v.RequireNotManagerNotPhotographer)

	season := base.Group("season", v.RequiresLogin, v.RequireNotManagerNotPhotographer)
	season.Match(validMethods, "/add", v.ProgrammeSeasonAddFunc)
	seasonsID := season.Group("/:id")
	seasonsID.Match(validMethods, "/edit", v.ProgrammeSeasonEditFunc)
	seasonsID.Match(validMethods, "/delete", v.ProgrammeSeasonDeleteFunc)

	base.Match(validMethods, "sponsors", v.SponsorsFunc)

	sponsor := base.Group("sponsor", v.RequiresLogin, v.RequireNotManagerNotPhotographer)
	sponsor.Match(validMethods, "/add", v.SponsorAddFunc)
	sponsor.Match(validMethods, "/:id/delete", v.SponsorDeleteFunc)

	base.Match(validMethods, "teams", v.TeamsFunc)

	team := base.Group("team")
	team.Match(validMethods, "/add", v.TeamAddFunc, v.RequiresLogin, v.RequireNotManagerNotPhotographer)
	teamID := team.Group("/:id")
	teamID.Match(validMethods, "/edit", v.TeamEditFunc, v.RequiresLogin, v.RequireNotManagerNotPhotographer)
	teamID.Match(validMethods, "/delete", v.TeamDeleteFunc, v.RequiresLogin, v.RequireNotManagerNotPhotographer)
	teamID.Match(validMethods, "", v.TeamFunc)

	whatson := base.Group("whatson")
	whatson.Match(validMethods, "/add", v.WhatsOnAddFunc, v.RequiresLogin, v.RequireNotManagerNotPhotographer)
	whatsonID := whatson.Group("/:id")
	whatsonID.Match(validMethods, "/edit", v.WhatsOnEditFunc, v.RequiresLogin, v.RequireNotManagerNotPhotographer)
	whatsonID.Match(validMethods, "/delete", v.WhatsOnDeleteFunc, v.RequiresLogin, v.RequireNotManagerNotPhotographer)
	whatsonID.Match(validMethods, "", v.WhatsOnArticleFunc)
	whatson.Match(validMethods, "period/:timePeriod", v.WhatsOnTomePeriodFunc)
	whatson.Match(validMethods, "", v.WhatsOnFunc)
	base.Match(validMethods, "whatsonselect", v.WhatsOnSelectFunc)

	base.Match(validMethods, "login", v.LoginFunc)

	base.Match(validMethods, "players", v.PlayersFunc, v.RequiresLogin)

	player := base.Group("player", v.RequiresLogin, v.RequireNotManagerNotPhotographer)
	player.Match(validMethods, "/add", v.PlayerAddFunc)
	playerID := player.Group("/:id")
	playerID.Match(validMethods, "/edit", v.PlayerEditFunc)
	playerID.Match(validMethods, "/delete", v.PlayerDeleteFunc)

	users := base.Group("users", v.RequiresLogin, v.RequireClubSecretaryHigher)
	users.Match(validMethods, "/setdisplayemail", v.UsersSetDisplayEmailFunc)
	users.Match(validMethods, "", v.UsersFunc)

	user := base.Group("user", v.RequiresLogin, v.RequireClubSecretaryHigher)
	user.Match(validMethods, "/add", v.UserAddFunc)
	userID := user.Group("/:id")
	userID.Match(validMethods, "/edit", v.UserEditFunc)
	userID.Match(validMethods, "/delete", v.UserDeleteFunc)
	userID.Match(validMethods, "/reset", v.ResetUserPasswordFunc)

	base.Match(validMethods, "reset/:url", v.ResetURLFunc)

	base.Match(validMethods, "changepassword", v.ChangePasswordFunc, v.RequiresLogin)

	base.Match(validMethods, "logout", v.LogoutFunc, v.RequiresLogin)
}
