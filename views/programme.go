package views

import (
	"fmt"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"

	"github.com/COMTOP1/AFC-GO/programme"
	"github.com/COMTOP1/AFC-GO/templates"
	"github.com/COMTOP1/AFC-GO/user"
)

type (
	ProgrammeTemplate struct {
		ID              int
		Name            string
		DateOfProgramme string
		Season          SeasonTemplate
	}

	SeasonTemplate struct {
		ID    int
		Name  string
		Valid bool
	}
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

	programmesTemplate, err := DBProgrammesToTemplateFormat(programmesDB, seasonsDB)
	if err != nil {
		return fmt.Errorf("failed to parse programmes for programmes: %w", err)
	}

	year, _, _ := time.Now().Date()

	data := struct {
		Year       int
		Programmes []ProgrammeTemplate
		Seasons    []programme.Season
		User       user.User
	}{
		Year:       year,
		Programmes: programmesTemplate,
		Seasons:    seasonsDB,
		User:       c1.User,
	}

	return v.template.RenderTemplate(c.Response().Writer, data, templates.ProgrammesTemplate, templates.RegularType)
}

func DBProgrammesToTemplateFormat(programmesDB []programme.Programme, seasonsDB []programme.Season) ([]ProgrammeTemplate, error) {
	programmesTemplate := make([]ProgrammeTemplate, 0, len(programmesDB))
	for _, programmeDB := range programmesDB {
		var programmeTemplate ProgrammeTemplate
		programmeTemplate.ID = programmeDB.ID
		programmeTemplate.Name = programmeDB.Name
		programmeTemplate.DateOfProgramme = time.UnixMilli(programmeDB.TempDOP).Format("2006-01-02")
		found := false
		for _, seasonDB := range seasonsDB {
			if seasonDB.ID == programmeDB.SeasonID {
				var seasonTemplate SeasonTemplate
				seasonTemplate.ID = seasonDB.ID
				seasonTemplate.Name = seasonDB.Season
				seasonTemplate.Valid = true
				programmeTemplate.Season = seasonTemplate
				found = true
				break
			}
		}
		if !found {
			log.Printf("failed to find season for programme: %d", programmeDB.ID)
			programmeTemplate.Season = SeasonTemplate{Valid: false}
		}
		programmesTemplate = append(programmesTemplate, programmeTemplate)
	}
	return programmesTemplate, nil
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
