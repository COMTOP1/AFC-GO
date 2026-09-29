package views

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/COMTOP1/AFC-GO/server/internal/legacy/templates"
	"github.com/COMTOP1/AFC-GO/server/internal/player"
	"github.com/COMTOP1/AFC-GO/server/internal/user"
)

func (v *Views) PlayersFunc(c echo.Context) error {
	spanCtx, span := tracer.Start(c.Request().Context(), "views.PlayersFunc")
	defer span.End()
	c.SetRequest(c.Request().WithContext(spanCtx))
	c1 := v.getSessionData(c)

	playersDB, err := v.player.GetPlayers(c.Request().Context())
	if err != nil {
		return fmt.Errorf("failed to get players for players, error: %w", err)
	}

	teamsDB, err := v.team.GetTeams(c.Request().Context())
	if err != nil {
		return fmt.Errorf("failed to get teams for players, error: %w", err)
	}

	year, _, _ := time.Now().Date()

	data := struct {
		Year         int
		VisitorCount int
		Players      []PlayerTemplate
		Teams        []TeamTemplate
		User         user.User
		Context      *Context
	}{
		Year:         year,
		VisitorCount: v.GetVisitorCount(),
		Players:      DBPlayersToTemplateFormat(playersDB, teamsDB),
		Teams:        DBTeamsToTemplateFormat(teamsDB),
		User:         c1.User,
		Context:      c1,
	}

	return v.template.RenderTemplate(c.Request().Context(), c.Response().Writer, data, templates.PlayersTemplate, templates.RegularType)
}

// legacyPlayerInput reads the legacy player form.
func legacyPlayerInput(c echo.Context) (player.CreateInput, error) {
	teamID, err := strconv.Atoi(c.FormValue("playerTeam"))
	if err != nil {
		return player.CreateInput{}, fmt.Errorf("failed to parse playerTeam: %w", err)
	}
	dob, err := time.Parse("02/01/2006", c.FormValue("dateOfBirth"))
	if err != nil {
		return player.CreateInput{}, fmt.Errorf("failed to parse dateOfBirth: %w", err)
	}
	return player.CreateInput{
		Name:        c.FormValue("name"),
		Position:    c.FormValue("position"),
		TeamID:      teamID,
		DateOfBirth: dob,
		IsCaptain:   formYes(c, "isCaptain"),
	}, nil
}

func (v *Views) PlayerAddFunc(c echo.Context) error {
	spanCtx, span := tracer.Start(c.Request().Context(), "views.PlayerAddFunc")
	defer span.End()
	c.SetRequest(c.Request().WithContext(spanCtx))
	if c.Request().Method != http.MethodPost {
		return v.invalidMethodUsed(c)
	}
	c1 := v.getSessionData(c)
	data := struct {
		Error string `json:"error"`
	}{}
	in, err := legacyPlayerInput(c)
	if err != nil {
		data.Error = fmt.Sprintf("failed to parse player add: %+v", err)
		return c.JSON(http.StatusOK, data)
	}
	image, err := legacyUpload(c, "upload")
	if err != nil {
		data.Error = fmt.Sprintf("failed to get file for player add: %+v", err)
		return c.JSON(http.StatusOK, data)
	}
	created, err := v.playerSvc.Create(c.Request().Context(), in, image)
	if err != nil {
		slog.Info(fmt.Sprintf("failed to add player for player add, error: %+v", err))
		data.Error = fmt.Sprintf("failed to add player for player add: %+v", err)
		return c.JSON(http.StatusOK, data)
	}
	v.flash(c, c1, fmt.Sprintf("successfully added \"%s\"", created.Name))
	return c.JSON(http.StatusOK, data)
}

func (v *Views) PlayerEditFunc(c echo.Context) error {
	spanCtx, span := tracer.Start(c.Request().Context(), "views.PlayerEditFunc")
	defer span.End()
	c.SetRequest(c.Request().WithContext(spanCtx))
	if c.Request().Method != http.MethodPost {
		return v.invalidMethodUsed(c)
	}
	c1 := v.getSessionData(c)
	playerID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return fmt.Errorf("failed to parse id for player edit, error: %w", err)
	}
	data := struct {
		Error string `json:"error"`
	}{}
	in, err := legacyPlayerInput(c)
	if err != nil {
		data.Error = fmt.Sprintf("failed to parse player edit: %+v", err)
		return c.JSON(http.StatusOK, data)
	}
	remove := c.FormValue("removePlayerImage")
	if remove != "" && remove != "Y" {
		data.Error = "failed to parse removePlayerImage for player edit: " + remove
		return c.JSON(http.StatusOK, data)
	}
	image, err := legacyUpload(c, "upload")
	if err != nil {
		data.Error = fmt.Sprintf("failed to get file for player edit: %+v", err)
		return c.JSON(http.StatusOK, data)
	}
	updated, err := v.playerSvc.Update(c.Request().Context(), playerID, player.UpdateInput{
		Name: &in.Name, Position: &in.Position, TeamID: &in.TeamID, DateOfBirth: &in.DateOfBirth,
		IsCaptain: &in.IsCaptain, RemoveImage: remove == "Y",
	}, image)
	if err != nil {
		slog.Info(fmt.Sprintf("failed to edit player for player edit, player id: %d, error: %+v", playerID, err))
		data.Error = fmt.Sprintf("failed to edit player for player edit: %+v", err)
		return c.JSON(http.StatusOK, data)
	}
	v.flash(c, c1, fmt.Sprintf("successfully edited \"%s\"", updated.Name))
	return c.JSON(http.StatusOK, data)
}

func (v *Views) PlayerDeleteFunc(c echo.Context) error {
	spanCtx, span := tracer.Start(c.Request().Context(), "views.PlayerDeleteFunc")
	defer span.End()
	c.SetRequest(c.Request().WithContext(spanCtx))
	if c.Request().Method != http.MethodPost {
		return v.invalidMethodUsed(c)
	}
	c1 := v.getSessionData(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return fmt.Errorf("failed to get id for player delete, error: %w", err)
	}
	deleted, err := v.playerSvc.Delete(c.Request().Context(), id)
	if err != nil {
		return fmt.Errorf("failed to delete player for player delete, player id: %d, error: %w", id, err)
	}
	v.flash(c, c1, fmt.Sprintf("successfully deleted \"%s\"", deleted.Name))
	return c.Redirect(http.StatusFound, "/players")
}
