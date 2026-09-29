package account

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/COMTOP1/AFC-GO/server/internal/auth"
	"github.com/COMTOP1/AFC-GO/server/internal/upload"
	"github.com/COMTOP1/AFC-GO/server/internal/web"
)

// Handlers serves /account.
type Handlers struct {
	svc   *Service
	files *upload.Files
}

func NewHandlers(svc *Service, files *upload.Files) *Handlers {
	return &Handlers{svc: svc, files: files}
}

// Register mounts the account routes on the /api/v1 group.
func (h *Handlers) Register(g *echo.Group, guards web.Guards) {
	g.GET("/account", h.get, guards.Login)
	g.PUT("/account/image", h.setImage, guards.Login)
	g.DELETE("/account/image", h.removeImage, guards.Login)
}

// get returns the logged-in user's account.
//
//	@Summary	My account
//	@Tags		account
//	@Produce	json
//	@Success	200	{object}	auth.CurrentUser
//	@Failure	401	{object}	web.ErrorResponse
//	@Router		/account [get]
func (h *Handlers) get(c echo.Context) error {
	u, _ := auth.Current(c)
	return c.JSON(http.StatusOK, auth.NewCurrentUser(u, h.files))
}

// setImage replaces my photo.
//
//	@Summary	Set my photo
//	@Tags		account
//	@Accept		multipart/form-data
//	@Produce	json
//	@Param		image	formData	file	true	"Photo"
//	@Success	200		{object}	auth.CurrentUser
//	@Failure	422		{object}	web.ErrorResponse
//	@Router		/account/image [put]
func (h *Handlers) setImage(c echo.Context) error {
	if err := web.RequireForm(c); err != nil {
		return err
	}
	image, err := web.FormFile(c, "image")
	if err != nil {
		return err
	}
	u, _ := auth.Current(c)
	updated, err := h.svc.SetImage(c.Request().Context(), u, image)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, auth.NewCurrentUser(updated, h.files))
}

// removeImage clears my photo.
//
//	@Summary	Remove my photo
//	@Tags		account
//	@Produce	json
//	@Success	200	{object}	auth.CurrentUser
//	@Router		/account/image [delete]
func (h *Handlers) removeImage(c echo.Context) error {
	u, _ := auth.Current(c)
	updated, err := h.svc.RemoveImage(c.Request().Context(), u)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, auth.NewCurrentUser(updated, h.files))
}
