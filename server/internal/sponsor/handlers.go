package sponsor

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/COMTOP1/AFC-GO/server/internal/web"
)

// Handlers serves /sponsors.
type Handlers struct {
	svc *Service
}

func NewHandlers(svc *Service) *Handlers {
	return &Handlers{svc: svc}
}

// Register mounts the sponsor routes on the /api/v1 group.
func (h *Handlers) Register(g *echo.Group, guards web.Guards) {
	g.GET("/sponsors", h.list)
	g.POST("/sponsors", h.create, guards.Editor)
	g.DELETE("/sponsors/:id", h.remove, guards.Editor)
}

// list returns every sponsor.
//
//	@Summary	List sponsors
//	@Tags		sponsors
//	@Produce	json
//	@Success	200	{array}	sponsor.Public
//	@Router		/sponsors [get]
func (h *Handlers) list(c echo.Context) error {
	out, err := h.svc.List(c.Request().Context())
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, out)
}

// create adds a sponsor.
//
//	@Summary	Create a sponsor
//	@Tags		sponsors
//	@Accept		multipart/form-data
//	@Produce	json
//	@Param		name	formData	string	true	"Name"
//	@Param		website	formData	string	false	"Full URL"
//	@Param		purpose	formData	string	false	"What they sponsor"
//	@Param		team	formData	string	false	"A, O, Y or a team ID"
//	@Param		image	formData	file	true	"Logo"
//	@Success	201		{object}	sponsor.Public
//	@Failure	422		{object}	web.ErrorResponse
//	@Router		/sponsors [post]
func (h *Handlers) create(c echo.Context) error {
	image, err := web.FormFile(c, "image")
	if err != nil {
		return err
	}
	out, err := h.svc.Create(c.Request().Context(), CreateInput{
		Name:    c.FormValue("name"),
		Website: c.FormValue("website"),
		Purpose: c.FormValue("purpose"),
		Team:    c.FormValue("team"),
	}, image)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, out)
}

// remove deletes a sponsor.
//
//	@Summary	Delete a sponsor
//	@Tags		sponsors
//	@Param		id	path	int	true	"Sponsor ID"
//	@Success	204
//	@Failure	404	{object}	web.ErrorResponse
//	@Router		/sponsors/{id} [delete]
func (h *Handlers) remove(c echo.Context) error {
	id, err := web.ParamID(c, "id")
	if err != nil {
		return err
	}
	if _, err = h.svc.Delete(c.Request().Context(), id); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}
