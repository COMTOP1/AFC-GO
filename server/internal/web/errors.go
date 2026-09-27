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

// ErrorHandler renders errors on /api/* as the JSON envelope and hands every
// other path to the legacy (HTML) handler.
func ErrorHandler(legacy echo.HTTPErrorHandler) echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		path := c.Request().URL.Path
		if path != "/api" && !strings.HasPrefix(path, APIPrefix) {
			legacy(err, c)
			return
		}
		if c.Response().Committed {
			return
		}
		ctx := c.Request().Context()
		trace.SpanFromContext(ctx).RecordError(err)

		apiErr := toAPIError(err)
		if apiErr.Code >= http.StatusInternalServerError {
			slog.ErrorContext(ctx, fmt.Sprintf("api error on %s %s: %+v", c.Request().Method, path, err))
		}

		var writeErr error
		if c.Request().Method == http.MethodHead {
			writeErr = c.NoContent(apiErr.Code)
		} else {
			writeErr = c.JSON(apiErr.Code, ErrorResponse{Error: apiErr})
		}
		if writeErr != nil {
			slog.ErrorContext(ctx, fmt.Sprintf("failed to write api error: %+v", writeErr))
		}
	}
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
