package views

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/COMTOP1/AFC-GO/server/internal/legacy/templates"
	"github.com/COMTOP1/AFC-GO/server/internal/team"
	"github.com/COMTOP1/AFC-GO/server/internal/user"
)

func (v *Views) TeamsFunc(c echo.Context) error {
	spanCtx, span := tracer.Start(c.Request().Context(), "views.TeamsFunc")
	defer span.End()
	c.SetRequest(c.Request().WithContext(spanCtx))
	c1 := v.getSessionData(c)

	var teams []team.Team
	var err error

	if c1.User.ID != 0 {
		teams, err = v.team.GetTeams(c.Request().Context())
	} else {
		teams, err = v.team.GetTeamsActive(c.Request().Context())
	}
	if err != nil {
		return v.error(http.StatusInternalServerError, "failed to get teams",
			fmt.Errorf("failed to get teams for teams: %w", err))
	}

	year, _, _ := time.Now().Date()

	data := struct {
		Year         int
		VisitorCount int
		Teams        []TeamTemplate
		User         user.User
		Context      *Context
	}{
		Year:         year,
		VisitorCount: v.GetVisitorCount(),
		Teams:        DBTeamsToTemplateFormat(teams),
		User:         c1.User,
		Context:      c1,
	}

	return v.template.RenderTemplate(c.Request().Context(), c.Response().Writer, data, templates.TeamsTemplate, templates.RegularType)
}

func (v *Views) TeamFunc(c echo.Context) error {
	spanCtx, span := tracer.Start(c.Request().Context(), "views.TeamFunc")
	defer span.End()
	c.SetRequest(c.Request().WithContext(spanCtx))
	c1 := v.getSessionData(c)

	teamID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return v.error(http.StatusBadRequest, "invalid id provided for team",
			fmt.Errorf("failed to parse id for team, error: %w", err))
	}
	teamDB, err := v.team.GetTeam(c.Request().Context(), team.Team{ID: teamID})
	if err != nil {
		return v.error(http.StatusInternalServerError, fmt.Sprintf("failed to get team, team id: %d", teamID),
			fmt.Errorf("failed to get team for team, team id: %d, error: %w", teamID, err))
	}

	managersDB, err := v.user.GetUsersManagersTeam(c.Request().Context(), teamDB)
	if err != nil {
		return v.error(http.StatusInternalServerError,
			fmt.Sprintf("failed to get managers for team, team name: \"%s\"", teamDB.Name),
			fmt.Errorf("failed to get managers for team, team id: %d, error: %w", teamID, err))
	}

	sponsorsDB, err := v.sponsor.GetSponsorsTeam(c.Request().Context(), teamDB)
	if err != nil {
		return v.error(http.StatusInternalServerError,
			fmt.Sprintf("failed to get sponsors for team, team name: \"%s\"", teamDB.Name),
			fmt.Errorf("failed to get sponsors for team, team id: %d, error: %w", teamID, err))
	}

	playersDB, err := v.player.GetPlayersTeam(c.Request().Context(), teamDB)
	if err != nil {
		return v.error(http.StatusInternalServerError,
			fmt.Sprintf("failed to get players for team, team name: \"%s\"", teamDB.Name),
			fmt.Errorf("failed to get players for team, team id: %d, error: %w", teamID, err))
	}

	year, _, _ := time.Now().Date()

	data := struct {
		Year         int
		VisitorCount int
		Team         team.Team
		Managers     []ManagerTemplate
		Sponsors     []SponsorTemplate
		Players      []PlayerTemplate
		User         user.User
		Context      *Context
	}{
		Year:         year,
		VisitorCount: v.GetVisitorCount(),
		Team:         teamDB,
		Managers:     DBManagersToTemplateFormat(managersDB),
		Sponsors:     DBSponsorsToTemplateFormat(sponsorsDB),
		Players:      DBPlayersTeamToTemplateFormat(playersDB, teamDB.IsYouth),
		User:         c1.User,
		Context:      c1,
	}

	return v.template.RenderTemplate(c.Request().Context(), c.Response().Writer, data, templates.TeamTemplate, templates.RegularType)
}

// legacyTeamInput reads the legacy team form. Checkbox fields are "Y" when
// ticked; unticked now means false (see Plan Decision on checkbox fields).
func legacyTeamInput(c echo.Context) (team.CreateInput, error) {
	ages, err := strconv.Atoi(c.FormValue("ages"))
	if err != nil {
		return team.CreateInput{}, fmt.Errorf("failed to parse ages: %w", err)
	}
	return team.CreateInput{
		Name:        c.FormValue("name"),
		Description: c.FormValue("description"),
		League:      c.FormValue("league"),
		Division:    c.FormValue("division"),
		LeagueTable: c.FormValue("leagueTable"),
		Fixtures:    c.FormValue("fixtures"),
		Coach:       c.FormValue("coach"),
		Physio:      c.FormValue("physio"),
		IsActive:    formYes(c, "isActive"),
		IsYouth:     formYes(c, "isYouth"),
		Ages:        ages,
	}, nil
}

func (v *Views) TeamAddFunc(c echo.Context) error {
	spanCtx, span := tracer.Start(c.Request().Context(), "views.TeamAddFunc")
	defer span.End()
	c.SetRequest(c.Request().WithContext(spanCtx))
	if c.Request().Method != http.MethodPost {
		return v.invalidMethodUsed(c)
	}
	c1 := v.getSessionData(c)
	data := struct {
		Error string `json:"error"`
	}{}
	in, err := legacyTeamInput(c)
	if err != nil {
		data.Error = fmt.Sprintf("failed to parse team add: %+v", err)
		return c.JSON(http.StatusOK, data)
	}
	image, err := legacyUpload(c, "upload")
	if err != nil {
		data.Error = fmt.Sprintf("failed to get file for team add: %+v", err)
		return c.JSON(http.StatusOK, data)
	}
	created, err := v.teamSvc.Create(c.Request().Context(), in, image)
	if err != nil {
		slog.Info(fmt.Sprintf("failed to add team for team add, error: %+v", err))
		data.Error = fmt.Sprintf("failed to add team for team add: %+v", err)
		return c.JSON(http.StatusOK, data)
	}
	v.flash(c, c1, fmt.Sprintf("successfully added \"%s\"", created.Name))
	return c.JSON(http.StatusOK, data)
}

func (v *Views) TeamEditFunc(c echo.Context) error {
	spanCtx, span := tracer.Start(c.Request().Context(), "views.TeamEditFunc")
	defer span.End()
	c.SetRequest(c.Request().WithContext(spanCtx))
	if c.Request().Method != http.MethodPost {
		return v.invalidMethodUsed(c)
	}
	c1 := v.getSessionData(c)
	teamID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return fmt.Errorf("failed to parse id for team edit: %w", err)
	}
	data := struct {
		Error string `json:"error"`
	}{}
	in, err := legacyTeamInput(c)
	if err != nil {
		data.Error = fmt.Sprintf("failed to parse team edit: %+v", err)
		return c.JSON(http.StatusOK, data)
	}
	remove := c.FormValue("removeTeamImage")
	if remove != "" && remove != "Y" {
		data.Error = "failed to parse removeTeamImage for team edit: " + remove
		return c.JSON(http.StatusOK, data)
	}
	image, err := legacyUpload(c, "upload")
	if err != nil {
		data.Error = fmt.Sprintf("failed to get file for team edit: %+v", err)
		return c.JSON(http.StatusOK, data)
	}
	updated, err := v.teamSvc.Update(c.Request().Context(), teamID, team.UpdateInput{
		Name: &in.Name, Description: &in.Description, League: &in.League, Division: &in.Division,
		LeagueTable: &in.LeagueTable, Fixtures: &in.Fixtures, Coach: &in.Coach, Physio: &in.Physio,
		IsActive: &in.IsActive, IsYouth: &in.IsYouth, Ages: &in.Ages, RemoveImage: remove == "Y",
	}, image)
	if err != nil {
		slog.Info(fmt.Sprintf("failed to edit team for team edit, team id: %d, error: %+v", teamID, err))
		data.Error = fmt.Sprintf("failed to edit team for team edit: %+v", err)
		return c.JSON(http.StatusOK, data)
	}
	v.flash(c, c1, fmt.Sprintf("successfully edited \"%s\"", updated.Name))
	return c.JSON(http.StatusOK, data)
}

func (v *Views) TeamDeleteFunc(c echo.Context) error {
	spanCtx, span := tracer.Start(c.Request().Context(), "views.TeamDeleteFunc")
	defer span.End()
	c.SetRequest(c.Request().WithContext(spanCtx))
	if c.Request().Method != http.MethodPost {
		return v.invalidMethodUsed(c)
	}
	c1 := v.getSessionData(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return fmt.Errorf("failed to get id for team delete, error: %w", err)
	}
	deleted, err := v.teamSvc.Delete(c.Request().Context(), id)
	if err != nil {
		return fmt.Errorf("failed to delete team for team delete, team id: %d, error: %w", id, err)
	}
	v.flash(c, c1, fmt.Sprintf("successfully deleted \"%s\"", deleted.Name))
	return c.Redirect(http.StatusFound, "/teams")
}
