package affiliation

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/COMTOP1/AFC-GO/server/internal/web"
)

// Handlers serves /affiliations.
type Handlers struct {
	svc *Service
}

func NewHandlers(svc *Service) *Handlers {
	return &Handlers{svc: svc}
}

// Register mounts the affiliation routes on the /api/v1 group.
func (h *Handlers) Register(g *echo.Group, guards web.Guards) {
	g.GET("/affiliations", h.list)
	g.POST("/affiliations", h.create, guards.Editor)
	g.DELETE("/affiliations/:id", h.remove, guards.Editor)
}

// list returns every affiliation.
//
//	@Summary	List affiliations
//	@Tags		affiliations
//	@Produce	json
//	@Success	200	{array}	affiliation.Public
//	@Router		/affiliations [get]
func (h *Handlers) list(c echo.Context) error {
	out, err := h.svc.List(c.Request().Context())
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, out)
}

// create adds an affiliation.
//
//	@Summary	Create an affiliation
//	@Tags		affiliations
//	@Accept		multipart/form-data
//	@Produce	json
//	@Param		name	formData	string	true	"Name"
//	@Param		website	formData	string	false	"Full URL"
//	@Param		image	formData	file	true	"Logo"
//	@Success	201		{object}	affiliation.Public
//	@Failure	422		{object}	web.ErrorResponse
//	@Router		/affiliations [post]
func (h *Handlers) create(c echo.Context) error {
	image, err := web.FormFile(c, "image")
	if err != nil {
		return err
	}
	out, err := h.svc.Create(c.Request().Context(), CreateInput{
		Name:    c.FormValue("name"),
		Website: c.FormValue("website"),
	}, image)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, out)
}

// remove deletes an affiliation.
//
//	@Summary	Delete an affiliation
//	@Tags		affiliations
//	@Param		id	path	int	true	"Affiliation ID"
//	@Success	204
//	@Failure	404	{object}	web.ErrorResponse
//	@Router		/affiliations/{id} [delete]
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
