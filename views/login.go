package views

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/patrickmn/go-cache"
	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/role"
	"github.com/COMTOP1/AFC-GO/user"
)

// LoginFunc implements the login functionality, will
// add a cookie to the cookie store for managing authentication
func (v *Views) LoginFunc(c echo.Context) error {
	session, _ := v.cookie.Get(c.Request(), v.conf.SessionCookieName)
	// We're ignoring the error here since sometimes the cookie keys change, and then we
	// can overwrite it instead, it does need to stay as it is written to here

	if c.Request().Method == http.MethodPost {
		// Parsing form to struct
		email := c.FormValue("email")
		password := c.FormValue("password")
		var u user.User

		u.Email = email
		u.Password = null.StringFrom(password)

		var message struct {
			Error error `json:"error"`
		}

		_ = message

		// Authentication
		u, resetPw, err := v.user.VerifyUser(c.Request().Context(), u, v.conf.Security.Iterations, v.conf.Security.ScryptWorkFactor, v.conf.Security.ScryptBlockSize, v.conf.Security.ScryptParallelismFactor, v.conf.Security.KeyLength)
		if err != nil {
			log.Printf("failed login for \"%s\": %v", u.Email, err)
			err = session.Save(c.Request(), c.Response())
			if err != nil {
				return fmt.Errorf("failed to save session for login: %w", err)
			}

			if resetPw {
				ctx := v.getSessionData(c)
				ctx.Message = "Password reset required"
				ctx.MsgType = "is-danger"

				err = v.setMessagesInSession(c, ctx)
				if err != nil {
					return fmt.Errorf("failed to set message for login: %w", err)
				}

				url1 := uuid.NewString()
				v.cache.Set(url1, u.ID, cache.DefaultExpiration)

				data := struct {
					Error         string `json:"error"`
					ResetPassword bool   `json:"resetPassword"`
					URL           string `json:"url"`
				}{
					ResetPassword: true,
					URL:           "/reset/" + url1,
				}
				return c.JSON(http.StatusOK, data)
			}
			ctx := v.getSessionData(c)
			ctx.Message = "Invalid email or password"
			ctx.MsgType = "is-danger"
			err = v.setMessagesInSession(c, ctx)
			if err != nil {
				return fmt.Errorf("failed to set message for login: %w", err)
			}
			data := struct {
				Error         string `json:"error"`
				ResetPassword bool   `json:"resetPassword"`
			}{
				Error:         "Invalid email or password",
				ResetPassword: false,
			}
			return c.JSON(http.StatusOK, data)
		}
		u.Authenticated = true

		err = v.clearMessagesInSession(c)
		if err != nil {
			return fmt.Errorf("failed to clear message: %w", err)
		}

		u.Role, err = role.GetRole(u.TempRole)
		if err != nil {
			return fmt.Errorf("failed to get role for login: %w", err)
		}

		u.TempRole = ""

		session.Values["user"] = u

		c.SetCookie(&http.Cookie{
			Name:     "test-pass",
			Value:    "hello",
			MaxAge:   60,
			Secure:   false,
			HttpOnly: false,
		})

		if c.FormValue("remember") != "on" {
			session.Options.MaxAge = 86400 * 31
		}

		err = session.Save(c.Request(), c.Response())
		if err != nil {
			return fmt.Errorf("failed to save user session for login: %w", err)
		}

		log.Printf("user \"%s\" is authenticated", u.Email)
		data := struct {
			Error         string `json:"error"`
			ResetPassword bool   `json:"resetPassword"`
		}{
			Error:         "",
			ResetPassword: false,
		}
		return c.JSON(http.StatusOK, data)
	}
	return errors.New("failed to parse method")
}
