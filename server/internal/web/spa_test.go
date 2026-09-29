package web_test

import (
	"context"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/COMTOP1/AFC-GO/server/internal/web"
)

const indexHTML = `<!doctype html><html><body><div id="root"></div><script type="module" src="/app/assets/index-abc123.js"></script></body></html>`

func builtFS() fstest.MapFS {
	return fstest.MapFS{
		"index.html":             {Data: []byte(indexHTML)},
		"assets/index-abc123.js": {Data: []byte("console.log('app')")},
		"assets/style-9f8e.css":  {Data: []byte("body{}")},
		"favicon.svg":            {Data: []byte("<svg/>")},
		".keep":                  {Data: nil},
	}
}

func spaEcho(files fs.FS, proxy *url.URL) *echo.Echo {
	e := echo.New()
	e.Pre(middleware.RemoveTrailingSlash())
	e.RouteNotFound("/*", func(c echo.Context) error { return c.HTML(http.StatusNotFound, "legacy 404") })
	web.MountSPA(e, files, proxy)
	return e
}

func get(e *echo.Echo, method, path string) *httptest.ResponseRecorder {
	req := httptest.NewRequestWithContext(context.Background(), method, path, nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

func TestSPAIndexFallback(t *testing.T) {
	e := spaEcho(builtFS(), nil)
	for _, path := range []string{"/app", "/app/", "/app/news", "/app/news/5", "/app/teams/2/players"} {
		rec := get(e, http.MethodGet, path)
		assert.Equal(t, http.StatusOK, rec.Code, path)
		assert.Equal(t, indexHTML, rec.Body.String(), path)
		assert.Contains(t, rec.Header().Get(echo.HeaderContentType), "text/html", path)
	}
}

func TestSPAAssets(t *testing.T) {
	e := spaEcho(builtFS(), nil)

	rec := get(e, http.MethodGet, "/app/assets/index-abc123.js")
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "console.log('app')", rec.Body.String())
	assert.Contains(t, rec.Header().Get(echo.HeaderContentType), "javascript")

	rec = get(e, http.MethodGet, "/app/assets/style-9f8e.css")
	assert.Contains(t, rec.Header().Get(echo.HeaderContentType), "text/css")

	for _, missing := range []string{"/app/assets/index-old999.js", "/app/assets", "/app/favicon.png"} {
		rec = get(e, http.MethodGet, missing)
		assert.Equal(t, http.StatusNotFound, rec.Code, missing)
		assert.NotContains(t, rec.Body.String(), `<div id="root">`, "must never fall back to index: %s", missing)
	}

	rec = get(e, http.MethodGet, "/app/favicon.svg")
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "<svg/>", rec.Body.String())
}

func TestSPACacheHeaders(t *testing.T) {
	e := spaEcho(builtFS(), nil)
	assert.Equal(t, "no-cache", get(e, http.MethodGet, "/app/news/5").Header().Get("Cache-Control"))
	assert.Equal(t, "public, max-age=31536000, immutable",
		get(e, http.MethodGet, "/app/assets/index-abc123.js").Header().Get("Cache-Control"))
	assert.Equal(t, "no-cache", get(e, http.MethodGet, "/app/favicon.svg").Header().Get("Cache-Control"))
}

func TestSPANotBuilt(t *testing.T) {
	for name, files := range map[string]fs.FS{
		"only .keep": fstest.MapFS{".keep": {Data: nil}},
		"nil fs":     nil, // a true nil interface, as app.Config.UI is in tests
	} {
		t.Run(name, func(t *testing.T) {
			e := spaEcho(files, nil)
			for _, path := range []string{"/app", "/app/news/5", "/app/assets/index-abc123.js"} {
				rec := get(e, http.MethodGet, path)
				assert.Equal(t, http.StatusServiceUnavailable, rec.Code, path)
				assert.Contains(t, rec.Body.String(), "The web client has not been built", path)
				assert.Equal(t, "no-cache", rec.Header().Get("Cache-Control"), path)
			}
		})
	}
}

func TestSPAMethods(t *testing.T) {
	e := spaEcho(builtFS(), nil)
	assert.Equal(t, http.StatusMethodNotAllowed, get(e, http.MethodPost, "/app/news").Code)
	assert.Equal(t, http.StatusMethodNotAllowed, get(e, http.MethodDelete, "/app").Code)
	head := get(e, http.MethodHead, "/app/news")
	assert.Equal(t, http.StatusOK, head.Code)
}

func TestSPARejectsTraversal(t *testing.T) {
	e := spaEcho(builtFS(), nil)
	for _, path := range []string{"/app/../go.mod", "/app/assets/../../go.mod", "/app/%2e%2e/go.mod", "/app/assets/%2e%2e/index.html"} {
		rec := get(e, http.MethodGet, path)
		assert.NotContains(t, rec.Body.String(), "module github.com", path)
		assert.NotEqual(t, "console.log('app')", rec.Body.String(), path)
	}
}

func TestSPADoesNotShadowOtherRoutes(t *testing.T) {
	e := spaEcho(builtFS(), nil)
	e.GET("/news", func(c echo.Context) error { return c.String(http.StatusOK, "legacy news") })
	assert.Equal(t, "legacy news", get(e, http.MethodGet, "/news").Body.String())
	assert.Equal(t, "legacy 404", get(e, http.MethodGet, "/application").Body.String())
	assert.Equal(t, "legacy 404", get(e, http.MethodGet, "/apps/x").Body.String())
}

func TestSPAProxy(t *testing.T) {
	var gotPath, gotMethod, gotHost, gotBody string
	vite := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath, gotMethod, gotHost = r.URL.Path, r.Method, r.Host
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		_, _ = w.Write([]byte("from vite"))
	}))
	defer vite.Close()
	target, err := url.Parse(vite.URL)
	require.NoError(t, err)

	e := spaEcho(builtFS(), target)
	rec := get(e, http.MethodGet, "/app/src/main.tsx")
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "from vite", rec.Body.String())
	assert.Equal(t, "/app/src/main.tsx", gotPath)
	assert.Equal(t, target.Host, gotHost, "Host must match Vite so its host check passes")

	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/app/@vite/ping", strings.NewReader("payload"))
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.MethodPost, gotMethod, "proxy mode forwards every method")
	assert.Equal(t, "payload", gotBody)

	// RemoveTrailingSlash turns /app/ into /app, but Vite (base '/app/') only
	// serves /app/ and its HMR websocket lives at /app/?token=…
	rec = get(e, http.MethodGet, "/app/?token=abc")
	assert.Equal(t, "/app/", gotPath)
	assert.Equal(t, "from vite", rec.Body.String())
}
