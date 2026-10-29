package web_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"github.com/COMTOP1/AFC-GO/server/internal/web"
)

func TestRedactPath(t *testing.T) {
	for in, want := range map[string]string{
		"/reset/abc-123":             "/reset/:token",
		"/api/v1/auth/reset/abc-123": "/api/v1/auth/reset/:token",
		"/auth/reset/abc-123":        "/auth/reset/:token",
		"/auth/reset/abc-123?x=1":    "/auth/reset/:token?x=1",
		"/reset/abc-123/extra":       "/reset/:token/extra",
		"/reset/":                    "/reset/",
		"/news/5":                    "/news/5",
		"/api/v1/auth/login":         "/api/v1/auth/login",
		"/password-reset/abc":        "/password-reset/abc",
	} {
		assert.Equal(t, want, web.RedactPath(in), in)
	}
}

func TestRedactSpanPathOverwritesTheRecordedPath(t *testing.T) {
	rec := tracetest.NewSpanRecorder()
	tracer := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(rec)).Tracer("test")

	e := echo.New()
	// Stands in for otelecho, which records the raw path as url.path.
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			ctx, span := tracer.Start(c.Request().Context(), "request")
			defer span.End()
			span.SetAttributes(attribute.String("url.path", c.Request().URL.Path))
			c.SetRequest(c.Request().WithContext(ctx))
			return next(c)
		}
	})
	e.Use(web.RedactSpanPath)
	e.GET("/api/v1/auth/reset/:token", func(c *echo.Context) error { return c.NoContent(http.StatusOK) })

	e.ServeHTTP(httptest.NewRecorder(), httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/auth/reset/secret-token", nil))

	spans := rec.Ended()
	require.Len(t, spans, 1)
	var got string
	for _, a := range spans[0].Attributes() {
		if a.Key == "url.path" {
			got = a.Value.AsString()
		}
	}
	assert.Equal(t, "/api/v1/auth/reset/:token", got)
}
