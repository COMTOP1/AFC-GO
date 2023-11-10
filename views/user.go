package views

import (
	"fmt"
	"time"

	"github.com/labstack/echo/v4"

	role2 "github.com/COMTOP1/AFC-GO/role"
	"github.com/COMTOP1/AFC-GO/templates"
	"github.com/COMTOP1/AFC-GO/user"
)

type UserTemplate struct {
	ID     int
	Name   string
	Email  string
	Phone  string
	TeamID int
	Role   string
}

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

func DBUsersToTemplateFormat(u1 []user.User) []UserTemplate {
	usersTemplate := make([]UserTemplate, 0, len(u1))
	for _, userDB := range u1 {
		var u2 UserTemplate
		u2.ID = userDB.ID
		u2.Name = userDB.Name
		u2.Email = userDB.Email
		u2.Phone = userDB.Phone
		if userDB.TeamID.Valid {
			u2.TeamID = int(userDB.TeamID.Int64)
		}
		role, err := role2.GetRole(userDB.TempRole)
		u2.Role = role.String()
		if err != nil {
			u2.Role = fmt.Sprintf("failed to get role for users: %+v", err)
		}
		usersTemplate = append(usersTemplate, u2)
	}
	return usersTemplate
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
