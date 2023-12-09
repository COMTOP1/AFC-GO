package views

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
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
	if c.Request().Method == http.MethodPost {
		c1 := v.getSessionData(c)

		name := c.FormValue("name")

		data := struct {
			Error string `json:"error"`
		}{
			Error: "",
		}

		programmeSeason, err := strconv.Atoi(c.FormValue("programmeSeason"))
		if err != nil {
			log.Printf("failed to parse programmeSeason for programmeAdd: %+v", err)
			data.Error = fmt.Sprintf("failed to parse programmeSeason for programmeAdd: %+v", err)
			return c.JSON(http.StatusOK, data)
		}

		dateOfProgramme := c.Request().FormValue("dateOfProgramme")

		parsed, err := time.Parse("02/01/2006", dateOfProgramme)
		if err != nil {
			log.Printf("failed to parse dateOfProgramme for programmeAdd: %+v", err)
			data.Error = fmt.Sprintf("failed to parse dateOfProgramme for programmeAdd: %+v", err)
			return c.JSON(http.StatusOK, data)
		}

		file, err := c.FormFile("upload")
		if err != nil {
			log.Printf("failed to get file for programmeAdd: %+v", err)
			data.Error = fmt.Sprintf("failed to get file for documentAdd: %+v", err)
			return c.JSON(http.StatusOK, data)
		}
		fileName, err := v.fileUpload(file)
		if err != nil {
			log.Printf("failed to upload file for programmeAdd: %+v", err)
			data.Error = fmt.Sprintf("failed to upload file for programmeAdd: %+v", err)
			return c.JSON(http.StatusOK, data)
		}

		_, err = v.programme.AddProgramme(c.Request().Context(), programme.Programme{Name: name, FileName: fileName, DateOfProgramme: parsed, SeasonID: programmeSeason})
		if err != nil {
			log.Printf("failed to add programme for programmeAdd: %+v", err)
			data.Error = fmt.Sprintf("failed to add programme for programmeAdd: %+v", err)
			return c.JSON(http.StatusOK, data)
		}

		c1.Message = fmt.Sprintf("successfully added \"%s\"", name)
		c1.MsgType = "is-success"
		err = v.setMessagesInSession(c, c1)
		if err != nil {
			log.Printf("failed to set data for programmeAdd: %+v", err)
		}

		return c.JSON(http.StatusOK, data)
	}
	return echo.NewHTTPError(http.StatusMethodNotAllowed, fmt.Errorf("invalid method used"))
}

func (v *Views) ProgrammeDeleteFunc(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		c1 := v.getSessionData(c)

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return fmt.Errorf("failed to get id for programmeDelete: %w", err)
		}

		programmeDB, err := v.programme.GetProgramme(c.Request().Context(), programme.Programme{ID: id})
		if err != nil {
			return fmt.Errorf("failed to get programme for programmeDelete: %w", err)
		}

		err = os.Remove(filepath.Join(v.conf.FileDir, programmeDB.FileName))
		if err != nil {
			log.Printf("failed to delete programme image for programmeDelete: %+v", err)
		}

		err = v.programme.DeleteProgramme(c.Request().Context(), programmeDB)
		if err != nil {
			return fmt.Errorf("failed to delete programme for programmeDelete: %w", err)
		}

		c1.Message = fmt.Sprintf("successfully deleted \"%s\"", programmeDB.Name)
		c1.MsgType = "is-success"
		err = v.setMessagesInSession(c, c1)
		if err != nil {
			log.Printf("failed to set data for programmeDelete: %+v", err)
		}

		return c.Redirect(http.StatusFound, "/programmes")
	}
	return echo.NewHTTPError(http.StatusMethodNotAllowed, fmt.Errorf("invalid method used"))
}

func (v *Views) ProgrammeSeasonAddFunc(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		c1 := v.getSessionData(c)

		season := c.FormValue("season")

		data := struct {
			Error string `json:"error"`
		}{
			Error: "",
		}

		_, err := v.programme.AddSeason(c.Request().Context(), programme.Season{Season: season})
		if err != nil {
			log.Printf("failed to add season for seasonAdd: %+v", err)
			data.Error = fmt.Sprintf("failed to add season for seasonAdd: %+v", err)
			return c.JSON(http.StatusOK, data)
		}

		c1.Message = fmt.Sprintf("successfully added \"%s\"", season)
		c1.MsgType = "is-success"
		err = v.setMessagesInSession(c, c1)
		if err != nil {
			log.Printf("failed to set data for programmeSeasonAdd: %+v", err)
		}

		return c.JSON(http.StatusOK, data)
	}
	return echo.NewHTTPError(http.StatusMethodNotAllowed, fmt.Errorf("invalid method used"))
}

func (v *Views) ProgrammeSeasonEditFunc(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		c1 := v.getSessionData(c)

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return fmt.Errorf("failed to get id for programmeSeasonEdit: %w", err)
		}

		seasonDB, err := v.programme.GetSeason(c.Request().Context(), programme.Season{ID: id})
		if err != nil {
			return fmt.Errorf("failed to get season for programmeSeasonEdit: %w", err)
		}

		seasonDB.Season = c.FormValue("season")

		data := struct {
			Error string `json:"error"`
		}{
			Error: "",
		}

		_, err = v.programme.EditSeason(c.Request().Context(), seasonDB)
		if err != nil {
			log.Printf("failed to edit season for seasonEdit: %+v", err)
			data.Error = fmt.Sprintf("failed to edit season for seasonEdit: %+v", err)
			return c.JSON(http.StatusOK, data)
		}

		c1.Message = fmt.Sprintf("successfully edited \"%s\"", seasonDB.Season)
		c1.MsgType = "is-success"
		err = v.setMessagesInSession(c, c1)
		if err != nil {
			log.Printf("failed to set data for programmeSeasonEdit: %+v", err)
		}

		return c.JSON(http.StatusOK, data)
	}
	return echo.NewHTTPError(http.StatusMethodNotAllowed, fmt.Errorf("invalid method used"))
}

func (v *Views) ProgrammeSeasonDeleteFunc(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		c1 := v.getSessionData(c)

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return fmt.Errorf("failed to get id for programmeSeasonDelete: %w", err)
		}

		seasonDB, err := v.programme.GetSeason(c.Request().Context(), programme.Season{ID: id})
		if err != nil {
			return fmt.Errorf("failed to get season for programmeSeasonDelete: %w", err)
		}

		programmesDB, err := v.programme.GetProgrammesSeason(c.Request().Context(), seasonDB)
		if err != nil {
			return fmt.Errorf("failed to get programmes for programmeSeasonDelete: %w", err)
		}

		for _, programmeDB := range programmesDB {
			programmeDB.SeasonID = 0
			_, err = v.programme.EditProgramme(c.Request().Context(), programmeDB)
			if err != nil {
				return fmt.Errorf("failed to edit programme for programmeSeasonDelete")
			}
		}

		err = v.programme.DeleteSeason(c.Request().Context(), seasonDB)
		if err != nil {
			return fmt.Errorf("failed to delete season for programmeSeasonDelete: %w", err)
		}

		c1.Message = fmt.Sprintf("successfully deleted \"%s\"", seasonDB.Season)
		c1.MsgType = "is-success"
		err = v.setMessagesInSession(c, c1)
		if err != nil {
			log.Printf("failed to set data for programmeSeasonDelete: %+v", err)
		}

		return c.Redirect(http.StatusFound, "/programmes")
	}
	return echo.NewHTTPError(http.StatusMethodNotAllowed, fmt.Errorf("invalid method used"))
}
