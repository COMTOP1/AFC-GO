package auth

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/COMTOP1/AFC-GO/server/internal/role"
	"github.com/COMTOP1/AFC-GO/server/internal/user"
	"github.com/COMTOP1/AFC-GO/server/internal/web"
)

const contextKey = "auth.user"

// Current returns the user loaded by a guard earlier in the chain.
func Current(c echo.Context) (user.User, bool) {
	u, ok := c.Get(contextKey).(user.User)
	return u, ok
}

// RequireLogin rejects anonymous requests with 401. It reloads the user from
// the database so role changes and deletions take effect immediately.
func (s *Sessions) RequireLogin(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if _, err := s.load(c); err != nil {
			return err
		}
		return next(c)
	}
}

// RequireRole rejects anonymous requests with 401 and users whose role fails
// allowed with 403.
func (s *Sessions) RequireRole(allowed func(role.Role) bool) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			u, err := s.load(c)
			if err != nil {
				return err
			}
			if !allowed(u.Role) {
				return echo.NewHTTPError(http.StatusForbidden, "you are not authorised for accessing this")
			}
			return next(c)
		}
	}
}

// Guards returns the guard set handlers attach to routes.
func (s *Sessions) Guards() web.Guards {
	return web.Guards{
		Login:               s.RequireLogin,
		Editor:              s.RequireRole(role.Role.CanEdit),
		NotManager:          s.RequireRole(role.Role.CanManageGallery),
		ClubSecretaryHigher: s.RequireRole(role.Role.IsClubSecretaryHigher),
		Identify:            s.Identify,
	}
}

func (s *Sessions) load(c echo.Context) (user.User, error) {
	if u, ok := Current(c); ok {
		return u, nil
	}
	u, ok := s.User(c.Request())
	if !ok {
		return user.User{}, echo.NewHTTPError(http.StatusUnauthorized, "login required")
	}
	fresh, err := s.users.GetUser(c.Request().Context(), u)
	if err != nil || fresh.ID != u.ID {
		return user.User{}, echo.NewHTTPError(http.StatusUnauthorized, "login required").SetInternal(err)
	}
	fresh.Authenticated = true
	c.Set(contextKey, fresh)
	c.Set(web.LoggedInKey, true)
	return fresh, nil
}

// Identify loads the user when a valid session is present, but lets
// anonymous (or stale-session) requests through.
func (s *Sessions) Identify(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if _, ok := s.User(c.Request()); ok {
			_, _ = s.load(c) //nolint:errcheck // a stale session is simply anonymous here
		}
		return next(c)
	}
}
