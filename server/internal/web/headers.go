package web

import (
	"net/url"
	"strings"

	"github.com/labstack/echo/v5"
)

// contentSecurityPolicy only lets the site run its own scripts, so injected
// script in an article or a compromised dependency URL can't run. Styles
// allow inline attributes (React and article formatting use them), images
// may come from any HTTPS host (the file CDN, article images), the page may
// fetch from the file store (programme PDFs are drawn in the page), and only
// Google Maps may be framed (the contact page map).
func contentSecurityPolicy(fileOrigin string) string {
	connect := "'self'"
	if fileOrigin != "" {
		connect += " " + fileOrigin
	}
	return "default-src 'self'; script-src 'self'; style-src 'self'; " +
		"img-src 'self' data: blob: https:; font-src 'self'; connect-src " + connect + "; " +
		"frame-src https://www.google.com; frame-ancestors 'none'; base-uri 'self'; form-action 'self'; object-src 'none'"
}

// Origin returns the scheme and host of rawURL (e.g. the S3 endpoint), or
// "" when it has none.
func Origin(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return ""
	}
	return u.Scheme + "://" + u.Host
}

// SecurityHeaders sets the browser protections on every response. hsts is
// for sites served only over HTTPS. csp is off in development, where the
// Vite dev server needs inline scripts and websockets; the Swagger UI under
// /api/v1/swagger never gets it, for the same reason. fileOrigin is where
// uploaded files are served from (see Origin).
func SecurityHeaders(hsts, csp bool, fileOrigin string) echo.MiddlewareFunc {
	policy := contentSecurityPolicy(fileOrigin)
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			h := c.Response().Header()
			h.Set("X-Content-Type-Options", "nosniff")
			h.Set("X-Frame-Options", "DENY")
			h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
			h.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=(), payment=()")
			if hsts {
				h.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
			}
			if csp && !strings.HasPrefix(c.Request().URL.Path, APIPrefix+"v1/swagger") {
				h.Set("Content-Security-Policy", policy)
			}
			return next(c)
		}
	}
}
