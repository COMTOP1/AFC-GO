package web

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// CSRF protects unsafe methods. Browsers that send Sec-Fetch-Site:
// same-origin pass without a token; other clients must echo the _csrf
// cookie in X-CSRF-Token. Every failure is a 403.
func CSRF(secureCookie bool) echo.MiddlewareFunc {
	return middleware.CSRFWithConfig(middleware.CSRFConfig{ //nolint:gosec // header/cookie names, not credential values
		TokenLookup:    "header:X-CSRF-Token",
		CookieName:     "_csrf",
		CookiePath:     "/",
		CookieSameSite: http.SameSiteStrictMode,
		CookieSecure:   secureCookie,
		CookieHTTPOnly: false,
		ErrorHandler: func(_ error, _ echo.Context) error {
			return echo.NewHTTPError(http.StatusForbidden, "invalid or missing CSRF token")
		},
	})
}
