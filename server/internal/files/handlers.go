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
	u, err := h.svc.URL(c.Request().Context(), c.Param("kind"), id)
	if err != nil {
		return err
	}
	c.Response().Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	return c.Redirect(http.StatusFound, u)
}
