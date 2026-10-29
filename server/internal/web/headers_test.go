package web_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"

	"github.com/COMTOP1/AFC-GO/server/internal/web"
)

func TestSecurityHeaders(t *testing.T) {
	serve := func(hsts, csp bool, path string) http.Header {
		e := echo.New()
		e.Use(web.SecurityHeaders(hsts, csp, "https://files.example.test"))
		e.GET("/*", func(c *echo.Context) error { return c.NoContent(http.StatusOK) })
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, httptest.NewRequestWithContext(context.Background(), http.MethodGet, path, nil))
		return rec.Header()
	}

	h := serve(true, true, "/")
	assert.Equal(t, "nosniff", h.Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", h.Get("X-Frame-Options"))
	assert.Equal(t, "strict-origin-when-cross-origin", h.Get("Referrer-Policy"))
	assert.Contains(t, h.Get("Strict-Transport-Security"), "max-age=31536000")
	csp := h.Get("Content-Security-Policy")
	assert.Contains(t, csp, "script-src 'self';")
	assert.Contains(t, csp, "frame-src https://www.google.com")
	assert.Contains(t, csp, "connect-src 'self' https://files.example.test;", "programme PDFs are fetched from the file store")
	assert.NotContains(t, strings.Split(csp, "style-src")[0], "unsafe-inline", "only styles may be inline, never scripts")

	assert.Empty(t, serve(true, true, "/api/v1/swagger/index.html").Get("Content-Security-Policy"), "Swagger UI needs inline scripts")
	dev := serve(false, false, "/")
	assert.Empty(t, dev.Get("Content-Security-Policy"), "the Vite dev server needs inline scripts")
	assert.Empty(t, dev.Get("Strict-Transport-Security"), "no HSTS over plain HTTP")
	assert.Equal(t, "DENY", dev.Get("X-Frame-Options"))
}

func TestOrigin(t *testing.T) {
	assert.Equal(t, "https://s3.example.test:7480", web.Origin("https://s3.example.test:7480/"))
	assert.Equal(t, "https://s3.example.test", web.Origin("https://s3.example.test/path?q=1"))
	assert.Empty(t, web.Origin("not a url"))
	assert.Empty(t, web.Origin(""))
}
