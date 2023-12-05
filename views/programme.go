package views

import (
	"fmt"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/COMTOP1/AFC-GO/programme"
	"github.com/COMTOP1/AFC-GO/templates"
	"github.com/COMTOP1/AFC-GO/user"
)

type ProgrammeTemplateStruct struct {
	Year           int
	Programmes     []ProgrammeTemplate
	Seasons        []programme.Season
	SelectedSeason int
	User           user.User
	Context        *Context
}

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

	data := ProgrammeTemplateStruct{
		Year:       year,
		Programmes: DBProgrammesToTemplateFormat(programmesDB, seasonsDB),
		Seasons:    seasonsDB,
		User:       c1.User,
		Context:    c1,
	}

	return v.template.RenderTemplate(c.Response().Writer, data, templates.ProgrammesTemplate, templates.RegularType)
}

func (v *Views) ProgrammesSeasonsFunc(c echo.Context) error {
	c1 := v.getSessionData(c)

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return fmt.Errorf("failed to get id for programmesSeason: %w", err)
	}

	if id == 0 {
		return c.Redirect(http.StatusFound, "/programmes")
	}

	seasonDB, err := v.programme.GetSeason(c.Request().Context(), programme.Season{ID: id})
	if err != nil {
		return fmt.Errorf("failed to get season for programmesSeason: %w", err)
	}

	programmesDB, err := v.programme.GetProgrammesSeason(c.Request().Context(), seasonDB)
	if err != nil {
		return fmt.Errorf("failed to get programmes for programme: %w", err)
	}

	seasonsDB, err := v.programme.GetSeasons(c.Request().Context())
	if err != nil {
		return fmt.Errorf("failed to get seasons for programme: %w", err)
	}

	year, _, _ := time.Now().Date()

	data := ProgrammeTemplateStruct{
		Year:           year,
		Programmes:     DBProgrammesToTemplateFormat(programmesDB, seasonsDB),
		Seasons:        seasonsDB,
		SelectedSeason: id,
		User:           c1.User,
		Context:        c1,
	}

	return v.template.RenderTemplate(c.Response().Writer, data, templates.ProgrammesTemplate, templates.RegularType)
}

func (v *Views) ProgrammeSeasonSelectFunc(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		seasonID, err := strconv.Atoi(c.FormValue("season"))
		if err != nil {
			return fmt.Errorf("failed to parse season for programmeSelect: %w", err)
		}

		if seasonID == 0 {
			return c.Redirect(http.StatusFound, "/programmes")
		}

		return c.Redirect(http.StatusFound, fmt.Sprintf("/programmes/%d", seasonID))
	}
	return echo.NewHTTPError(http.StatusMethodNotAllowed, fmt.Errorf("invalid method used"))
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
