package views

import (
	"fmt"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/COMTOP1/AFC-GO/templates"
	"github.com/COMTOP1/AFC-GO/user"
)

func (v *Views) UsersFunc(c echo.Context) error {
	c1 := v.getSessionData(c)

	var u1 []user.User
	var err error

	u1, err = v.user.GetUsers(c.Request().Context())
	if err != nil {
		return fmt.Errorf("failed to get users for users: %w", err)
	}

	year, _, _ := time.Now().Date()

	data := struct {
		Year  int
		Users []UserTemplate
		User  user.User
	}{
		Year:  year,
		Users: DBUsersToTemplateFormat(u1),
		User:  c1.User,
	}

	return v.template.RenderTemplate(c.Response().Writer, data, templates.UsersTemplate, templates.RegularType)
}

func (v *Views) UserAddFunc(c echo.Context) error {
	_ = c
	return fmt.Errorf("not implemented yet")
}

func (v *Views) UserEditFunc(c echo.Context) error {
	_ = c
	return fmt.Errorf("not implemented yet")
}

func (v *Views) UserDeleteFunc(c echo.Context) error {
	_ = c
	return fmt.Errorf("not implemented yet")
}
