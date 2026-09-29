package web_test

import (
	"errors"
	"net/http"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"

	"github.com/COMTOP1/AFC-GO/server/internal/svcerr"
	"github.com/COMTOP1/AFC-GO/server/internal/web"
	"github.com/COMTOP1/AFC-GO/server/internal/web/apitest"
)

func TestErrorMapping(t *testing.T) {
	cases := []struct {
		name    string
		err     error
		code    int
		message string
		fields  map[string]string
	}{
		{"not found", svcerr.NotFound("news article not found", nil), 404, "news article not found", nil},
		{"forbidden", svcerr.Forbidden("nope"), 403, "nope", nil},
		{"invalid", svcerr.InvalidField("title", "title is required"), 422, "validation failed", map[string]string{"title": "title is required"}},
		{"conflict", svcerr.Conflict("email already used", nil), 409, "email already used", nil},
		{"echo error keeps message", echo.NewHTTPError(http.StatusUnauthorized, "login required"), 401, "login required", nil},
		{"echo 5xx hides message", echo.NewHTTPError(http.StatusBadGateway, "upstream exploded at 10.0.0.1"), 502, "Bad Gateway", nil},
		{"plain error is 500", errors.New("pq: connection refused"), 500, "internal server error", nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := apitest.NewEcho()
			api := web.NewAPI(e, false)
			api.GET("/boom", func(echo.Context) error { return tc.err })

			rec := apitest.New(e).Get(t, "/api/v1/boom")
			assert.Equal(t, tc.code, rec.Code)
			got := apitest.ErrorOf(t, rec)
			assert.Equal(t, tc.code, got.Code)
			assert.Equal(t, tc.message, got.Message)
			assert.Equal(t, tc.fields, got.Fields)
		})
	}
}

func TestNonAPIErrorsGoToLegacyHandler(t *testing.T) {
	e := apitest.NewEcho()
	web.NewAPI(e, false)
	e.GET("/news", func(echo.Context) error { return errors.New("boom") })

	rec := apitest.New(e).Get(t, "/news")
	assert.Equal(t, "legacy error", rec.Body.String())
}
