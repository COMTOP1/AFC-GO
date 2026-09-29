package team

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/COMTOP1/AFC-GO/server/internal/svcerr"
	"github.com/COMTOP1/AFC-GO/server/internal/web"
)

// Handlers serves /teams (the detail view GET /teams/{id} is served by the
// site package, because it aggregates players, managers and sponsors).
type Handlers struct {
	svc *Service
}

func NewHandlers(svc *Service) *Handlers {
	return &Handlers{svc: svc}
}

// Register mounts the team routes on the /api/v1 group.
func (h *Handlers) Register(g *echo.Group, guards web.Guards) {
	g.GET("/teams", h.list, guards.Identify)
	g.POST("/teams", h.create, guards.Editor)
	g.PATCH("/teams/:id", h.update, guards.Editor)
	g.DELETE("/teams/:id", h.remove, guards.Editor)
}

// list returns active teams, or all teams for logged-in users.
//
//	@Summary	List teams
//	@Tags		teams
//	@Produce	json
//	@Success	200	{array}	team.Public
//	@Router		/teams [get]
func (h *Handlers) list(c echo.Context) error {
	out, err := h.svc.List(c.Request().Context(), web.LoggedIn(c))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, out)
}

// create adds a team.
//
//	@Summary	Create a team
//	@Tags		teams
//	@Accept		multipart/form-data
//	@Produce	json
//	@Param		name		formData	string	true	"Name"
//	@Param		ages		formData	int		true	"Upper age (under 19 makes a youth team)"
//	@Param		description	formData	string	false	"Description"
//	@Param		league		formData	string	false	"League"
//	@Param		division	formData	string	false	"Division"
//	@Param		leagueTable	formData	string	false	"League table URL"
//	@Param		fixtures	formData	string	false	"Fixtures URL"
//	@Param		coach		formData	string	false	"Coach"
//	@Param		physio		formData	string	false	"Physio"
//	@Param		isActive	formData	bool	false	"Active"
//	@Param		isYouth		formData	bool	false	"Youth"
//	@Param		image		formData	file	false	"Team photo"
//	@Success	201			{object}	team.Public
//	@Failure	422			{object}	web.ErrorResponse
//	@Router		/teams [post]
func (h *Handlers) create(c echo.Context) error {
	if err := web.RequireForm(c); err != nil {
		return err
	}
	ages, err := web.FormInt(c, "ages")
	if err != nil {
		return err
	}
	if ages == nil {
		return svcerr.InvalidField("ages", "ages is required")
	}
	active, err := web.FormBool(c, "isActive")
	if err != nil {
		return err
	}
	youth, err := web.FormBool(c, "isYouth")
	if err != nil {
		return err
	}
	image, err := web.FormFile(c, "image")
	if err != nil {
		return err
	}
	out, err := h.svc.Create(c.Request().Context(), CreateInput{
		Name:        c.FormValue("name"),
		Description: c.FormValue("description"),
		League:      c.FormValue("league"),
		Division:    c.FormValue("division"),
		LeagueTable: c.FormValue("leagueTable"),
		Fixtures:    c.FormValue("fixtures"),
		Coach:       c.FormValue("coach"),
		Physio:      c.FormValue("physio"),
		IsActive:    active != nil && *active,
		IsYouth:     youth != nil && *youth,
		Ages:        *ages,
	}, image)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, out)
}

// update changes a team; omitted fields are left as they are.
//
//	@Summary	Update a team
//	@Tags		teams
//	@Accept		multipart/form-data
//	@Produce	json
//	@Param		id			path		int		true	"Team ID"
//	@Param		name		formData	string	false	"Name"
//	@Param		ages		formData	int		false	"Upper age"
//	@Param		description	formData	string	false	"Description; empty clears"
//	@Param		league		formData	string	false	"League; empty clears"
//	@Param		division	formData	string	false	"Division; empty clears"
//	@Param		leagueTable	formData	string	false	"League table URL; empty clears"
//	@Param		fixtures	formData	string	false	"Fixtures URL; empty clears"
//	@Param		coach		formData	string	false	"Coach; empty clears"
//	@Param		physio		formData	string	false	"Physio; empty clears"
//	@Param		isActive	formData	bool	false	"Active"
//	@Param		isYouth		formData	bool	false	"Youth"
//	@Param		image		formData	file	false	"Replacement photo"
//	@Param		removeImage	formData	bool	false	"Remove the current photo"
//	@Success	200			{object}	team.Public
//	@Failure	404			{object}	web.ErrorResponse
//	@Failure	422			{object}	web.ErrorResponse
//	@Router		/teams/{id} [patch]
func (h *Handlers) update(c echo.Context) error {
	id, err := web.ParamID(c, "id")
	if err != nil {
		return err
	}
	if err = web.RequireForm(c); err != nil {
		return err
	}
	in := UpdateInput{
		Name:        web.FormString(c, "name"),
		Description: web.FormString(c, "description"),
		League:      web.FormString(c, "league"),
		Division:    web.FormString(c, "division"),
		LeagueTable: web.FormString(c, "leagueTable"),
		Fixtures:    web.FormString(c, "fixtures"),
		Coach:       web.FormString(c, "coach"),
		Physio:      web.FormString(c, "physio"),
	}
	if in.Ages, err = web.FormInt(c, "ages"); err != nil {
		return err
	}
	if in.IsActive, err = web.FormBool(c, "isActive"); err != nil {
		return err
	}
	if in.IsYouth, err = web.FormBool(c, "isYouth"); err != nil {
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

// remove deletes a team, unlinking its players, sponsors and managers.
//
//	@Summary	Delete a team
//	@Tags		teams
//	@Param		id	path	int	true	"Team ID"
//	@Success	204
//	@Failure	404	{object}	web.ErrorResponse
//	@Router		/teams/{id} [delete]
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
