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
	rec := serve(a, "/app")
	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
	assert.Contains(t, rec.Body.String(), "The web client has not been built")
}

func TestAppServesSPAFallback(t *testing.T) {
	conf := testConfig()
	conf.UI = fstest.MapFS{"index.html": {Data: []byte(`<div id="root"></div>`)}}
	a := app.Build(conf, app.Stores{}, uploadtest.New(), mail.NewMailer(mail.Config{}))
	t.Cleanup(a.Stop)

	for _, path := range []string{"/app", "/app/news/5"} {
		rec := serve(a, path)
		assert.Equal(t, http.StatusOK, rec.Code, path)
		assert.Contains(t, rec.Body.String(), `<div id="root">`, path)
	}
	// The API and legacy 404 handling are untouched.
	assert.Equal(t, http.StatusOK, serve(a, "/api/v1/health").Code)
	assert.Equal(t, http.StatusNotFound, serve(a, "/api/v1/nope").Code)
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

	rec := serve(a, "/app/news/5")
	require.Equal(t, http.StatusOK, rec.Code)
	m := regexp.MustCompile(`/app/assets/[^"]+\.js`).FindString(rec.Body.String())
	require.NotEmpty(t, m, "index.html should reference a hashed script under /app/assets")
	asset := serve(a, m)
	assert.Equal(t, http.StatusOK, asset.Code)
	assert.Equal(t, "public, max-age=31536000, immutable", asset.Header().Get("Cache-Control"))
}
