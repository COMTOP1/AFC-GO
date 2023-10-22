package views

import (
	"fmt"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"

	role1 "github.com/COMTOP1/AFC-GO/infrastructure/role"
	"github.com/COMTOP1/AFC-GO/role"
	"github.com/COMTOP1/AFC-GO/user"
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
		userFromDB, err := v.user.GetUser(c.Request().Context(), c1.User)
		if err != nil {
			log.Printf("failed to get user from db: %+v", err)
			return c.Redirect(http.StatusFound, "/")
		}
		if userFromDB.DeletedBy.Valid || !c1.User.Enabled {
			session.Values["user"] = &user.User{}
			session.Options.MaxAge = -1
			err = session.Save(c.Request(), c.Response())
			if err != nil {
				return fmt.Errorf("failed to save session for requiresLogin: %w", err)
			}
			return c.Redirect(http.StatusFound, "/")
		}
		return next(c)
	}
}

// RequiresLoginJSON is a middleware which is used for each
// httpHandler to check if there is any active session and returns json if not
func (v *Views) RequiresLoginJSON(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		session, err := v.cookie.Get(c.Request(), v.conf.SessionCookieName)
		if err != nil {
			log.Printf("failed to get session for requiresLoginJSON: %+v", err)
			data := struct {
				Error error `json:"error"`
			}{
				Error: err,
			}
			return c.JSON(http.StatusInternalServerError, data)
		}
		c1 := v.getSessionData(c)
		if !c1.User.Authenticated {
			data := struct {
				Error string `json:"error"`
			}{
				Error: "user not logged in",
			}
			return c.JSON(http.StatusUnauthorized, data)
		}
		userFromDB, err := v.user.GetUser(c.Request().Context(), c1.User)
		if err != nil {
			log.Printf("failed to get user for requiresLoginJSON: %+v", err)
			data := struct {
				Error error `json:"error"`
			}{
				Error: fmt.Errorf("failed to get user for requiresLoginJSON: %w", err),
			}
			return c.JSON(http.StatusInternalServerError, data)
		}
		if userFromDB.DeletedBy.Valid || !c1.User.Enabled {
			session.Values["user"] = &user.User{}
			session.Options.MaxAge = -1
			err = session.Save(c.Request(), c.Response())
			if err != nil {
				log.Printf("failed to save session for requiresLoginJSON: %+v", err)
				data := struct {
					Error error `json:"error"`
				}{
					Error: fmt.Errorf("failed to save session for requiresLoginJSON: %w", err),
				}
				return c.JSON(http.StatusInternalServerError, data)
			}
			data := struct {
				Error string `json:"error"`
			}{
				Error: "user deleted or not enabled",
			}
			return c.JSON(http.StatusUnauthorized, data)
		}
		return next(c)
	}
}

func (v *Views) RequirePermission(r role.Role) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c1 := v.getSessionData(c)
			if c1 == nil {
				return fmt.Errorf("failed to get session data")
			}
			perms, err := v.user.GetPermissionsForUser(c.Request().Context(), c1.User)
			if err != nil {
				return fmt.Errorf("failed to get permissions for requirePermission: %w", err)
			}

			acceptedPerms := role1.SufficientPermissionsFor(r)

			for _, perm := range perms {
				if acceptedPerms[perm.Name] {
					return next(c)
				}
			}

			return echo.NewHTTPError(http.StatusForbidden, fmt.Errorf("you are not authorised for accessing this"))
		}
	}
}
