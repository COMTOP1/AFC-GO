package image

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/COMTOP1/AFC-GO/server/internal/web"
)

// Handlers serves /gallery.
type Handlers struct {
	svc *Service
}

func NewHandlers(svc *Service) *Handlers {
	return &Handlers{svc: svc}
}

// Register mounts the gallery routes on the /api/v1 group. Photographers may
// write; Managers may not.
func (h *Handlers) Register(g *echo.Group, guards web.Guards) {
	g.GET("/gallery", h.list)
	g.POST("/gallery", h.create, guards.NotManager)
	g.DELETE("/gallery/:id", h.remove, guards.NotManager)
}

// list returns every gallery photo.
//
//	@Summary	List gallery photos
//	@Tags		gallery
//	@Produce	json
//	@Success	200	{array}	image.Public
//	@Router		/gallery [get]
func (h *Handlers) list(c echo.Context) error {
	out, err := h.svc.List(c.Request().Context())
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, out)
}

// create uploads a photo.
//
//	@Summary	Upload a gallery photo
//	@Tags		gallery
//	@Accept		multipart/form-data
//	@Produce	json
//	@Param		caption	formData	string	false	"Caption"
//	@Param		image	formData	file	true	"Photo"
//	@Success	201		{object}	image.Public
//	@Failure	422		{object}	web.ErrorResponse
//	@Router		/gallery [post]
func (h *Handlers) create(c echo.Context) error {
	if err := web.RequireForm(c); err != nil {
		return err
	}
	photo, err := web.FormFile(c, "image")
	if err != nil {
		return err
	}
	out, err := h.svc.Create(c.Request().Context(), CreateInput{Caption: c.FormValue("caption")}, photo)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, out)
}

// remove deletes a photo.
//
//	@Summary	Delete a gallery photo
//	@Tags		gallery
//	@Param		id	path	int	true	"Photo ID"
//	@Success	204
//	@Failure	404	{object}	web.ErrorResponse
//	@Router		/gallery/{id} [delete]
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
