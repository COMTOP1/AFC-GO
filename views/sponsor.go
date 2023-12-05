package views

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/sponsor"
	"github.com/COMTOP1/AFC-GO/team"
	"github.com/COMTOP1/AFC-GO/templates"
	"github.com/COMTOP1/AFC-GO/user"
)

func (v *Views) SponsorsFunc(c echo.Context) error {
	c1 := v.getSessionData(c)

	sponsorsDB, err := v.sponsor.GetSponsors(c.Request().Context())
	if err != nil {
		return fmt.Errorf("failed to get sponsors for sponsors: %w", err)
	}

	teamsDB, err := v.team.GetTeams(c.Request().Context())
	if err != nil {
		return fmt.Errorf("failed to get teams for sponsors: %w", err)
	}

	year, _, _ := time.Now().Date()

	data := struct {
		Year     int
		Sponsors []SponsorTemplate
		Teams    []TeamTemplate
		User     user.User
		Context  *Context
	}{
		Year:     year,
		Sponsors: DBSponsorsToTemplateFormat(sponsorsDB),
		Teams:    DBTeamsToTemplateFormat(teamsDB),
		User:     c1.User,
		Context:  c1,
	}

	return v.template.RenderTemplate(c.Response().Writer, data, templates.SponsorsTemplate, templates.RegularType)
}

func (v *Views) SponsorAddFunc(c echo.Context) error {
	_ = c
	return fmt.Errorf("not implemented yet")
}

func (v *Views) SponsorDeleteFunc(c echo.Context) error {
	_ = c
	return fmt.Errorf("not implemented yet")
}
