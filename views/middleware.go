package views

import (
	"fmt"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/COMTOP1/AFC-GO/role"
)

// RequiresLogin is a middleware which will be used for each
// httpHandler to check if there is any active session
func (v *Views) RequiresLogin(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		session, err := v.cookie.Get(c.Request(), v.conf.SessionCookieName)
		if err != nil {
			log.Printf("failed to get session: %+v", err)
			session, err = v.cookie.New(c.Request(), v.conf.SessionCookieName)
			if err != nil {
				panic(fmt.Errorf("failed to make new session: %w", err))
			}
			err = session.Save(c.Request(), c.Response())
			if err != nil {
				log.Printf("failed to save session for logout: %+v", err)
			}
			return c.Redirect(http.StatusFound, "/")
		}
		c1 := v.getSessionData(c)
		if !c1.User.Authenticated {
			return c.Redirect(http.StatusFound, "/")
		}
		c1.User, err = v.user.GetUser(c.Request().Context(), c1.User)
		if err != nil {
			log.Printf("failed to get user from db: %+v", err)
			return c.Redirect(http.StatusFound, "/")
		}
		c1.User.Authenticated = true
		session.Values["user"] = c1.User
		err = session.Save(c.Request(), c.Response())
		if err != nil {
			log.Printf("failed to save session for logout: %+v", err)
		}
		return next(c)
	}
}

func (v *Views) RequireNotManager(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c1 := v.getSessionData(c)
		if c1 == nil {
			return fmt.Errorf("failed to get session data")
		}

		if c1.User.ID > 0 && c1.User.Role != role.Manager {
			return next(c)
		}

		return echo.NewHTTPError(http.StatusForbidden, fmt.Errorf("you are not authorised for accessing this"))
	}
}

func (v *Views) RequireClubSecretaryHigher(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c1 := v.getSessionData(c)
		if c1 == nil {
			return fmt.Errorf("failed to get session data")
		}

		if c1.User.ID <= 0 {
			return echo.NewHTTPError(http.StatusForbidden, fmt.Errorf("you are not authorised for accessing this"))
		}

		if c1.User.Role == role.ClubSecretary || c1.User.Role == role.Chairperson || c1.User.Role == role.Webmaster {
			return next(c)
		}

		return echo.NewHTTPError(http.StatusForbidden, fmt.Errorf("you are not authorised for accessing this"))
	}
}
