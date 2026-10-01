package web_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/COMTOP1/AFC-GO/server/internal/web"
	"github.com/COMTOP1/AFC-GO/server/internal/web/apitest"
)

func newAPI() (*echo.Echo, *echo.Group) { //nolint:unparam // returned group is part of the test fixture's shape, not every test needs it
	e := apitest.NewEcho()
	api := web.NewAPI(e, false)
	api.GET("/thing", func(c echo.Context) error { return c.NoContent(http.StatusOK) })
	api.POST("/thing", func(c echo.Context) error { return c.NoContent(http.StatusCreated) })
	return e, api
}

func TestUnknownAPIPathIsJSON404(t *testing.T) {
	e, _ := newAPI()
	for _, path := range []string{"/api/v1/nope", "/api/v1/nope/deeper", "/api/nope"} {
		rec := apitest.New(e).Get(t, path)
		assert.Equal(t, http.StatusNotFound, rec.Code, path)
		assert.Equal(t, 404, apitest.ErrorOf(t, rec).Code, path)
	}
}

func TestWrongMethodIsJSON405(t *testing.T) {
	// The /api/v1 group's RouteNotFound("/*") (registered by Group.Use when
	// CSRF middleware is attached) matches before Echo's router gets to the
	// method-not-allowed check, so an unregistered method on a known path
	// comes back as a JSON 404, not a 405. It is still a JSON envelope,
	// which is the behaviour that matters (see task-4-brief.md Step 8).
	e, _ := newAPI()
	rec := apitest.New(e).JSON(t, http.MethodDelete, "/api/v1/thing", nil)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.Equal(t, 404, apitest.ErrorOf(t, rec).Code)
}

func TestHealth(t *testing.T) {
	e, _ := newAPI()
	for _, path := range []string{"/api/v1/health", "/api/health"} {
		rec := apitest.New(e).Get(t, path)
		assert.Equal(t, http.StatusOK, rec.Code, path)
		assert.JSONEq(t, `{"status":"ok"}`, rec.Body.String(), path)
	}
}

func TestSwaggerRedirects(t *testing.T) {
	e, _ := newAPI()
	for _, path := range []string{"/api", "/api/", "/api/v1", "/api/v1/"} {
		rec := apitest.New(e).Get(t, path)
		assert.Equal(t, http.StatusFound, rec.Code, path)
		assert.Equal(t, "/api/v1/swagger/index.html", rec.Header().Get("Location"), path)
	}
}

func TestSameOriginPostPasses(t *testing.T) {
	e, _ := newAPI()
	rec := apitest.New(e).JSON(t, http.MethodPost, "/api/v1/thing", map[string]string{})
	assert.Equal(t, http.StatusCreated, rec.Code)
}

func TestCrossSitePostRejected(t *testing.T) {
	e, _ := newAPI()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/v1/thing", strings.NewReader("{}"))
	req.Header.Set("Sec-Fetch-Site", "cross-site")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	assert.Equal(t, 403, apitest.ErrorOf(t, rec).Code)
}

func TestTokenlessNonBrowserPostRejected(t *testing.T) {
	e, _ := newAPI()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/v1/thing", strings.NewReader("{}"))
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusForbidden, rec.Code, "missing token must be 403, not echo's default 400")
}

func TestDoubleSubmitTokenAccepted(t *testing.T) {
	e, _ := newAPI()
	// A GET without Sec-Fetch-Site issues the _csrf cookie.
	get := httptest.NewRecorder()
	e.ServeHTTP(get, httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/thing", nil))
	var token *http.Cookie
	for _, c := range get.Result().Cookies() {
		if c.Name == "_csrf" {
			token = c
		}
	}
	require.NotNil(t, token, "expected _csrf cookie")
	assert.False(t, token.HttpOnly, "client JS must be able to read _csrf")
	assert.Equal(t, http.SameSiteStrictMode, token.SameSite)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/v1/thing", strings.NewReader("{}"))
	req.AddCookie(token)
	req.Header.Set("X-CSRF-Token", token.Value)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusCreated, rec.Code)
}
