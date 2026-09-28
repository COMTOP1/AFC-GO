package views

import (
	"fmt"
	"html"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/COMTOP1/AFC-GO/server/internal/legacy/templates"
	"github.com/COMTOP1/AFC-GO/server/internal/user"
)

func (v *Views) UsersFunc(c echo.Context) error {
	spanCtx, span := tracer.Start(c.Request().Context(), "views.UsersFunc")
	defer span.End()
	c.SetRequest(c.Request().WithContext(spanCtx))
	c1 := v.getSessionData(c)

	usersDB, err := v.user.GetUsers(c.Request().Context())
	if err != nil {
		return fmt.Errorf("failed to get users for users: %w", err)
	}

	teamsDB, err := v.team.GetTeams(c.Request().Context())
	if err != nil {
		return fmt.Errorf("failed to get teams for users: %w", err)
	}

	displayEmail, err := v.setting.GetSetting(c.Request().Context(), "displayEmail")
	if err != nil {
		slog.Info(fmt.Sprintf("failed to get displayEmail for users, error: %+v, continuing", err))
	}

	year, _, _ := time.Now().Date()

	data := struct {
		Year         int
		VisitorCount int
		DisplayEmail string
		Users        []UserTemplate
		Teams        []TeamTemplate
		User         user.User
		Context      *Context
	}{
		Year:         year,
		VisitorCount: v.GetVisitorCount(),
		DisplayEmail: displayEmail.SettingText,
		Users:        DBUsersToTemplateFormat(usersDB),
		Teams:        DBTeamsToTemplateFormat(teamsDB),
		User:         c1.User,
		Context:      c1,
	}

	return v.template.RenderTemplate(c.Request().Context(), c.Response().Writer, data, templates.UsersTemplate, templates.RegularType)
}

func (v *Views) UsersSetDisplayEmailFunc(c echo.Context) error {
	spanCtx, span := tracer.Start(c.Request().Context(), "views.UsersSetDisplayEmailFunc")
	defer span.End()
	c.SetRequest(c.Request().WithContext(spanCtx))
	if c.Request().Method != http.MethodPost {
		return v.invalidMethodUsed(c)
	}
	c1 := v.getSessionData(c)
	var data struct {
		Error string `json:"error"`
	}
	if _, err := v.settingSvc.SetDisplayEmail(c.Request().Context(), c.FormValue("email")); err != nil {
		slog.Info(fmt.Sprintf("failed to set display email, error: %+v", err))
		data.Error = fmt.Sprintf("failed to set display email: %+v", err)
		return c.JSON(http.StatusOK, data)
	}
	v.flash(c, c1, "successfully edited display email")
	return c.JSON(http.StatusOK, data)
}

func (v *Views) UserAddFunc(c echo.Context) error {
	spanCtx, span := tracer.Start(c.Request().Context(), "views.UserAddFunc")
	defer span.End()
	c.SetRequest(c.Request().WithContext(spanCtx))
	if c.Request().Method != http.MethodPost {
		return v.invalidMethodUsed(c)
	}
	c1 := v.getSessionData(c)
	var data struct {
		Error string `json:"error"`
	}
	teamID, err := strconv.Atoi(c.FormValue("userTeam"))
	if err != nil || teamID < 0 {
		teamID = 0
	}
	image, err := legacyUpload(c, "upload")
	if err != nil {
		data.Error = fmt.Sprintf("failed to get file for user add: %+v", err)
		return c.JSON(http.StatusOK, data)
	}
	created, err := v.userSvc.Create(c.Request().Context(), user.CreateInput{
		Name: c.FormValue("name"), Email: c.FormValue("email"), Phone: c.FormValue("phone"),
		Role: c.FormValue("role"), TeamID: teamID,
	}, image)
	if err != nil {
		slog.Info(fmt.Sprintf("failed to add user for user add, error: %+v", err))
		data.Error = fmt.Sprintf("failed to add user for user add: %+v", err)
		return c.JSON(http.StatusOK, data)
	}
	msg := fmt.Sprintf("successfully created user, sent signup email to: \"%s\"", created.User.Email)
	if !created.EmailSent {
		msg = html.UnescapeString(fmt.Sprintf("successfully created user - failed to send email. Please send the username and password to this email: %s, password: %s",
			created.User.Email, created.TempPassword))
	}
	v.flash(c, c1, msg)
	return c.JSON(http.StatusOK, data)
}

func (v *Views) UserEditFunc(c echo.Context) error {
	spanCtx, span := tracer.Start(c.Request().Context(), "views.UserEditFunc")
	defer span.End()
	c.SetRequest(c.Request().WithContext(spanCtx))
	if c.Request().Method != http.MethodPost {
		return v.invalidMethodUsed(c)
	}
	c1 := v.getSessionData(c)
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return fmt.Errorf("failed to parse id for user edit, error: %w", err)
	}
	var data struct {
		Error string `json:"error"`
	}
	remove := c.FormValue("removeUserImage")
	if remove != "" && remove != "Y" {
		data.Error = "failed to parse removeUserImage for user edit: " + remove
		return c.JSON(http.StatusOK, data)
	}
	teamID, err := strconv.Atoi(c.FormValue("userTeam"))
	if err != nil || teamID < 0 {
		teamID = 0
	}
	image, err := legacyUpload(c, "upload")
	if err != nil {
		data.Error = fmt.Sprintf("failed to get file for user edit: %+v", err)
		return c.JSON(http.StatusOK, data)
	}
	name, email, phone, roleCode := c.FormValue("name"), c.FormValue("email"), c.FormValue("phone"), c.FormValue("role")
	updated, err := v.userSvc.Update(c.Request().Context(), userID, user.UpdateInput{
		Name: &name, Email: &email, Phone: &phone, Role: &roleCode, TeamID: &teamID, RemoveImage: remove == "Y",
	}, image)
	if err != nil {
		slog.Info(fmt.Sprintf("failed to edit user for user edit, user id: %d, error: %+v", userID, err))
		data.Error = fmt.Sprintf("failed to edit user for user edit: %+v", err)
		return c.JSON(http.StatusOK, data)
	}
	v.flash(c, c1, fmt.Sprintf("successfully edited \"%s\"", updated.Name))
	return c.JSON(http.StatusOK, data)
}

func (v *Views) UserDeleteFunc(c echo.Context) error {
	spanCtx, span := tracer.Start(c.Request().Context(), "views.UserDeleteFunc")
	defer span.End()
	c.SetRequest(c.Request().WithContext(spanCtx))
	if c.Request().Method != http.MethodPost {
		return v.invalidMethodUsed(c)
	}
	c1 := v.getSessionData(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return fmt.Errorf("failed to get id for user delete, error: %w", err)
	}
	deleted, err := v.userSvc.Delete(c.Request().Context(), id)
	if err != nil {
		return fmt.Errorf("failed to delete user for user delete, user id: %d, error: %w", id, err)
	}
	v.flash(c, c1, fmt.Sprintf("successfully deleted \"%s\"", deleted.Name))
	return c.Redirect(http.StatusFound, "/users")
}
