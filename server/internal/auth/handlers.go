package auth

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/COMTOP1/AFC-GO/server/internal/upload"
	"github.com/COMTOP1/AFC-GO/server/internal/web"
)

// Handlers serves the /auth endpoints.
type Handlers struct {
	sessions *Sessions
	files    *upload.Files
}

func NewHandlers(sessions *Sessions, files *upload.Files) *Handlers {
	return &Handlers{sessions: sessions, files: files}
}

// Register mounts the /auth routes on the /api/v1 group.
func (h *Handlers) Register(g *echo.Group, guards web.Guards) {
	g.GET("/auth/me", h.me, guards.Login)
}

// me returns the logged-in user.
//
//	@Summary	Current user
//	@Tags		auth
//	@Produce	json
//	@Success	200	{object}	auth.CurrentUser
//	@Failure	401	{object}	web.ErrorResponse
//	@Router		/auth/me [get]
func (h *Handlers) me(c echo.Context) error {
	u, _ := Current(c)
	return c.JSON(http.StatusOK, NewCurrentUser(u, h.files))
}
