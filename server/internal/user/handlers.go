package user

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/COMTOP1/AFC-GO/server/internal/web"
)

// Handlers serves /users. Every route needs Club Secretary or higher.
type Handlers struct {
	svc *Service
}

func NewHandlers(svc *Service) *Handlers {
	return &Handlers{svc: svc}
}

// Register mounts the user-management routes on the /api/v1 group.
func (h *Handlers) Register(g *echo.Group, guards web.Guards) {
	admin := guards.ClubSecretaryHigher
	g.GET("/users", h.list, admin)
	g.POST("/users", h.create, admin)
	g.GET("/users/:id", h.get, admin)
	g.PATCH("/users/:id", h.update, admin)
	g.DELETE("/users/:id", h.remove, admin)
	g.POST("/users/:id/reset", h.reset, admin)
}

// list returns every user.
//
//	@Summary	List users
//	@Tags		users
//	@Produce	json
//	@Success	200	{array}		user.Admin
//	@Failure	403	{object}	web.ErrorResponse
//	@Router		/users [get]
func (h *Handlers) list(c echo.Context) error {
	out, err := h.svc.List(c.Request().Context())
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, out)
}

// get returns one user.
//
//	@Summary	Get a user
//	@Tags		users
//	@Produce	json
//	@Param		id	path		int	true	"User ID"
//	@Success	200	{object}	user.Admin
//	@Failure	404	{object}	web.ErrorResponse
//	@Router		/users/{id} [get]
func (h *Handlers) get(c echo.Context) error {
	id, err := web.ParamID(c, "id")
	if err != nil {
		return err
	}
	out, err := h.svc.Get(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, out)
}

// create adds a user and emails them a temporary password.
//
//	@Summary	Create a user
//	@Tags		users
//	@Accept		multipart/form-data
//	@Produce	json
//	@Param		name	formData	string	true	"Name"
//	@Param		email	formData	string	true	"Email"
//	@Param		phone	formData	string	false	"Phone"
//	@Param		role	formData	string	true	"Role code, e.g. manager, club_secretary"
//	@Param		teamId	formData	int		false	"Team (managers only)"
//	@Param		image	formData	file	false	"Photo"
//	@Success	201		{object}	user.Created
//	@Failure	409		{object}	web.ErrorResponse
//	@Failure	422		{object}	web.ErrorResponse
//	@Router		/users [post]
func (h *Handlers) create(c echo.Context) error {
	teamID, err := web.FormInt(c, "teamId")
	if err != nil {
		return err
	}
	image, err := web.FormFile(c, "image")
	if err != nil {
		return err
	}
	in := CreateInput{Name: c.FormValue("name"), Email: c.FormValue("email"), Phone: c.FormValue("phone"), Role: c.FormValue("role")}
	if teamID != nil {
		in.TeamID = *teamID
	}
	out, err := h.svc.Create(c.Request().Context(), in, image)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, out)
}

// update changes a user; omitted fields are left as they are.
//
//	@Summary	Update a user
//	@Tags		users
//	@Accept		multipart/form-data
//	@Produce	json
//	@Param		id			path		int		true	"User ID"
//	@Param		name		formData	string	false	"Name"
//	@Param		email		formData	string	false	"Email"
//	@Param		phone		formData	string	false	"Phone; empty clears"
//	@Param		role		formData	string	false	"Role code"
//	@Param		teamId		formData	int		false	"Team (managers only)"
//	@Param		image		formData	file	false	"Replacement photo"
//	@Param		removeImage	formData	bool	false	"Remove the current photo"
//	@Success	200			{object}	user.Admin
//	@Failure	404			{object}	web.ErrorResponse
//	@Failure	409			{object}	web.ErrorResponse
//	@Failure	422			{object}	web.ErrorResponse
//	@Router		/users/{id} [patch]
func (h *Handlers) update(c echo.Context) error {
	id, err := web.ParamID(c, "id")
	if err != nil {
		return err
	}
	in := UpdateInput{
		Name:  web.FormString(c, "name"),
		Email: web.FormString(c, "email"),
		Phone: web.FormString(c, "phone"),
		Role:  web.FormString(c, "role"),
	}
	if in.TeamID, err = web.FormInt(c, "teamId"); err != nil {
		return err
	}
	remove, err := web.FormBool(c, "removeImage")
	if err != nil {
		return err
	}
	in.RemoveImage = remove != nil && *remove
	image, err := web.FormFile(c, "image")
	if err != nil {
		return err
	}
	out, err := h.svc.Update(c.Request().Context(), id, in, image)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, out)
}

// remove deletes a user.
//
//	@Summary	Delete a user
//	@Tags		users
//	@Param		id	path	int	true	"User ID"
//	@Success	204
//	@Failure	404	{object}	web.ErrorResponse
//	@Router		/users/{id} [delete]
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

// reset forces a password reset and emails the user a link.
//
//	@Summary	Reset a user's password
//	@Tags		users
//	@Produce	json
//	@Param		id	path		int	true	"User ID"
//	@Success	200	{object}	user.ResetResult
//	@Failure	404	{object}	web.ErrorResponse
//	@Router		/users/{id}/reset [post]
func (h *Handlers) reset(c echo.Context) error {
	id, err := web.ParamID(c, "id")
	if err != nil {
		return err
	}
	out, err := h.svc.ResetPassword(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, out)
}
