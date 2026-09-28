package web_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/COMTOP1/AFC-GO/server/internal/web"
	"github.com/COMTOP1/AFC-GO/server/internal/web/apitest"
)

// TestRequireFormRejectsNonFormBody pins Review Focus #2: web.FormString
// silently returns nil when the body isn't a form, so a JSON PATCH would
// otherwise fall through and be treated as "no fields sent" (200, nothing
// changed). RequireForm must reject it outright with 415 before any handler
// logic runs.
func TestRequireFormRejectsNonFormBody(t *testing.T) {
	e := apitest.NewEcho()
	api := web.NewAPI(e, false)
	api.POST("/form", func(c echo.Context) error {
		if err := web.RequireForm(c); err != nil {
			return err
		}
		return c.NoContent(http.StatusOK)
	})

	rec := apitest.New(e).JSON(t, http.MethodPost, "/api/v1/form", map[string]string{"title": "x"})
	assert.Equal(t, http.StatusUnsupportedMediaType, rec.Code, rec.Body.String())
}

// TestRequireFormRejectsMalformedMultipart pins Review Focus #2: a body that
// claims to be multipart/form-data but fails to parse must be a 422
// (svcerr.InvalidField), not silently ignored.
func TestRequireFormRejectsMalformedMultipart(t *testing.T) {
	e := apitest.NewEcho()
	api := web.NewAPI(e, false)
	api.POST("/form", func(c echo.Context) error {
		if err := web.RequireForm(c); err != nil {
			return err
		}
		return c.NoContent(http.StatusOK)
	})

	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/v1/form", strings.NewReader("not a real multipart body"))
	req.Header.Set(echo.HeaderContentType, "multipart/form-data; boundary=broken")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	require.Equal(t, http.StatusUnprocessableEntity, rec.Code, rec.Body.String())
	assert.Contains(t, apitest.ErrorOf(t, rec).Fields, "body")
}

// TestRequireFormAcceptsFormTypes pins Review Focus #2: both form content
// types RequireForm is meant to allow must pass through untouched.
func TestRequireFormAcceptsFormTypes(t *testing.T) {
	e := apitest.NewEcho()
	api := web.NewAPI(e, false)
	api.POST("/form", func(c echo.Context) error {
		if err := web.RequireForm(c); err != nil {
			return err
		}
		return c.NoContent(http.StatusOK)
	})

	rec := apitest.New(e).Multipart(t, http.MethodPost, "/api/v1/form", map[string]string{"title": "x"})
	assert.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/v1/form", strings.NewReader("title=x"))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	rec2 := httptest.NewRecorder()
	e.ServeHTTP(rec2, req)
	assert.Equal(t, http.StatusOK, rec2.Code, rec2.Body.String())
}
