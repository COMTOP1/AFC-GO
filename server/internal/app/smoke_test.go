package app_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/COMTOP1/AFC-GO/server/internal/app"
	"github.com/COMTOP1/AFC-GO/server/internal/infrastructure/mail"
	"github.com/COMTOP1/AFC-GO/server/internal/testdb"
	"github.com/COMTOP1/AFC-GO/server/internal/upload/uploadtest"
)

// TestLegacyPagesSmoke guards the template site through the restructure:
// every public page renders and every guarded page redirects when logged out.
func TestLegacyPagesSmoke(t *testing.T) {
	db, _ := testdb.Open(t)
	a := app.Build(testConfig(), app.NewStores(db), uploadtest.New(), mail.NewMailer(mail.Config{}))
	t.Cleanup(a.Stop)

	cases := []struct {
		path string
		want int
	}{
		{"/", http.StatusOK},
		{"/news", http.StatusOK},
		{"/news/1", http.StatusOK},
		{"/whatson", http.StatusOK},
		{"/whatson/1", http.StatusOK},
		{"/teams", http.StatusOK},
		{"/team/1", http.StatusOK},
		{"/sponsors", http.StatusOK},
		{"/gallery", http.StatusOK},
		{"/documents", http.StatusOK},
		{"/programmes", http.StatusOK},
		{"/info", http.StatusOK},
		{"/contact", http.StatusOK},
		{"/public/stylesheet.css", http.StatusOK},
		{"/does-not-exist", http.StatusNotFound},
		{"/players", http.StatusFound},
		{"/account", http.StatusFound},
		{"/users", http.StatusFound},
	}
	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			rec := httptest.NewRecorder()
			a.Echo.ServeHTTP(rec, httptest.NewRequestWithContext(context.Background(), http.MethodGet, tc.path, nil))
			assert.Equal(t, tc.want, rec.Code, "body: %.300s", rec.Body.String())
		})
	}
}
