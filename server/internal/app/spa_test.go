package app_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/COMTOP1/AFC-GO/server/internal/app"
	"github.com/COMTOP1/AFC-GO/server/internal/infrastructure/mail"
	"github.com/COMTOP1/AFC-GO/server/internal/upload/uploadtest"
)

func serve(a *app.App, path string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	a.Echo.ServeHTTP(rec, httptest.NewRequestWithContext(context.Background(), http.MethodGet, path, nil))
	return rec
}

func TestAppSPANotBuiltByDefault(t *testing.T) {
	a := bareApp(t)
	rec := serve(a, "/")
	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
	assert.Contains(t, rec.Body.String(), "The web client has not been built")
}

func TestAppServesTheClientAtTheRoot(t *testing.T) {
	conf := testConfig()
	conf.UI = fstest.MapFS{"index.html": {Data: []byte(`<div id="root"></div>`)}}
	a := app.Build(conf, app.Stores{}, uploadtest.New(), mail.NewMailer(mail.Config{}))
	t.Cleanup(a.Stop)

	for _, path := range []string{"/", "/news", "/news/5", "/players", "/reset/abc"} {
		rec := serve(a, path)
		assert.Equal(t, http.StatusOK, rec.Code, path)
		assert.Contains(t, rec.Body.String(), `<div id="root">`, path)
	}
	rec := serve(a, "/nonsense")
	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.Contains(t, rec.Body.String(), `<div id="root">`)

	// Old URLs redirect; the API keeps its JSON 404; /download is the server's.
	assert.Equal(t, "/news/5", serve(a, "/app/news/5").Header().Get("Location"))
	assert.Equal(t, "/programmes?season=2", serve(a, "/programmes/2").Header().Get("Location"))
	assert.Equal(t, http.StatusOK, serve(a, "/api/v1/health").Code)
	api404 := serve(a, "/api/v1/nope")
	assert.Equal(t, http.StatusNotFound, api404.Code)
	assert.Contains(t, api404.Header().Get("Content-Type"), "application/json")
	assert.Equal(t, http.StatusBadRequest, serve(a, "/download?s=zz&id=1").Code)
	assert.Equal(t, http.StatusNotFound, serve(a, "/public/stylesheet.css").Code, "classic static files are gone")
}

func TestAppServesBuiltClient(t *testing.T) {
	dir := os.Getenv("AFC_BUILT_UI")
	if dir == "" {
		t.Skip("AFC_BUILT_UI not set; run after `yarn build:client` with AFC_BUILT_UI=$PWD/build/client")
	}
	conf := testConfig()
	conf.UI = os.DirFS(dir)
	a := app.Build(conf, app.Stores{}, uploadtest.New(), mail.NewMailer(mail.Config{}))
	t.Cleanup(a.Stop)

	rec := serve(a, "/news/5")
	require.Equal(t, http.StatusOK, rec.Code)
	m := regexp.MustCompile(`/assets/[^"]+\.js`).FindString(rec.Body.String())
	require.NotEmpty(t, m, "index.html should reference a hashed script under /assets")
	asset := serve(a, m)
	assert.Equal(t, http.StatusOK, asset.Code)
	assert.Equal(t, "public, max-age=31536000, immutable", asset.Header().Get("Cache-Control"))
	assert.Equal(t, http.StatusOK, serve(a, "/favicon.ico").Code)
	assert.Equal(t, http.StatusOK, serve(a, "/AFC.png").Code)
}
