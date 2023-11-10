package views

import (
	"fmt"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/COMTOP1/AFC-GO/team"
	"github.com/COMTOP1/AFC-GO/templates"
	"github.com/COMTOP1/AFC-GO/user"
)

type (
	TeamsTemplate struct {
		ID       int
		Name     string
		IsActive bool
	}
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
		Teams []TeamsTemplate
		User  user.User
	}{
		Year:  year,
		Teams: DBTeamsToTemplateFormat(teams),
		User:  c1.User,
	}

	return v.template.RenderTemplate(c.Response().Writer, data, templates.TeamsTemplate, templates.RegularType)
}

func DBTeamsToTemplateFormat(teams []team.Team) []TeamsTemplate {
	teamsTemplate := make([]TeamsTemplate, 0, len(teams))
	for _, t := range teams {
		var t1 TeamsTemplate
		t1.ID = t.ID
		t1.Name = t.Name
		t1.IsActive = t.IsActive
		teamsTemplate = append(teamsTemplate, t1)
	}
	return teamsTemplate
}

func (v *Views) TeamFunc(c echo.Context) error {
	_ = c
	return fmt.Errorf("not implemented yet")
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
