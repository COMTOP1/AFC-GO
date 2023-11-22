package views

import (
	"fmt"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/COMTOP1/AFC-GO/templates"
	"github.com/COMTOP1/AFC-GO/user"
)

func (v *Views) ContactFunc(c echo.Context) error {
	c1 := v.getSessionData(c)

	dbContactUsers, err := v.user.GetUsersContact(c.Request().Context())
	if err != nil {
		return fmt.Errorf("failed to get contact users: %w", err)
	}

	contactTmpl, err := DBUsersContactToTemplateFormat(dbContactUsers)
	if err != nil {
		return fmt.Errorf("failed to format contact users: %w", err)
	}

	year, _, _ := time.Now().Date()

	data := struct {
		Year         int
		ContactUsers []ContactUserTemplate
		User         user.User
	}{
		Year:         year,
		ContactUsers: contactTmpl,
		User:         c1.User,
	}

	return v.template.RenderTemplate(c.Response().Writer, data, templates.ContactTemplate, templates.RegularType)
}
