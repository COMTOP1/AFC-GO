package views

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/COMTOP1/AFC-GO/team"
	"github.com/COMTOP1/AFC-GO/templates"
	"github.com/COMTOP1/AFC-GO/user"
)

func (v *Views) TeamsFunc(c echo.Context) error {
	c1 := v.getSessionData(c)

	var teams []team.Team
	var err error

	if c1.User.ID != 0 {
		teams, err = v.team.GetTeams(c.Request().Context())
	} else {
		teams, err = v.team.GetTeamsActive(c.Request().Context())
	}
	if err != nil {
		return fmt.Errorf("failed to get teams for teams: %w", err)
	}

	year, _, _ := time.Now().Date()

	data := struct {
		Year  int
		Teams []TeamTemplate
		User  user.User
	}{
		Year:  year,
		Teams: DBTeamsToTemplateFormat(teams),
		User:  c1.User,
	}

	return v.template.RenderTemplate(c.Response().Writer, data, templates.TeamsTemplate, templates.RegularType)
}

func (v *Views) TeamFunc(c echo.Context) error {
	c1 := v.getSessionData(c)

	teamID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("failed to parse id for team: %w", err))
	}
	teamDB, err := v.team.GetTeam(c.Request().Context(), team.Team{ID: teamID})
	if err != nil {
		return fmt.Errorf("failed to get team for team: %w", err)
	}

	managersDB, err := v.user.GetUsersManagersTeam(c.Request().Context(), teamDB)
	if err != nil {
		return fmt.Errorf("failed to get managers for team: %w, id: %d", err, teamID)
	}

	sponsorsDB, err := v.sponsor.GetSponsorsTeam(c.Request().Context(), teamDB)
	if err != nil {
		return fmt.Errorf("failed to get sponsors for team: %w, id: %d", err, teamID)
	}

	playersDB, err := v.player.GetPlayersTeam(c.Request().Context(), teamDB)
	if err != nil {
		return fmt.Errorf("failed to get players for team: %w, id: %d", err, teamID)
	}

	year, _, _ := time.Now().Date()

	data := struct {
		Year     int
		Team     team.Team
		Managers []string
		Sponsors []SponsorTemplate
		Players  []PlayerTemplate
		User     user.User
	}{
		Year:     year,
		Team:     teamDB,
		Managers: DBManagersToTemplateFormat(managersDB),
		Sponsors: DBSponsorsToTemplateFormat(sponsorsDB),
		Players:  DBPlayersTeamToTemplateFormat(playersDB),
		User:     c1.User,
	}

	return v.template.RenderTemplate(c.Response().Writer, data, templates.TeamTemplate, templates.RegularType)
}

func (v *Views) TeamAddFunc(c echo.Context) error {
	_ = c
	return fmt.Errorf("not implemented yet")
}

func (v *Views) TeamEditFunc(c echo.Context) error {
	_ = c
	return fmt.Errorf("not implemented yet")
}

func (v *Views) TeamDeleteFunc(c echo.Context) error {
	_ = c
	return fmt.Errorf("not implemented yet")
}
