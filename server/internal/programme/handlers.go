package programme

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/COMTOP1/AFC-GO/server/internal/svcerr"
	"github.com/COMTOP1/AFC-GO/server/internal/web"
)

// Handlers serves /programmes and /seasons.
type Handlers struct {
	svc *Service
}

func NewHandlers(svc *Service) *Handlers {
	return &Handlers{svc: svc}
}

// Register mounts the programme and season routes on the /api/v1 group.
func (h *Handlers) Register(g *echo.Group, guards web.Guards) {
	g.GET("/programmes", h.list)
	g.POST("/programmes", h.create, guards.Editor)
	g.DELETE("/programmes/:id", h.remove, guards.Editor)
	g.GET("/seasons", h.seasons)
	g.POST("/seasons", h.createSeason, guards.Editor)
	g.PATCH("/seasons/:id", h.renameSeason, guards.Editor)
	g.DELETE("/seasons/:id", h.deleteSeason, guards.Editor)
}

// SeasonInput is the JSON body for creating or renaming a season.
type SeasonInput struct {
	Season string `json:"season"`
}

// list returns programmes, optionally for one season.
//
//	@Summary	List programmes
//	@Tags		programmes
//	@Produce	json
//	@Param		season	query		int	false	"Season ID; omit for all"
//	@Success	200		{array}		programme.Public
//	@Failure	404		{object}	web.ErrorResponse
//	@Router		/programmes [get]
func (h *Handlers) list(c echo.Context) error {
	seasonID := 0
	if raw := c.QueryParam("season"); raw != "" {
		id, err := strconv.Atoi(raw)
		if err != nil || id < 0 {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid season")
		}
		seasonID = id
	}
	out, err := h.svc.List(c.Request().Context(), seasonID)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, out)
}

// create uploads a programme.
//
//	@Summary	Upload a programme
//	@Tags		programmes
//	@Accept		multipart/form-data
//	@Produce	json
//	@Param		name		formData	string	true	"Name"
//	@Param		date		formData	string	true	"YYYY-MM-DD"
//	@Param		seasonId	formData	int		false	"Season ID (0 or omitted for none)"
//	@Param		file		formData	file	true	"Programme PDF"
//	@Success	201			{object}	programme.Public
//	@Failure	422			{object}	web.ErrorResponse
//	@Router		/programmes [post]
func (h *Handlers) create(c echo.Context) error {
	date, err := web.FormDate(c, "date")
	if err != nil {
		return err
	}
	file, err := web.FormFile(c, "file")
	if err != nil {
		return err
	}
	in := CreateInput{Name: c.FormValue("name")}
	if date != nil {
		in.Date = *date
	}
	if raw := c.FormValue("seasonId"); raw != "" {
		if in.SeasonID, err = strconv.Atoi(raw); err != nil {
			return svcerr.InvalidField("seasonId", "season must be a number")
		}
	}
	out, err := h.svc.Create(c.Request().Context(), in, file)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, out)
}

// remove deletes a programme.
//
//	@Summary	Delete a programme
//	@Tags		programmes
//	@Param		id	path	int	true	"Programme ID"
//	@Success	204
//	@Failure	404	{object}	web.ErrorResponse
//	@Router		/programmes/{id} [delete]
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

// seasons lists programme seasons.
//
//	@Summary	List seasons
//	@Tags		programmes
//	@Produce	json
//	@Success	200	{array}	programme.PublicSeason
//	@Router		/seasons [get]
func (h *Handlers) seasons(c echo.Context) error {
	out, err := h.svc.Seasons(c.Request().Context())
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, out)
}

// createSeason adds a season.
//
//	@Summary	Create a season
//	@Tags		programmes
//	@Accept		json
//	@Produce	json
//	@Param		input	body		programme.SeasonInput	true	"Season name"
//	@Success	201		{object}	programme.PublicSeason
//	@Failure	422		{object}	web.ErrorResponse
//	@Router		/seasons [post]
func (h *Handlers) createSeason(c echo.Context) error {
	var in SeasonInput
	if err := web.BindJSON(c, &in); err != nil {
		return err
	}
	out, err := h.svc.CreateSeason(c.Request().Context(), in.Season)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, out)
}

// renameSeason renames a season.
//
//	@Summary	Rename a season
//	@Tags		programmes
//	@Accept		json
//	@Produce	json
//	@Param		id		path		int						true	"Season ID"
//	@Param		input	body		programme.SeasonInput	true	"New name"
//	@Success	200		{object}	programme.PublicSeason
//	@Failure	404		{object}	web.ErrorResponse
//	@Failure	422		{object}	web.ErrorResponse
//	@Router		/seasons/{id} [patch]
func (h *Handlers) renameSeason(c echo.Context) error {
	id, err := web.ParamID(c, "id")
	if err != nil {
		return err
	}
	var in SeasonInput
	if err = web.BindJSON(c, &in); err != nil {
		return err
	}
	out, err := h.svc.RenameSeason(c.Request().Context(), id, in.Season)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, out)
}

// deleteSeason deletes a season; its programmes are kept but unlinked.
//
//	@Summary	Delete a season
//	@Tags		programmes
//	@Param		id	path	int	true	"Season ID"
//	@Success	204
//	@Failure	404	{object}	web.ErrorResponse
//	@Router		/seasons/{id} [delete]
func (h *Handlers) deleteSeason(c echo.Context) error {
	id, err := web.ParamID(c, "id")
	if err != nil {
		return err
	}
	if _, err = h.svc.DeleteSeason(c.Request().Context(), id); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}
