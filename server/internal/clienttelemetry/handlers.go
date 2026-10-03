package clienttelemetry

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/COMTOP1/AFC-GO/server/internal/svcerr"
	"github.com/COMTOP1/AFC-GO/server/internal/web"
)

// Handlers serves the telemetry ingest endpoint.
type Handlers struct{}

func NewHandlers() *Handlers { return &Handlers{} }

// Register mounts POST /telemetry on the /api/v1 group. Deliberately
// unguarded: a signed-out visitor hitting a login error, or any page that
// crashes before a session loads, must still be reportable.
func (h *Handlers) Register(g *echo.Group) {
	g.POST("/telemetry", h.ingest, middleware.BodyLimit("16K"))
}

// ingest records one client-reported telemetry event as a span.
//
//	@Summary	Report a client telemetry event
//	@Tags		system
//	@Accept		json
//	@Param		request	body	clienttelemetry.Event	true	"Event"
//	@Success	204
//	@Router		/telemetry [post]
func (h *Handlers) ingest(c echo.Context) error {
	var ev Event
	if err := web.BindJSON(c, &ev); err != nil {
		return err
	}
	switch {
	case ev.Name == "":
		return svcerr.InvalidField("name", "name is required")
	case ev.StartTime == 0:
		return svcerr.InvalidField("startTime", "startTime is required")
	case ev.Type != "api" && ev.Type != "error":
		return svcerr.InvalidField("type", `type must be "api" or "error"`)
	}
	Record(c.Request().Context(), ev)
	return c.NoContent(http.StatusNoContent)
}
