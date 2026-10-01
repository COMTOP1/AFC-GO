package web

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
)

// MountRedirects keeps classic-site and /app URLs working now that the web
// client is served at the root. Only GET and HEAD are redirected. /logout
// deliberately doesn't sign out: a GET link would let any site sign people
// out. Signing out is POST /api/v1/auth/logout.
func MountRedirects(e *echo.Echo) {
	methods := []string{http.MethodGet, http.MethodHead}
	e.Match(methods, "/app", fromApp)
	e.Match(methods, "/app/*", fromApp)
	e.Match(methods, "/programmes/:id", programmeSeason)
	e.Match(methods, "/whatson/period/:period", whatsOnPeriod)
	e.Match(methods, "/changepassword", redirectTo("/account", http.StatusMovedPermanently))
	for _, p := range []string{"/login", "/logout", "/programmeselect", "/whatsonselect"} {
		e.Match(methods, p, redirectTo("/", http.StatusFound))
	}
}

// fromApp strips the old /app prefix. Leading slashes and backslashes are
// collapsed so the target can never be a protocol-relative URL (//host).
func fromApp(c echo.Context) error {
	rest := strings.TrimPrefix(c.Request().URL.Path, "/app")
	rest = "/" + strings.TrimLeft(rest, `/\`)
	return redirect(c, http.StatusMovedPermanently, rest, nil)
}

func programmeSeason(c echo.Context) error {
	extra := url.Values{}
	if id, err := strconv.Atoi(c.Param("id")); err == nil && id > 0 {
		extra.Set("season", strconv.Itoa(id))
	}
	return redirect(c, http.StatusMovedPermanently, "/programmes", extra)
}

func whatsOnPeriod(c echo.Context) error {
	extra := url.Values{}
	switch p := c.Param("period"); p {
	case "future", "past", "all":
		extra.Set("period", p)
	}
	return redirect(c, http.StatusMovedPermanently, "/whatson", extra)
}

func redirectTo(target string, code int) echo.HandlerFunc {
	return func(c echo.Context) error { return redirect(c, code, target, nil) }
}

// redirect sends code to target, keeping the request's query and adding extra.
func redirect(c echo.Context, code int, target string, extra url.Values) error {
	q := c.Request().URL.Query()
	for k, vs := range extra {
		q[k] = vs
	}
	if enc := q.Encode(); enc != "" {
		target += "?" + enc
	}
	return c.Redirect(code, target)
}
