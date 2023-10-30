package views

import (
	"time"

	"github.com/labstack/echo/v4"

	"github.com/COMTOP1/AFC-GO/templates"
	"github.com/COMTOP1/AFC-GO/user"
)

func (v *Views) InfoFunc(c echo.Context) error {
	c1 := v.getSessionData(c)

	year, _, _ := time.Now().Date()

	_ = c1.User

	data := struct {
		Year int
		User user.User
	}{
		Year: year,
		User: c1.User,
	}

	return v.template.RenderTemplate(c.Response().Writer, data, templates.InfoTemplate, templates.RegularType)
}
