package views

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/patrickmn/go-cache"
	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/templates"
	"github.com/COMTOP1/AFC-GO/user"
)

func (v *Views) ResetURLFunc(c echo.Context) error {
	c1 := v.getSessionData(c)

	url := c.Param("url")

	id, found := v.cache.Get(url)
	if !found {
		return fmt.Errorf("failed to get url for reset")
	}

	originalUser, err := v.user.GetUser(c.Request().Context(), user.User{ID: id.(int)})
	if err != nil {
		v.cache.Delete(url)
		return fmt.Errorf("url is invalid, failed to get user : %w", err)
	}

	switch c.Request().Method {
	case "GET":
		return v.template.RenderTemplate(c.Response(), c1, templates.ResetTemplate, templates.NoNavType)
	case "POST":
		password := c.FormValue("password")
		if password != c.FormValue("confirmpassword") {
			return v.template.RenderTemplate(c.Response(), nil, templates.ResetTemplate, templates.NoNavType)
		}

		originalUser.Password = null.StringFrom(password)

		errString := minRequirementsMet(password)
		if len(errString) > 0 {
			data := struct{ Error string }{Error: errString}
			return v.template.RenderTemplate(c.Response().Writer, data, templates.ResetTemplate, templates.NoNavType)
		}

		err = v.user.EditUserPassword(c.Request().Context(), originalUser, v.conf.Security.Iterations, v.conf.Security.KeyLength)
		if err != nil {
			log.Printf("failed to reset password: %+v", err)
		}
		v.cache.Delete(url)
		log.Printf("updated user: %s", originalUser.Email)
		return c.Redirect(http.StatusFound, "/")
	}
	return nil
}

func (v *Views) ResetUserPasswordFunc(c echo.Context) error {
	c1 := v.getSessionData(c)

	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return fmt.Errorf("failed to parse userid for reset: %w", err)
	}

	userFromDB, err := v.user.GetUser(c.Request().Context(), user.User{ID: userID})
	if err != nil {
		return fmt.Errorf("failed to get user for reset: %w", err)
	}

	userFromDB.ResetPassword = true

	_, err = v.user.EditUser(c.Request().Context(), userFromDB, userFromDB.Email)
	if err != nil {
		return fmt.Errorf("failed to update user for reset: %w", err)
	}

	url := uuid.NewString()
	v.cache.Set(url, userFromDB.ID, cache.DefaultExpiration)

	var message struct {
		Message string `json:"message"`
		Error   error  `json:"error"`
	}

	message.Message = fmt.Sprintf("Please forward the link to this email: %s, reset link: https://afcaldermaston.co.uk/reset/%s", userFromDB.Email, url)

	log.Printf("reset for %s (%s) requested by %s (%s)", userFromDB.Email, userFromDB.Name, c1.User.Email, c1.User.Name)
	var status int
	return c.JSON(status, message)
}
