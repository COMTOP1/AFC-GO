package files

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/COMTOP1/AFC-GO/server/internal/web"
)

// Handlers serves /files.
type Handlers struct {
	svc *Service
}

func NewHandlers(svc *Service) *Handlers {
	return &Handlers{svc: svc}
}

// Register mounts the file route on the /api/v1 group.
func (h *Handlers) Register(g *echo.Group, _ web.Guards) {
	g.GET("/files/:kind/:id", h.get)
}

// get redirects to a stored file.
//
//	@Summary	Download a file
//	@Tags		files
//	@Param		kind	path	string	true	"affiliation, document, gallery, news, player, programme, sponsor, team, user or whatson"
//	@Param		id		path	int		true	"Resource ID"
//	@Success	302
//	@Failure	404	{object}	web.ErrorResponse
//	@Router		/files/{kind}/{id} [get]
func (h *Handlers) get(c echo.Context) error {
	id, err := web.ParamID(c, "id")
	if err != nil {
		return err
	}
	kind := c.Param("kind")
	u, err := h.svc.URL(c.Request().Context(), kind, id)
	if err != nil {
		return err
	}
	c.Response().Header().Set("Cache-Control", CacheControl(kind))
	return c.Redirect(http.StatusFound, u)
}

// CacheControl returns the Cache-Control for a file redirect of the given
// kind (shared by the API handler and the legacy /download route). Player
// photos must not be cached publicly: if a player moves to a youth team, or
// an under-18's photo must stop being shown, a year-long public cache would
// keep serving the old redirect to browsers and CDNs regardless.
func CacheControl(kind string) string {
	if kind == "player" {
		return "private, no-cache"
	}
	return "public, max-age=31536000, immutable"
}
