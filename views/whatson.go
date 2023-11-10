package views

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/COMTOP1/AFC-GO/templates"
	"github.com/COMTOP1/AFC-GO/user"
	"github.com/COMTOP1/AFC-GO/whatson"
)

func (v *Views) WhatsOnFunc(c echo.Context) error {
	c1 := v.getSessionData(c)

	var w1 []whatson.WhatsOn
	var err error

	w1, err = v.whatsOn.GetWhatsOn(c.Request().Context())
	if err != nil {
		return fmt.Errorf("failed to get whatsOn for whatsOn: %w", err)
	}

	year, _, _ := time.Now().Date()

	data := struct {
		Year    int
		WhatsOn []WhatsOnTemplate
		User    user.User
	}{
		Year:    year,
		WhatsOn: DBWhatsOnToTemplateFormat(w1),
		User:    c1.User,
	}

	return v.template.RenderTemplate(c.Response().Writer, data, templates.WhatsOnTemplate, templates.RegularType)
}

func DBWhatsOnToTemplateFormat(whatsOnDB []whatson.WhatsOn) []WhatsOnTemplate {
	whatsOnTemplate := make([]WhatsOnTemplate, 0, len(whatsOnDB))
	for _, w1 := range whatsOnDB {
		var w WhatsOnTemplate
		w.ID = w1.ID
		w.Title = w1.Title
		w.Date = time.UnixMilli(w1.TempDate).Format("2006-01-02 15:04:05 MST")
		w.DateOfEvent = time.UnixMilli(w1.TempDOE).Format("2006-01-02")
		whatsOnTemplate = append(whatsOnTemplate, w)
	}
	return whatsOnTemplate
}

func (v *Views) WhatsOnArticleFunc(c echo.Context) error {
	c1 := v.getSessionData(c)

	whatsOnID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("failed to parse id for whatsOnArticle: %w", err))
	}
	whatsOnFromDB, err := v.whatsOn.GetWhatsOnArticle(c.Request().Context(), whatson.WhatsOn{ID: whatsOnID})
	if err != nil {
		return fmt.Errorf("failed to get whatsOn for whatsOnArticle: %w", err)
	}

	year, _, _ := time.Now().Date()

	data := struct {
		Year    int
		WhatsOn WhatsOnTemplate
		User    user.User
	}{
		Year:    year,
		WhatsOn: DBWhatsOnToArticleTemplateFormat(whatsOnFromDB),
		User:    c1.User,
	}

	return v.template.RenderTemplate(c.Response().Writer, data, templates.WhatsOnArticleTemplate, templates.RegularType)
}

func DBWhatsOnToArticleTemplateFormat(whatsOnDB whatson.WhatsOn) WhatsOnTemplate {
	var w WhatsOnTemplate
	w.ID = whatsOnDB.ID
	w.Title = whatsOnDB.Title
	w.Content = whatsOnDB.Content
	w.Date = time.UnixMilli(whatsOnDB.TempDate).Format("2006-01-02 15:04:05 MST")
	w.DateOfEvent = time.UnixMilli(whatsOnDB.TempDOE).Format("2006-01-02")
	return w
}

func (v *Views) WhatsOnAddFunc(c echo.Context) error {
	_ = c
	return fmt.Errorf("not implemented yet")
}

func (v *Views) WhatsOnEditFunc(c echo.Context) error {
	_ = c
	return fmt.Errorf("not implemented yet")
}

func (v *Views) WhatsOnDeleteFunc(c echo.Context) error {
	_ = c
	return fmt.Errorf("not implemented yet")
}
