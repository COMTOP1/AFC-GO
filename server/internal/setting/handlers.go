package setting

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/COMTOP1/AFC-GO/server/internal/web"
)

// Handlers serves /info and /settings.
type Handlers struct {
	svc *Service
}

func NewHandlers(svc *Service) *Handlers {
	return &Handlers{svc: svc}
}

// Register mounts the settings routes on the /api/v1 group.
func (h *Handlers) Register(g *echo.Group, guards web.Guards) {
	g.GET("/info", h.info)
	g.PUT("/info", h.setInfo, guards.Editor)
	g.PUT("/settings/display-email", h.setDisplayEmail, guards.ClubSecretaryHigher)
}

// info returns the club information page HTML.
//
//	@Summary	Get the info page
//	@Tags		settings
//	@Produce	json
//	@Success	200	{object}	setting.InfoContent
//	@Router		/info [get]
func (h *Handlers) info(c echo.Context) error {
	content, err := h.svc.Info(c.Request().Context())
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, InfoContent{Content: content})
}

// setInfo replaces the info page HTML (sanitised).
//
//	@Summary	Update the info page
//	@Tags		settings
//	@Accept		json
//	@Produce	json
//	@Param		request	body		setting.InfoContent	true	"HTML content"
//	@Success	200		{object}	setting.InfoContent
//	@Router		/info [put]
func (h *Handlers) setInfo(c echo.Context) error {
	var in InfoContent
	if err := web.BindJSON(c, &in); err != nil {
		return err
	}
	saved, err := h.svc.SetInfo(c.Request().Context(), in.Content)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, InfoContent{Content: saved})
}

// setDisplayEmail sets (or, with an empty email, clears) the public contact email.
//
//	@Summary	Set the public contact email
//	@Tags		settings
//	@Accept		json
//	@Produce	json
//	@Param		request	body		setting.DisplayEmail	true	"Email; empty removes it"
//	@Success	200		{object}	setting.DisplayEmail
//	@Failure	422		{object}	web.ErrorResponse
//	@Router		/settings/display-email [put]
func (h *Handlers) setDisplayEmail(c echo.Context) error {
	var in DisplayEmail
	if err := web.BindJSON(c, &in); err != nil {
		return err
	}
	saved, err := h.svc.SetDisplayEmail(c.Request().Context(), in.Email)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, DisplayEmail{Email: saved})
}
