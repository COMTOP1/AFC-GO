package web

import (
	"strings"

	"github.com/labstack/echo/v5"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// tokenPathPrefixes are paths whose next segment is a password reset token.
// A token is a live credential for up to a week, so it must never reach
// traces or logs.
var tokenPathPrefixes = []string{"/reset/", "/api/v1/auth/reset/", "/auth/reset/"}

// RedactPath replaces a reset token in p with ":token". Other paths are
// returned unchanged. p may be a site path, an /api/v1 path, or an API path
// relative to /api/v1 as the client reports it.
func RedactPath(p string) string {
	for _, prefix := range tokenPathPrefixes {
		rest, ok := strings.CutPrefix(p, prefix)
		if !ok || rest == "" {
			continue
		}
		end := strings.IndexAny(rest, "/?#")
		if end < 0 {
			end = len(rest)
		}
		return prefix + ":token" + rest[end:]
	}
	return p
}

// RedactSpanPath overwrites the url.path the tracing middleware recorded
// when it holds a reset token. Register it after the tracing middleware.
func RedactSpanPath(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		p := c.Request().URL.Path
		if redacted := RedactPath(p); redacted != p {
			trace.SpanFromContext(c.Request().Context()).SetAttributes(attribute.String("url.path", redacted))
		}
		return next(c)
	}
}
