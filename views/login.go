package views

import (
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	emailParser "github.com/mcnijman/go-emailaddress"
	"github.com/patrickmn/go-cache"
	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/templates"
	"github.com/COMTOP1/AFC-GO/user"
)

// LoginFunc implements the login functionality, will
// add a cookie to the cookie store for managing authentication
func (v *Views) LoginFunc(c echo.Context) error {
	session, _ := v.cookie.Get(c.Request(), v.conf.SessionCookieName)
	// We're ignoring the error here since sometimes the cookies keys change, and then we
	// can overwrite it instead, it does need to stay as it is written to here

	switch c.Request().Method {
	case "GET":
		// Data for our HTML template
		context := v.getSessionData(c)

		// Check if there is a callback request
		callbackURL, err := url.Parse(c.QueryParam("callback"))
		if err == nil && /*strings.HasSuffix(callbackURL.Host, v.conf.BaseDomainName) &&*/ callbackURL.String() != "" {
			context.Callback = callbackURL.String()
		}
		// Check if authenticated
		if context.User.Authenticated {
			return c.Redirect(http.StatusFound, context.Callback)
		}

		return v.template.RenderTemplate(c.Response(), context, templates.LoginTemplate, templates.NoNavType)
	case "POST":
		// Parsing form to struct
		email := c.FormValue("email")
		password := c.FormValue("password")
		var u user.User

		u.Email = email
		u.Password = null.StringFrom(password)

		callback := "/internal"
		callbackURL, err := url.Parse(c.QueryParam("callback"))
		if err == nil /*&& strings.HasSuffix(callbackURL.Host, v.conf.BaseDomainName)*/ && callbackURL.String() != "" {
			callback = callbackURL.String()
		}
		// Authentication
		u, resetPw, err := v.user.VerifyUser(c.Request().Context(), u)
		if err != nil {
			address, _ := emailParser.Parse(username)
			if address != nil {
				u.LDAPUsername = null.StringFrom(address.LocalPart)
				_, err = v.user.GetUser(c.Request().Context(), u)
			}
			log.Printf("failed login for \"%s\": %v", u.Username, err)
			err = session.Save(c.Request(), c.Response())
			if err != nil {
				return fmt.Errorf("failed to save session for login: %w", err)
			}

			if resetPw {
				ctx := v.getSessionData(c)
				ctx.Callback = callback
				ctx.Message = "Password reset required"
				ctx.MsgType = "is-danger"

				err = v.setMessagesInSession(c, ctx)
				if err != nil {
					return fmt.Errorf("failed to set message for login: %w", err)
				}

				url1 := uuid.NewString()
				v.cache.Set(url1, u.UserID, cache.DefaultExpiration)

				return c.Redirect(http.StatusFound, fmt.Sprintf("https://%s/reset/%s", v.conf.DomainName, url1))
			}
			ctx := v.getSessionData(c)
			ctx.Callback = callback
			ctx.Message = "Invalid username or password"
			ctx.MsgType = "is-danger"
			err = v.setMessagesInSession(c, ctx)
			if err != nil {
				return fmt.Errorf("failed to set message for login: %w", err)
			}

			return c.Redirect(http.StatusFound, "/login")
		}
		prevLogin := u.LastLogin
		// Update last logged in
		err = v.user.SetUserLoggedIn(c.Request().Context(), u)
		if err != nil {
			return fmt.Errorf("failed to set user logged in for login: %w", err)
		}
		u.Authenticated = true
		// This is a bit of a cheat, just so we can have the last login displayed for internal
		u.LastLogin = prevLogin

		err = v.clearMessagesInSession(c)
		if err != nil {
			return fmt.Errorf("failed to clear message: %w", err)
		}

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

		log.Printf("user \"%s\" is authenticated", u.Username)
		return c.Redirect(http.StatusFound, callback)
	}
	return fmt.Errorf("failed to parse method")
}
