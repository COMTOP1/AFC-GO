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

func (v *Views) WhatsOnAddFunc(c echo.Context) error {
	_ = c
	dateOfEvent := c.Request().FormValue("dateOfEvent")

	dateOfEventParsed, err := time.Parse("02/01/2006", dateOfEvent)
	if err != nil {
		return fmt.Errorf("failed to parse dateOfEvent: %w", err)
	}
	_ = dateOfEventParsed
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
