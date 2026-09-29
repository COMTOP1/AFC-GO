package site

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/COMTOP1/AFC-GO/server/internal/web"
)

// Handlers serves /site, /home, /contact and /teams/{id}.
type Handlers struct {
	svc *Service
}

func NewHandlers(svc *Service) *Handlers {
	return &Handlers{svc: svc}
}

// Register mounts the site routes on the /api/v1 group.
func (h *Handlers) Register(g *echo.Group, _ web.Guards) {
	g.GET("/site", h.site)
	g.GET("/home", h.home)
	g.GET("/contact", h.contact)
	g.GET("/teams/:id", h.team)
}

// site returns layout data shared by every page.
//
//	@Summary	Site layout data
//	@Tags		site
//	@Produce	json
//	@Success	200	{object}	site.Info
//	@Router		/site [get]
func (h *Handlers) site(c echo.Context) error {
	out, err := h.svc.Site(c.Request().Context())
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, out)
}

// home returns the home page.
//
//	@Summary	Home page
//	@Tags		site
//	@Produce	json
//	@Success	200	{object}	site.Home
//	@Router		/home [get]
func (h *Handlers) home(c echo.Context) error {
	return c.JSON(http.StatusOK, h.svc.Home(c.Request().Context()))
}

// contact returns the contact page.
//
//	@Summary	Contact page
//	@Tags		site
//	@Produce	json
//	@Success	200	{object}	site.Contact
//	@Router		/contact [get]
func (h *Handlers) contact(c echo.Context) error {
	out, err := h.svc.Contact(c.Request().Context())
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, out)
}

// team returns a team's public page: details, managers, sponsors and squad.
//
//	@Summary	Team page
//	@Tags		teams
//	@Produce	json
//	@Param		id	path		int	true	"Team ID"
//	@Success	200	{object}	site.TeamDetail
//	@Failure	404	{object}	web.ErrorResponse
//	@Router		/teams/{id} [get]
func (h *Handlers) team(c echo.Context) error {
	id, err := web.ParamID(c, "id")
	if err != nil {
		return err
	}
	out, err := h.svc.Team(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, out)
}
