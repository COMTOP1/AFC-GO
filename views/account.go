package views

import (
	"time"

	"github.com/labstack/echo/v4"

	"github.com/COMTOP1/AFC-GO/templates"
	"github.com/COMTOP1/AFC-GO/user"
)

func (v *Views) AccountFunc(c echo.Context) error {
	c1 := v.getSessionData(c)

	year, _, _ := time.Now().Date()

	data := struct {
		Year    int
		User    UserTemplate
		Context *Context
	}{
		Year:    year,
		User:    DBUserToTemplateFormat(c1.User),
		Context: c1,
	}

	return v.template.RenderTemplate(c.Response().Writer, data, templates.AccountTemplate, templates.RegularType)
}
