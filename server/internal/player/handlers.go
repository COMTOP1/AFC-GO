package player

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/COMTOP1/AFC-GO/server/internal/web"
)

// Handlers serves /players.
type Handlers struct {
	svc *Service
}

func NewHandlers(svc *Service) *Handlers {
	return &Handlers{svc: svc}
}

// Register mounts the player routes on the /api/v1 group.
func (h *Handlers) Register(g *echo.Group, guards web.Guards) {
	g.GET("/players", h.list, guards.Login)
	g.POST("/players", h.create, guards.Editor)
	g.PATCH("/players/:id", h.update, guards.Editor)
	g.DELETE("/players/:id", h.remove, guards.Editor)
}

// list returns every player (logged-in users only).
//
//	@Summary	List players
//	@Tags		players
//	@Produce	json
//	@Success	200	{array}		player.Public
//	@Failure	401	{object}	web.ErrorResponse
//	@Router		/players [get]
func (h *Handlers) list(c echo.Context) error {
	out, err := h.svc.List(c.Request().Context())
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, out)
}

// create adds a player.
//
//	@Summary	Create a player
//	@Tags		players
//	@Accept		multipart/form-data
//	@Produce	json
//	@Param		name		formData	string	true	"Name"
//	@Param		teamId		formData	int		true	"Team ID"
//	@Param		dateOfBirth	formData	string	true	"YYYY-MM-DD"
//	@Param		position	formData	string	false	"Position"
//	@Param		isCaptain	formData	bool	false	"Captain"
//	@Param		image		formData	file	false	"Photo (never shown for youth or under-18 players)"
//	@Success	201			{object}	player.Public
//	@Failure	422			{object}	web.ErrorResponse
//	@Router		/players [post]
func (h *Handlers) create(c echo.Context) error {
	if err := web.RequireForm(c); err != nil {
		return err
	}
	teamID, err := web.FormInt(c, "teamId")
	if err != nil {
		return err
	}
	dob, err := web.FormDate(c, "dateOfBirth")
	if err != nil {
		return err
	}
	captain, err := web.FormBool(c, "isCaptain")
	if err != nil {
		return err
	}
	image, err := web.FormFile(c, "image")
	if err != nil {
		return err
	}
	in := CreateInput{Name: c.FormValue("name"), Position: c.FormValue("position"), IsCaptain: captain != nil && *captain}
	if teamID != nil {
		in.TeamID = *teamID
	}
	if dob != nil {
		in.DateOfBirth = *dob
	}
	out, err := h.svc.Create(c.Request().Context(), in, image)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, out)
}

// update changes a player; omitted fields are left as they are.
//
//	@Summary	Update a player
//	@Tags		players
//	@Accept		multipart/form-data
//	@Produce	json
//	@Param		id			path		int		true	"Player ID"
//	@Param		name		formData	string	false	"Name"
//	@Param		teamId		formData	int		false	"Team ID"
//	@Param		dateOfBirth	formData	string	false	"YYYY-MM-DD"
//	@Param		position	formData	string	false	"Position; empty clears"
//	@Param		isCaptain	formData	bool	false	"Captain"
//	@Param		image		formData	file	false	"Replacement photo"
//	@Param		removeImage	formData	bool	false	"Remove the current photo"
//	@Success	200			{object}	player.Public
//	@Failure	404			{object}	web.ErrorResponse
//	@Failure	422			{object}	web.ErrorResponse
//	@Router		/players/{id} [patch]
func (h *Handlers) update(c echo.Context) error {
	id, err := web.ParamID(c, "id")
	if err != nil {
		return err
	}
	if err = web.RequireForm(c); err != nil {
		return err
	}
	in := UpdateInput{Name: web.FormString(c, "name"), Position: web.FormString(c, "position")}
	if in.TeamID, err = web.FormInt(c, "teamId"); err != nil {
		return err
	}
	if in.DateOfBirth, err = web.FormDate(c, "dateOfBirth"); err != nil {
		return err
	}
	if in.IsCaptain, err = web.FormBool(c, "isCaptain"); err != nil {
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

// remove deletes a player.
//
//	@Summary	Delete a player
//	@Tags		players
//	@Param		id	path	int	true	"Player ID"
//	@Success	204
//	@Failure	404	{object}	web.ErrorResponse
//	@Router		/players/{id} [delete]
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
