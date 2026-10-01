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

const indexHTML = `<!doctype html><html><body><div id="root"></div><script type="module" src="/assets/index-abc123.js"></script></body></html>`

func builtFS() fstest.MapFS {
	return fstest.MapFS{
		"index.html":             {Data: []byte(indexHTML)},
		"assets/index-abc123.js": {Data: []byte("console.log('app')")},
		"assets/style-9f8e.css":  {Data: []byte("body{}")},
		"favicon.svg":            {Data: []byte("<svg/>")},
		"robots.txt":             {Data: []byte("User-agent: *")},
		".keep":                  {Data: nil},
	}
}

func spaEcho(files fs.FS, proxy *url.URL) *echo.Echo {
	e := echo.New()
	e.Pre(middleware.RemoveTrailingSlash())
	web.MountSPA(e, files, proxy)
	return e
}

func get(e *echo.Echo, method, path string) *httptest.ResponseRecorder {
	req := httptest.NewRequestWithContext(context.Background(), method, path, nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

func TestSPAKnownRoutesGetTheAppWith200(t *testing.T) {
	e := spaEcho(builtFS(), nil)
	for _, path := range []string{"/", "/news", "/news/5", "/news/5/edit", "/team/3", "/teams/new", "/reset/abc-123", "/whatson?period=past", "/users"} {
		rec := get(e, http.MethodGet, path)
		assert.Equal(t, http.StatusOK, rec.Code, path)
		assert.Equal(t, indexHTML, rec.Body.String(), path)
		assert.Contains(t, rec.Header().Get(echo.HeaderContentType), "text/html", path)
		assert.Equal(t, "no-cache", rec.Header().Get("Cache-Control"), path)
	}
}

func TestSPAUnknownRoutesGetTheAppWith404(t *testing.T) {
	e := spaEcho(builtFS(), nil)
	for _, path := range []string{"/nonsense", "/news/5/oops", "/team//edit", "/teams/new/x", "/applications"} {
		rec := get(e, http.MethodGet, path)
		assert.Equal(t, http.StatusNotFound, rec.Code, path)
		assert.Equal(t, indexHTML, rec.Body.String(), "the app renders its own not-found page: %s", path)
		assert.Equal(t, "no-cache", rec.Header().Get("Cache-Control"), path)
	}
}

func TestSPAAssetsAndRootFiles(t *testing.T) {
	e := spaEcho(builtFS(), nil)

	rec := get(e, http.MethodGet, "/assets/index-abc123.js")
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "console.log('app')", rec.Body.String())
	assert.Contains(t, rec.Header().Get(echo.HeaderContentType), "javascript")
	assert.Equal(t, "public, max-age=31536000, immutable", rec.Header().Get("Cache-Control"))

	assert.Contains(t, get(e, http.MethodGet, "/assets/style-9f8e.css").Header().Get(echo.HeaderContentType), "text/css")

	rec = get(e, http.MethodGet, "/robots.txt")
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "User-agent: *", rec.Body.String())
	assert.Equal(t, "no-cache", rec.Header().Get("Cache-Control"))

	for _, missing := range []string{"/assets/index-old999.js", "/assets", "/favicon.png"} {
		rec = get(e, http.MethodGet, missing)
		assert.Equal(t, http.StatusNotFound, rec.Code, missing)
		assert.NotContains(t, rec.Body.String(), `<div id="root">`, "must never fall back to index: %s", missing)
	}
}

func TestSPALeavesTheAPIAlone(t *testing.T) {
	e := spaEcho(builtFS(), nil)
	for _, path := range []string{"/api", "/api/v1/nope"} {
		rec := get(e, http.MethodGet, path)
		assert.Equal(t, http.StatusNotFound, rec.Code, path)
		assert.NotContains(t, rec.Body.String(), `<div id="root">`, path)
	}
}

func TestSPANotBuilt(t *testing.T) {
	for name, files := range map[string]fs.FS{
		"only .keep": fstest.MapFS{".keep": {Data: nil}},
		"nil fs":     nil, // a true nil interface, as app.Config.UI is in tests
	} {
		t.Run(name, func(t *testing.T) {
			e := spaEcho(files, nil)
			for _, path := range []string{"/", "/news/5", "/assets/index-abc123.js"} {
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
	assert.Equal(t, http.StatusMethodNotAllowed, get(e, http.MethodPost, "/news").Code)
	assert.Equal(t, http.StatusMethodNotAllowed, get(e, http.MethodDelete, "/").Code)
	assert.Equal(t, http.StatusOK, get(e, http.MethodHead, "/news").Code)
	assert.Equal(t, http.StatusNotFound, get(e, http.MethodHead, "/nonsense").Code)
}

func TestSPARejectsTraversal(t *testing.T) {
	e := spaEcho(builtFS(), nil)
	for _, path := range []string{"/../go.mod", "/assets/../../go.mod", "/%2e%2e/go.mod", "/assets/%2e%2e/index.html"} {
		rec := get(e, http.MethodGet, path)
		assert.NotContains(t, rec.Body.String(), "module github.com", path)
		assert.NotEqual(t, "console.log('app')", rec.Body.String(), path)
	}
}

func TestSPADoesNotShadowOtherRoutes(t *testing.T) {
	e := spaEcho(builtFS(), nil)
	e.GET("/download", func(c echo.Context) error { return c.String(http.StatusOK, "download") })
	assert.Equal(t, "download", get(e, http.MethodGet, "/download").Body.String())
}

func TestIsAppRoute(t *testing.T) {
	for path, want := range map[string]bool{
		"/": true, "/news": true, "/news/5": true, "/news/new": true, "/news/5/edit": true,
		"/reset/abc": true, "/design": true,
		"/news/5/oops": false, "/team//edit": false, "/nope": false, "": false, "/teams/5": false,
	} {
		assert.Equal(t, want, web.IsAppRoute(path), path)
	}
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
	rec := get(e, http.MethodGet, "/main.tsx")
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "from vite", rec.Body.String())
	assert.Equal(t, "/main.tsx", gotPath)
	assert.Equal(t, target.Host, gotHost, "Host must match Vite so its host check passes")

	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/@vite/ping", strings.NewReader("payload"))
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.MethodPost, gotMethod, "proxy mode forwards every method")
	assert.Equal(t, "payload", gotBody)

	// Vite's HMR websocket lives at /?token=…
	rec = get(e, http.MethodGet, "/?token=abc")
	assert.Equal(t, "/", gotPath)
	assert.Equal(t, "from vite", rec.Body.String())

	gotPath = ""
	rec = get(e, http.MethodGet, "/api/v1/nope")
	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.Empty(t, gotPath, "API paths are never proxied to Vite")
}
