// Package web holds the HTTP plumbing shared by every API handler: the JSON
// error envelope, request helpers, CSRF and the /api/v1 group.
package web

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel/trace"

	"github.com/COMTOP1/AFC-GO/server/internal/svcerr"
)

// APIPrefix is the path prefix whose errors are rendered as JSON.
const APIPrefix = "/api/"

// APIError is the body of every API error response.
type APIError struct {
	Code    int               `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

// ErrorResponse wraps APIError as {"error": {...}}.
type ErrorResponse struct {
	Error APIError `json:"error"`
}

// ErrorHandler renders errors on /api/* as the JSON envelope, and every other
// path as a small HTML page (the web client handles its own pages; this covers
// redirects, /download and anything else the server answers directly).
func ErrorHandler() echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		if c.Response().Committed {
			return
		}
		ctx := c.Request().Context()
		trace.SpanFromContext(ctx).RecordError(err)

		path := c.Request().URL.Path
		apiErr := toAPIError(err)
		if apiErr.Code >= http.StatusInternalServerError {
			slog.ErrorContext(ctx, fmt.Sprintf("error on %s %s: %+v", c.Request().Method, path, err))
		}

		var writeErr error
		switch {
		case c.Request().Method == http.MethodHead:
			writeErr = c.NoContent(apiErr.Code)
		case path == "/api" || strings.HasPrefix(path, APIPrefix):
			writeErr = c.JSON(apiErr.Code, ErrorResponse{Error: apiErr})
		default:
			writeErr = c.HTML(apiErr.Code, errorPage(apiErr.Code))
		}
		if writeErr != nil {
			slog.ErrorContext(ctx, fmt.Sprintf("failed to write error response: %+v", writeErr))
		}
	}
}

// errorPage is a minimal, dependency-free page for non-API errors. It never
// includes the error's text.
func errorPage(code int) string {
	title := "Something went wrong"
	if code == http.StatusNotFound {
		title = "Page not found"
	}
	return fmt.Sprintf(`<!doctype html><html lang="en-GB"><head><meta charset="utf-8">`+
		`<meta name="viewport" content="width=device-width, initial-scale=1">`+
		`<title>%[1]s · AFC Aldermaston</title></head>`+
		`<body style="font-family:system-ui,sans-serif;max-width:32rem;margin:4rem auto;padding:0 1rem">`+
		`<h1>%[1]s</h1><p>Error %[2]d.</p><p><a href="/">Go to the home page</a></p></body></html>`,
		title, code)
}

func toAPIError(err error) APIError {
	if se, ok := svcerr.As(err); ok {
		switch se.Kind {
		case svcerr.KindNotFound:
			return APIError{Code: http.StatusNotFound, Message: se.Message}
		case svcerr.KindForbidden:
			return APIError{Code: http.StatusForbidden, Message: se.Message}
		case svcerr.KindInvalid:
			return APIError{Code: http.StatusUnprocessableEntity, Message: se.Message, Fields: se.Fields}
		case svcerr.KindConflict:
			return APIError{Code: http.StatusConflict, Message: se.Message}
		}
	}

	var he *echo.HTTPError
	if errors.As(err, &he) {
		msg := http.StatusText(he.Code)
		if s, ok := he.Message.(string); ok && s != "" && he.Code < http.StatusInternalServerError {
			msg = s
		}
		return APIError{Code: he.Code, Message: msg}
	}

	return APIError{Code: http.StatusInternalServerError, Message: "internal server error"}
}
