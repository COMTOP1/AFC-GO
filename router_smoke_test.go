package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/COMTOP1/AFC-GO/infrastructure/storage"
	"github.com/COMTOP1/AFC-GO/testdb"
	"github.com/COMTOP1/AFC-GO/views"
)

// TestLegacyPagesSmoke guards the template site through the restructure:
// every public page renders and every guarded page redirects when logged out.
func TestLegacyPagesSmoke(t *testing.T) {
	_, dsn := testdb.Open(t)

	conf := &views.Config{
		DatabaseURL:       dsn,
		DomainName:        "localhost",
		SessionCookieName: "session",
		S3: storage.Config{
			Endpoint: "http://s3.invalid", Region: "us-east-1", Bucket: "test",
			AccessKey: "test", SecretKey: "test",
		},
		Security: views.SecurityConfig{
			Iterations: 1, ScryptWorkFactor: 2, ScryptBlockSize: 1, ScryptParallelismFactor: 1, KeyLength: 32,
		},
	}
	v := views.New(conf, "test", time.Hour)
	t.Cleanup(v.Stop)
	r := NewRouter(&RouterConf{Config: conf, Views: v})

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
			req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, tc.path, nil)
			rec := httptest.NewRecorder()
			r.router.ServeHTTP(rec, req)
			assert.Equal(t, tc.want, rec.Code, "body: %.300s", rec.Body.String())
		})
	}
}
