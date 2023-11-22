package views

import (
	"fmt"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/COMTOP1/AFC-GO/programme"
	"github.com/COMTOP1/AFC-GO/templates"
	"github.com/COMTOP1/AFC-GO/user"
)

func (v *Views) ProgrammesFunc(c echo.Context) error {
	c1 := v.getSessionData(c)

	programmesDB, err := v.programme.GetProgrammes(c.Request().Context())
	if err != nil {
		return fmt.Errorf("failed to get programmes for programme: %w", err)
	}

	seasonsDB, err := v.programme.GetSeasons(c.Request().Context())
	if err != nil {
		return fmt.Errorf("failed to get seasons for programme: %w", err)
	}

	year, _, _ := time.Now().Date()

	data := struct {
		Year       int
		Programmes []ProgrammeTemplate
		Seasons    []programme.Season
		User       user.User
	}{
		Year:       year,
		Programmes: DBProgrammesToTemplateFormat(programmesDB, seasonsDB),
		Seasons:    seasonsDB,
		User:       c1.User,
	}

	return v.template.RenderTemplate(c.Response().Writer, data, templates.ProgrammesTemplate, templates.RegularType)
}

func (v *Views) ProgrammesSeasonsFunc(c echo.Context) error {
	_ = c
	return fmt.Errorf("not implemented yet")
}

func (v *Views) ProgrammeAddFunc(c echo.Context) error {
	_ = c
	return fmt.Errorf("not implemented yet")
}

func (v *Views) ProgrammeDeleteFunc(c echo.Context) error {
	_ = c
	return fmt.Errorf("not implemented yet")
}

func (v *Views) ProgrammeSeasonAddFunc(c echo.Context) error {
	_ = c
	return fmt.Errorf("not implemented yet")
}

func (v *Views) ProgrammeSeasonEditFunc(c echo.Context) error {
	_ = c
	return fmt.Errorf("not implemented yet")
}

func (v *Views) ProgrammeSeasonDeleteFunc(c echo.Context) error {
	_ = c
	return fmt.Errorf("not implemented yet")
}
