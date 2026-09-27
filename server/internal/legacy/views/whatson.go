package views

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/COMTOP1/AFC-GO/server/internal/legacy/templates"
	"github.com/COMTOP1/AFC-GO/server/internal/user"
	"github.com/COMTOP1/AFC-GO/server/internal/whatson"
)

func (v *Views) WhatsOnFunc(c echo.Context) error {
	spanCtx, span := tracer.Start(c.Request().Context(), "views.WhatsOnFunc")
	defer span.End()
	c.SetRequest(c.Request().WithContext(spanCtx))
	c1 := v.getSessionData(c)

	whatsOnsDB, err := v.whatsOn.GetWhatsOn(c.Request().Context())
	if err != nil {
		return v.error(http.StatusInternalServerError, "failed to get whats on articles",
			fmt.Errorf("failed to get whats ons for whatsOn, error: %w", err))
	}

	year, _, _ := time.Now().Date()

	data := struct {
		Year         int
		VisitorCount int
		WhatsOn      []WhatsOnTemplate
		User         user.User
		TimePeriod   string
		Selected     string
		Context      *Context
	}{
		Year:         year,
		VisitorCount: v.GetVisitorCount(),
		WhatsOn:      DBWhatsOnToTemplateFormat(whatsOnsDB),
		User:         c1.User,
		TimePeriod:   "all the what's on articles",
		Selected:     "all",
		Context:      c1,
	}

	return v.template.RenderTemplate(c.Request().Context(), c.Response().Writer, data, templates.WhatsOnTemplate, templates.RegularType)
}

func (v *Views) WhatsOnTomePeriodFunc(c echo.Context) error {
	spanCtx, span := tracer.Start(c.Request().Context(), "views.WhatsOnTomePeriodFunc")
	defer span.End()
	c.SetRequest(c.Request().WithContext(spanCtx))
	c1 := v.getSessionData(c)

	timePeriod := c.Param("timePeriod")

	var whatsOnsDB []whatson.WhatsOn
	var err error

	switch timePeriod {
	case "future":
		whatsOnsDB, err = v.whatsOn.GetWhatsOnFuture(c.Request().Context())
		if err != nil {
			return v.error(http.StatusInternalServerError, "failed to get whats on articles in the future",
				fmt.Errorf("failed to get whats on future for whats on, error: %w", err))
		}
	case "past":
		whatsOnsDB, err = v.whatsOn.GetWhatsOnPast(c.Request().Context())
		if err != nil {
			return v.error(http.StatusInternalServerError, "failed to get whats on articles in the past",
				fmt.Errorf("failed to get whats on past for whats on, error: %w", err))
		}
	case "all":
		fallthrough
	default:
		return c.Redirect(http.StatusFound, "/whatson")
	}

	year, _, _ := time.Now().Date()

	data := struct {
		Year         int
		VisitorCount int
		WhatsOn      []WhatsOnTemplate
		User         user.User
		TimePeriod   string
		Selected     string
		Context      *Context
	}{
		Year:         year,
		VisitorCount: v.GetVisitorCount(),
		WhatsOn:      DBWhatsOnToTemplateFormat(whatsOnsDB),
		User:         c1.User,
		TimePeriod:   fmt.Sprintf("all the %s what's on articles", timePeriod),
		Selected:     timePeriod,
		Context:      c1,
	}

	return v.template.RenderTemplate(c.Request().Context(), c.Response().Writer, data, templates.WhatsOnTemplate, templates.RegularType)
}

func (v *Views) WhatsOnSelectFunc(c echo.Context) error {
	spanCtx, span := tracer.Start(c.Request().Context(), "views.WhatsOnSelectFunc")
	defer span.End()
	c.SetRequest(c.Request().WithContext(spanCtx))
	if c.Request().Method == http.MethodPost {
		timePeriod := c.FormValue("timePeriod")

		switch timePeriod {
		case "future":
			return c.Redirect(http.StatusFound, "/whatsonperiod/future")
		case "past":
			return c.Redirect(http.StatusFound, "/whatsonperiod/past")
		default:
			return c.Redirect(http.StatusFound, "/whatson")
		}
	}
	return v.invalidMethodUsed(c)
}

func (v *Views) WhatsOnArticleFunc(c echo.Context) error {
	spanCtx, span := tracer.Start(c.Request().Context(), "views.WhatsOnArticleFunc")
	defer span.End()
	c.SetRequest(c.Request().WithContext(spanCtx))
	c1 := v.getSessionData(c)

	whatsOnID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return v.error(http.StatusBadRequest, "invalid id provided for whats on article",
			fmt.Errorf("failed to parse whats on id for whats on article, error: %w", err))
	}
	whatsOnFromDB, err := v.whatsOn.GetWhatsOnArticle(c.Request().Context(), whatson.WhatsOn{ID: whatsOnID})
	if err != nil {
		return fmt.Errorf("failed to get whats on for whats on article, error: %w", err)
	}

	year, _, _ := time.Now().Date()

	data := struct {
		Year         int
		VisitorCount int
		WhatsOn      WhatsOnTemplate
		User         user.User
		Context      *Context
	}{
		Year:         year,
		VisitorCount: v.GetVisitorCount(),
		WhatsOn:      DBWhatsOnToArticleTemplateFormat(whatsOnFromDB),
		User:         c1.User,
		Context:      c1,
	}

	return v.template.RenderTemplate(c.Request().Context(), c.Response().Writer, data, templates.WhatsOnArticleTemplate, templates.RegularType)
}

func (v *Views) WhatsOnAddFunc(c echo.Context) error {
	spanCtx, span := tracer.Start(c.Request().Context(), "views.WhatsOnAddFunc")
	defer span.End()
	c.SetRequest(c.Request().WithContext(spanCtx))
	if c.Request().Method != http.MethodPost {
		return v.invalidMethodUsed(c)
	}
	c1 := v.getSessionData(c)
	data := struct {
		Error string `json:"error"`
	}{}

	dateOfEvent, err := time.Parse("02/01/2006", c.FormValue("dateOfEvent"))
	if err != nil {
		data.Error = fmt.Sprintf("failed to parse dateOfEvent for whats on add: %+v", err)
		return c.JSON(http.StatusOK, data)
	}
	image, err := legacyUpload(c, "upload")
	if err != nil {
		data.Error = fmt.Sprintf("failed to get file for whats on add: %+v", err)
		return c.JSON(http.StatusOK, data)
	}
	event, err := v.whatsOnSvc.Create(c.Request().Context(), whatson.CreateInput{
		Title:       c.FormValue("title"),
		Content:     c.FormValue("htmlContent"),
		DateOfEvent: dateOfEvent,
	}, image)
	if err != nil {
		slog.Info(fmt.Sprintf("failed to add whatsOn for whats on add, error: %+v", err))
		data.Error = fmt.Sprintf("failed to add whatsOn for whats on add: %+v", err)
		return c.JSON(http.StatusOK, data)
	}
	v.flash(c, c1, fmt.Sprintf("successfully added \"%s\"", event.Title))
	return c.JSON(http.StatusOK, data)
}

func (v *Views) WhatsOnEditFunc(c echo.Context) error {
	spanCtx, span := tracer.Start(c.Request().Context(), "views.WhatsOnEditFunc")
	defer span.End()
	c.SetRequest(c.Request().WithContext(spanCtx))
	if c.Request().Method != http.MethodPost {
		return v.invalidMethodUsed(c)
	}
	c1 := v.getSessionData(c)
	whatsOnID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("failed to parse id for whats on edit, error: %w", err))
	}
	data := struct {
		Error string `json:"error"`
	}{}

	dateOfEvent, err := time.Parse("02/01/2006", c.FormValue("dateOfEvent"))
	if err != nil {
		data.Error = fmt.Sprintf("failed to parse dateOfEvent for whats on edit: %+v", err)
		return c.JSON(http.StatusOK, data)
	}
	remove := c.FormValue("removeWhatsOnImage")
	if remove != "" && remove != "Y" {
		data.Error = "failed to parse removeWhatsOnImage for whats on edit, value: " + remove
		return c.JSON(http.StatusOK, data)
	}
	image, err := legacyUpload(c, "upload")
	if err != nil {
		data.Error = fmt.Sprintf("failed to get file for whats on edit: %+v", err)
		return c.JSON(http.StatusOK, data)
	}
	title, content := c.FormValue("title"), c.FormValue("htmlContent")
	event, err := v.whatsOnSvc.Update(c.Request().Context(), whatsOnID, whatson.UpdateInput{
		Title:       &title,
		Content:     &content,
		DateOfEvent: &dateOfEvent,
		RemoveImage: remove == "Y",
	}, image)
	if err != nil {
		slog.Info(fmt.Sprintf("failed to edit whatsOn for whats on edit, whats on id: %d, error: %+v", whatsOnID, err))
		data.Error = fmt.Sprintf("failed to edit whatsOn for whats on edit: %+v", err)
		return c.JSON(http.StatusOK, data)
	}
	v.flash(c, c1, fmt.Sprintf("successfully edited \"%s\"", event.Title))
	return c.JSON(http.StatusOK, data)
}

func (v *Views) WhatsOnDeleteFunc(c echo.Context) error {
	spanCtx, span := tracer.Start(c.Request().Context(), "views.WhatsOnDeleteFunc")
	defer span.End()
	c.SetRequest(c.Request().WithContext(spanCtx))
	if c.Request().Method != http.MethodPost {
		return v.invalidMethodUsed(c)
	}
	c1 := v.getSessionData(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return fmt.Errorf("failed to get id for whats on delete, error: %w", err)
	}
	event, err := v.whatsOnSvc.Delete(c.Request().Context(), id)
	if err != nil {
		return fmt.Errorf("failed to delete whatsOn for whats on delete, whats on id: %d, error: %w", id, err)
	}
	v.flash(c, c1, fmt.Sprintf("successfully deleted \"%s\"", event.Title))
	return c.Redirect(http.StatusFound, "/whatson")
}
