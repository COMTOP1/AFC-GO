package views

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
)

// LoginFunc implements the login functionality, will
// add a cookie to the cookie store for managing authentication
func (v *Views) LoginFunc(c echo.Context) error {
	spanCtx, span := tracer.Start(c.Request().Context(), "views.LoginFunc")
	defer span.End()
	c.SetRequest(c.Request().WithContext(spanCtx))
	session, _ := v.cookie.Get(c.Request(), v.conf.SessionCookieName)
	// We're ignoring the error here since sometimes the cookie keys change, and then we
	// can overwrite it instead, it does need to stay as it is written to here

	if c.Request().Method == http.MethodPost {
		res, err := v.authSvc.Login(c.Request().Context(), c.FormValue("email"), c.FormValue("password"))
		if err != nil {
			if saveErr := session.Save(c.Request(), c.Response()); saveErr != nil {
				return fmt.Errorf("failed to save session for login: %w", saveErr)
			}
			ctx := v.getSessionData(c)
			ctx.Message = "Invalid email or password"
			ctx.MsgType = "is-danger"
			if err = v.setMessagesInSession(c, ctx); err != nil {
				return fmt.Errorf("failed to set message for login: %w", err)
			}
			return c.JSON(http.StatusOK, struct {
				Error         string `json:"error"`
				ResetPassword bool   `json:"resetPassword"`
			}{Error: "Invalid email or password"})
		}
		if res.ResetRequired {
			ctx := v.getSessionData(c)
			ctx.Message = "Password reset required"
			ctx.MsgType = "is-danger"
			if err = v.setMessagesInSession(c, ctx); err != nil {
				return fmt.Errorf("failed to set message for login: %w", err)
			}
			return c.JSON(http.StatusOK, struct {
				Error         string `json:"error"`
				ResetPassword bool   `json:"resetPassword"`
				URL           string `json:"url"`
			}{ResetPassword: true, URL: res.ResetURL})
		}
		u := res.User
		u.Authenticated = true

		err = v.clearMessagesInSession(c)
		if err != nil {
			return fmt.Errorf("failed to clear message: %w", err)
		}

		session.Values["user"] = u

		if c.FormValue("remember") != "on" {
			session.Options.MaxAge = 86400 * 31
		}

		err = session.Save(c.Request(), c.Response())
		if err != nil {
			return fmt.Errorf("failed to save user session for login: %w", err)
		}

		slog.Info(fmt.Sprintf("user \"%s\" is authenticated", u.Email))
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
