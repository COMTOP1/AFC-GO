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
	VisitorCount   int
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
		return v.error(http.StatusInternalServerError, "failed to get programmes",
			fmt.Errorf("failed to get programmes for programmes: %w", err))
	}

	seasonsDB, err := v.programme.GetSeasons(c.Request().Context())
	if err != nil {
		return v.error(http.StatusInternalServerError, "failed to get seasons",
			fmt.Errorf("failed to get seasons for programmes: %w", err))
	}

	year, _, _ := time.Now().Date()

	data := ProgrammeTemplateStruct{
		Year:         year,
		VisitorCount: v.GetVisitorCount(),
		Programmes:   DBProgrammesToTemplateFormat(programmesDB, seasonsDB),
		Seasons:      seasonsDB,
		User:         c1.User,
		Context:      c1,
	}

	return v.template.RenderTemplate(c.Response().Writer, data, templates.ProgrammesTemplate, templates.RegularType)
}

func (v *Views) ProgrammesSeasonsFunc(c echo.Context) error {
	c1 := v.getSessionData(c)

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return v.error(http.StatusBadRequest, "invalid id provided for programmes seasons",
			fmt.Errorf("failed to parse id for programmes seasons, error: %w", err))
	}

	if id == 0 {
		return c.Redirect(http.StatusFound, "/programmes")
	}

	seasonDB, err := v.programme.GetSeason(c.Request().Context(), programme.Season{ID: id})
	if err != nil {
		return v.error(http.StatusInternalServerError, "failed to get season",
			fmt.Errorf("failed to get season for programmes seasons, season id: %d, error: %w", id, err))
	}

	programmesDB, err := v.programme.GetProgrammesSeason(c.Request().Context(), seasonDB)
	if err != nil {
		return v.error(http.StatusInternalServerError, "failed to get programmes, season: "+seasonDB.Season,
			fmt.Errorf("failed to get programmes for programmes season, season id: %d, error:: %w", id, err))
	}

	seasonsDB, err := v.programme.GetSeasons(c.Request().Context())
	if err != nil {
		return v.error(http.StatusInternalServerError, "failed to get seasons",
			fmt.Errorf("failed to get seasons for programmes seasons, season id: %d, error: %w", id, err))
	}

	year, _, _ := time.Now().Date()

	data := ProgrammeTemplateStruct{
		Year:           year,
		VisitorCount:   v.GetVisitorCount(),
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
			return v.error(http.StatusBadRequest, "invalid id provided for programme seasons select",
				fmt.Errorf("failed to parse season for programme season select, error: %w", err))
		}

		if seasonID == 0 {
			return c.Redirect(http.StatusFound, "/programmes")
		}

		return c.Redirect(http.StatusFound, fmt.Sprintf("/programmes/%d", seasonID))
	}
	return v.invalidMethodUsed(c)
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
			log.Printf("failed to parse programmeSeason for programme add, error: %+v", err)
			data.Error = fmt.Sprintf("failed to parse programmeSeason for programme add: %+v", err)
			return c.JSON(http.StatusOK, data)
		}

		dateOfProgramme := c.Request().FormValue("dateOfProgramme")

		parsed, err := time.Parse("02/01/2006", dateOfProgramme)
		if err != nil {
			log.Printf("failed to parse dateOfProgramme for programme add, error: %+v", err)
			data.Error = fmt.Sprintf("failed to parse dateOfProgramme for programme add: %+v", err)
			return c.JSON(http.StatusOK, data)
		}

		file, err := c.FormFile("upload")
		if err != nil {
			log.Printf("failed to get file for programme add, error: %+v", err)
			data.Error = fmt.Sprintf("failed to get file for programme add: %+v", err)
			return c.JSON(http.StatusOK, data)
		}
		fileName, err := v.fileUpload(file)
		if err != nil {
			log.Printf("failed to upload file for programme add, error: %+v", err)
			data.Error = fmt.Sprintf("failed to upload file for programme add: %+v", err)
			return c.JSON(http.StatusOK, data)
		}

		_, err = v.programme.AddProgramme(c.Request().Context(), programme.Programme{Name: name, FileName: fileName, DateOfProgramme: parsed, SeasonID: programmeSeason})
		if err != nil {
			log.Printf("failed to add programme for programme add, error: %+v", err)
			data.Error = fmt.Sprintf("failed to add programme for programme add: %+v", err)
			return c.JSON(http.StatusOK, data)
		}

		c1.Message = fmt.Sprintf("successfully added \"%s\"", name)
		c1.MsgType = "is-success"
		err = v.setMessagesInSession(c, c1)
		if err != nil {
			log.Printf("failed to set data for programme add, error: %+v", err)
		}

		return c.JSON(http.StatusOK, data)
	}
	return v.invalidMethodUsed(c)
}

func (v *Views) ProgrammeDeleteFunc(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		c1 := v.getSessionData(c)

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return fmt.Errorf("failed to get id for programme delete, error: %w", err)
		}

		programmeDB, err := v.programme.GetProgramme(c.Request().Context(), programme.Programme{ID: id})
		if err != nil {
			return fmt.Errorf("failed to get programme for programme delete, programme id: %d, error: %w", id, err)
		}

		err = os.Remove(filepath.Join(v.conf.FileDir, programmeDB.FileName))
		if err != nil {
			log.Printf("failed to delete programme image for programme delete, programme id: %d, error: %+v", id, err)
		}

		err = v.programme.DeleteProgramme(c.Request().Context(), programmeDB)
		if err != nil {
			return fmt.Errorf("failed to delete programme for programme delete, programme id: %d, error: %w", id, err)
		}

		c1.Message = fmt.Sprintf("successfully deleted \"%s\"", programmeDB.Name)
		c1.MsgType = "is-success"
		err = v.setMessagesInSession(c, c1)
		if err != nil {
			log.Printf("failed to set data for programme delete, programme id: %d, error: %+v", id, err)
		}

		return c.Redirect(http.StatusFound, "/programmes")
	}
	return v.invalidMethodUsed(c)
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
			log.Printf("failed to add season for season add, error: %+v", err)
			data.Error = fmt.Sprintf("failed to add season for season add: %+v", err)
			return c.JSON(http.StatusOK, data)
		}

		c1.Message = fmt.Sprintf("successfully added \"%s\"", season)
		c1.MsgType = "is-success"
		err = v.setMessagesInSession(c, c1)
		if err != nil {
			log.Printf("failed to set data for programme season add, error: %+v", err)
		}

		return c.JSON(http.StatusOK, data)
	}
	return v.invalidMethodUsed(c)
}

func (v *Views) ProgrammeSeasonEditFunc(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		c1 := v.getSessionData(c)

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return fmt.Errorf("failed to get id for programme season edit, error: %w", err)
		}

		seasonDB, err := v.programme.GetSeason(c.Request().Context(), programme.Season{ID: id})
		if err != nil {
			return fmt.Errorf("failed to get season for programme season edit, season id: %d, error: %w", id, err)
		}

		seasonDB.Season = c.FormValue("season")

		data := struct {
			Error string `json:"error"`
		}{
			Error: "",
		}

		_, err = v.programme.EditSeason(c.Request().Context(), seasonDB)
		if err != nil {
			log.Printf("failed to edit season for season edit, season id: %d, error: %+v", id, err)
			data.Error = fmt.Sprintf("failed to edit season for season edit: %+v", err)
			return c.JSON(http.StatusOK, data)
		}

		c1.Message = fmt.Sprintf("successfully edited \"%s\"", seasonDB.Season)
		c1.MsgType = "is-success"
		err = v.setMessagesInSession(c, c1)
		if err != nil {
			log.Printf("failed to set data for programme season edit, season id: %d, error: %+v", id, err)
		}

		return c.JSON(http.StatusOK, data)
	}
	return v.invalidMethodUsed(c)
}

func (v *Views) ProgrammeSeasonDeleteFunc(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		c1 := v.getSessionData(c)

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return fmt.Errorf("failed to get id for programme season delete, error: %w", err)
		}

		seasonDB, err := v.programme.GetSeason(c.Request().Context(), programme.Season{ID: id})
		if err != nil {
			return fmt.Errorf("failed to get season for programme season delete, season id: %d, error: %w", id, err)
		}

		programmesDB, err := v.programme.GetProgrammesSeason(c.Request().Context(), seasonDB)
		if err != nil {
			return fmt.Errorf("failed to get programmes for programme season delete, season id: %d, error: %w", id, err)
		}

		for _, programmeDB := range programmesDB {
			programmeDB.SeasonID = 0
			_, err = v.programme.EditProgramme(c.Request().Context(), programmeDB)
			if err != nil {
				return fmt.Errorf("failed to edit programme for programme season delete, season id: %d, error: %w", id, err)
			}
		}

		err = v.programme.DeleteSeason(c.Request().Context(), seasonDB)
		if err != nil {
			return fmt.Errorf("failed to delete season for programme season delete, season id: %d, error: %w", id, err)
		}

		c1.Message = fmt.Sprintf("successfully deleted \"%s\"", seasonDB.Season)
		c1.MsgType = "is-success"
		err = v.setMessagesInSession(c, c1)
		if err != nil {
			log.Printf("failed to set data for programme season delete, season id: %d, error: %+v", id, err)
		}

		return c.Redirect(http.StatusFound, "/programmes")
	}
	return v.invalidMethodUsed(c)
}
