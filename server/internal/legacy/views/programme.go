package views

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/COMTOP1/AFC-GO/server/internal/legacy/templates"
	"github.com/COMTOP1/AFC-GO/server/internal/programme"
	"github.com/COMTOP1/AFC-GO/server/internal/user"
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
	spanCtx, span := tracer.Start(c.Request().Context(), "views.ProgrammesFunc")
	defer span.End()
	c.SetRequest(c.Request().WithContext(spanCtx))
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

	return v.template.RenderTemplate(c.Request().Context(), c.Response().Writer, data, templates.ProgrammesTemplate, templates.RegularType)
}

func (v *Views) ProgrammesSeasonsFunc(c echo.Context) error {
	spanCtx, span := tracer.Start(c.Request().Context(), "views.ProgrammesSeasonsFunc")
	defer span.End()
	c.SetRequest(c.Request().WithContext(spanCtx))
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

	return v.template.RenderTemplate(c.Request().Context(), c.Response().Writer, data, templates.ProgrammesTemplate, templates.RegularType)
}

func (v *Views) ProgrammeSeasonSelectFunc(c echo.Context) error {
	spanCtx, span := tracer.Start(c.Request().Context(), "views.ProgrammeSeasonSelectFunc")
	defer span.End()
	c.SetRequest(c.Request().WithContext(spanCtx))
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
	spanCtx, span := tracer.Start(c.Request().Context(), "views.ProgrammeAddFunc")
	defer span.End()
	c.SetRequest(c.Request().WithContext(spanCtx))
	if c.Request().Method != http.MethodPost {
		return v.invalidMethodUsed(c)
	}
	c1 := v.getSessionData(c)
	data := struct {
		Error string `json:"error"`
	}{}
	seasonID, err := strconv.Atoi(c.FormValue("programmeSeason"))
	if err != nil {
		data.Error = fmt.Sprintf("failed to parse programmeSeason for programme add: %+v", err)
		return c.JSON(http.StatusOK, data)
	}
	date, err := time.Parse("02/01/2006", c.FormValue("dateOfProgramme"))
	if err != nil {
		data.Error = fmt.Sprintf("failed to parse dateOfProgramme for programme add: %+v", err)
		return c.JSON(http.StatusOK, data)
	}
	file, err := legacyUpload(c, "upload")
	if err != nil {
		data.Error = fmt.Sprintf("failed to get file for programme add: %+v", err)
		return c.JSON(http.StatusOK, data)
	}
	created, err := v.programmeSvc.Create(c.Request().Context(), programme.CreateInput{
		Name: c.FormValue("name"), Date: date, SeasonID: seasonID,
	}, file)
	if err != nil {
		slog.Info(fmt.Sprintf("failed to add programme for programme add, error: %+v", err))
		data.Error = fmt.Sprintf("failed to add programme for programme add: %+v", err)
		return c.JSON(http.StatusOK, data)
	}
	v.flash(c, c1, fmt.Sprintf("successfully added \"%s\"", created.Name))
	return c.JSON(http.StatusOK, data)
}

func (v *Views) ProgrammeDeleteFunc(c echo.Context) error {
	spanCtx, span := tracer.Start(c.Request().Context(), "views.ProgrammeDeleteFunc")
	defer span.End()
	c.SetRequest(c.Request().WithContext(spanCtx))
	if c.Request().Method != http.MethodPost {
		return v.invalidMethodUsed(c)
	}
	c1 := v.getSessionData(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return fmt.Errorf("failed to get id for programme delete, error: %w", err)
	}
	deleted, err := v.programmeSvc.Delete(c.Request().Context(), id)
	if err != nil {
		return fmt.Errorf("failed to delete programme for programme delete, programme id: %d, error: %w", id, err)
	}
	v.flash(c, c1, fmt.Sprintf("successfully deleted \"%s\"", deleted.Name))
	return c.Redirect(http.StatusFound, "/programmes")
}

func (v *Views) ProgrammeSeasonAddFunc(c echo.Context) error {
	spanCtx, span := tracer.Start(c.Request().Context(), "views.ProgrammeSeasonAddFunc")
	defer span.End()
	c.SetRequest(c.Request().WithContext(spanCtx))
	if c.Request().Method != http.MethodPost {
		return v.invalidMethodUsed(c)
	}
	c1 := v.getSessionData(c)
	data := struct {
		Error string `json:"error"`
	}{}
	created, err := v.programmeSvc.CreateSeason(c.Request().Context(), c.FormValue("season"))
	if err != nil {
		data.Error = fmt.Sprintf("failed to add season for season add: %+v", err)
		return c.JSON(http.StatusOK, data)
	}
	v.flash(c, c1, fmt.Sprintf("successfully added \"%s\"", created.Name))
	return c.JSON(http.StatusOK, data)
}

func (v *Views) ProgrammeSeasonEditFunc(c echo.Context) error {
	spanCtx, span := tracer.Start(c.Request().Context(), "views.ProgrammeSeasonEditFunc")
	defer span.End()
	c.SetRequest(c.Request().WithContext(spanCtx))
	if c.Request().Method != http.MethodPost {
		return v.invalidMethodUsed(c)
	}
	c1 := v.getSessionData(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return fmt.Errorf("failed to get id for programme season edit, error: %w", err)
	}
	data := struct {
		Error string `json:"error"`
	}{}
	renamed, err := v.programmeSvc.RenameSeason(c.Request().Context(), id, c.FormValue("season"))
	if err != nil {
		data.Error = fmt.Sprintf("failed to edit season for season edit: %+v", err)
		return c.JSON(http.StatusOK, data)
	}
	v.flash(c, c1, fmt.Sprintf("successfully edited \"%s\"", renamed.Name))
	return c.JSON(http.StatusOK, data)
}

func (v *Views) ProgrammeSeasonDeleteFunc(c echo.Context) error {
	spanCtx, span := tracer.Start(c.Request().Context(), "views.ProgrammeSeasonDeleteFunc")
	defer span.End()
	c.SetRequest(c.Request().WithContext(spanCtx))
	if c.Request().Method != http.MethodPost {
		return v.invalidMethodUsed(c)
	}
	c1 := v.getSessionData(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return fmt.Errorf("failed to get id for programme season delete, error: %w", err)
	}
	deleted, err := v.programmeSvc.DeleteSeason(c.Request().Context(), id)
	if err != nil {
		return fmt.Errorf("failed to delete season for programme season delete, season id: %d, error: %w", id, err)
	}
	v.flash(c, c1, fmt.Sprintf("successfully deleted \"%s\"", deleted.Name))
	return c.Redirect(http.StatusFound, "/programmes")
}
