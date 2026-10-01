package web_test

import (
	"net/http"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/stretchr/testify/assert"

	"github.com/COMTOP1/AFC-GO/server/internal/web"
)

func redirectEcho() *echo.Echo {
	e := echo.New()
	e.Pre(middleware.RemoveTrailingSlash())
	web.MountRedirects(e)
	return e
}

func TestRedirects(t *testing.T) {
	e := redirectEcho()
	cases := []struct {
		from, to string
		code     int
	}{
		{"/app", "/", http.StatusMovedPermanently},
		{"/app/", "/", http.StatusMovedPermanently},
		{"/app/news/5", "/news/5", http.StatusMovedPermanently},
		{"/app/whatson?period=past", "/whatson?period=past", http.StatusMovedPermanently},
		{"/programmes/3", "/programmes?season=3", http.StatusMovedPermanently},
		{"/programmes/3?q=marlow", "/programmes?q=marlow&season=3", http.StatusMovedPermanently},
		{"/programmes/abc", "/programmes", http.StatusMovedPermanently},
		{"/whatson/period/past", "/whatson?period=past", http.StatusMovedPermanently},
		{"/whatson/period/future", "/whatson?period=future", http.StatusMovedPermanently},
		{"/whatson/period/nonsense", "/whatson", http.StatusMovedPermanently},
		{"/changepassword", "/account", http.StatusMovedPermanently},
		{"/login", "/", http.StatusFound},
		{"/logout", "/", http.StatusFound},
		{"/programmeselect", "/", http.StatusFound},
		{"/whatsonselect", "/", http.StatusFound},
	}
	for _, tc := range cases {
		rec := get(e, http.MethodGet, tc.from)
		assert.Equal(t, tc.code, rec.Code, tc.from)
		assert.Equal(t, tc.to, rec.Header().Get(echo.HeaderLocation), tc.from)
	}
	assert.Equal(t, http.StatusMovedPermanently, get(e, http.MethodHead, "/app/news").Code)
}

func TestAppRedirectNeverLeavesTheSite(t *testing.T) {
	e := redirectEcho()
	for _, from := range []string{"/app//evil.example", "/app///evil.example/x", `/app/\evil.example`, `/app/\/evil.example`,
		"/app/%09/evil.example", "/app/%09%5Cevil.example", "/app/%0D%0A/evil.example", "/app/%2F%2Fevil.example"} {
		rec := get(e, http.MethodGet, from)
		loc := rec.Header().Get(echo.HeaderLocation)
		assert.Equal(t, http.StatusMovedPermanently, rec.Code, from)
		assert.Regexp(t, `^/[^/\\]`, loc, "must be a same-site path: %s -> %s", from, loc)
		assert.NotRegexp(t, `[\x00-\x1f]`, loc, "control characters are escaped: %s -> %q", from, loc)
	}
}

func TestAppRedirectKeepsEncodedPathCharacters(t *testing.T) {
	e := redirectEcho()
	rec := get(e, http.MethodGet, "/app/news/a%3Fb")
	assert.Equal(t, "/news/a%3Fb", rec.Header().Get(echo.HeaderLocation), "an encoded ? stays part of the path")
}

func TestLogoutLinkDoesNotSignOut(t *testing.T) {
	e := redirectEcho()
	rec := get(e, http.MethodGet, "/logout")
	assert.Empty(t, rec.Header().Values(echo.HeaderSetCookie), "GET /logout must not touch cookies")
}
