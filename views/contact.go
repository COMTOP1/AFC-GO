package views

import (
	"fmt"
	"github.com/COMTOP1/AFC-GO/role"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/COMTOP1/AFC-GO/templates"
	"github.com/COMTOP1/AFC-GO/user"
)

type (
	ContactUserTemplate struct {
		ID    int
		Name  string
		Email string
		Role  string
	}
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

func DBUsersContactToTemplateFormat(dbUsers []user.User) ([]ContactUserTemplate, error) {
	tplUsers := make([]ContactUserTemplate, 0, len(dbUsers))
	for _, u := range dbUsers {
		var contactUser ContactUserTemplate
		contactUser.ID = u.ID
		contactUser.Name = u.Name
		contactUser.Email = u.Email
		temp, err := role.GetRole(u.TempRole)
		if err != nil {
			return nil, fmt.Errorf("failed to parse role for contactTemplate: %w", err)
		}
		contactUser.Role = temp.String()
		tplUsers = append(tplUsers, contactUser)
	}
	return tplUsers, nil
}
