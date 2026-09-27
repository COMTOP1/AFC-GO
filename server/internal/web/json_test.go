package web_test

import (
	"net/http"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/COMTOP1/AFC-GO/server/internal/web"
	"github.com/COMTOP1/AFC-GO/server/internal/web/apitest"
)

func TestFormHelpers(t *testing.T) {
	e := apitest.NewEcho()
	api := web.NewAPI(e, false)
	type seen struct {
		Title   *string `json:"title"`
		Missing *string `json:"missing"`
		Flag    *bool   `json:"flag"`
		Date    string  `json:"date"`
		File    string  `json:"file"`
		NoFile  bool    `json:"noFile"`
	}
	api.POST("/form", func(c echo.Context) error {
		var s seen
		s.Title = web.FormString(c, "title")
		s.Missing = web.FormString(c, "missing")
		var err error
		if s.Flag, err = web.FormBool(c, "flag"); err != nil {
			return err
		}
		d, err := web.FormDate(c, "date")
		if err != nil {
			return err
		}
		s.Date = d.Format(web.DateLayout)
		f, err := web.FormFile(c, "image")
		if err != nil {
			return err
		}
		s.File = f.Name
		none, err := web.FormFile(c, "other")
		if err != nil {
			return err
		}
		s.NoFile = none == nil
		return c.JSON(http.StatusOK, s)
	})

	rec := apitest.New(e).Multipart(t, http.MethodPost, "/api/v1/form",
		map[string]string{"title": "", "flag": "true", "date": "2026-09-01"},
		apitest.FilePart{Field: "image", Name: "a.png", ContentType: "image/png", Body: "x"})
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	got := apitest.Decode[seen](t, rec)
	require.NotNil(t, got.Title)
	assert.Empty(t, *got.Title, "present-but-empty must be a non-nil empty string")
	assert.Nil(t, got.Missing, "absent must be nil")
	require.NotNil(t, got.Flag)
	assert.True(t, *got.Flag)
	assert.Equal(t, "2026-09-01", got.Date)
	assert.Equal(t, "a.png", got.File)
	assert.True(t, got.NoFile)
}

func TestFormStringIgnoresQueryString(t *testing.T) {
	e := apitest.NewEcho()
	api := web.NewAPI(e, false)
	type seen struct {
		Title   *string `json:"title"`
		Missing *string `json:"missing"`
	}
	api.POST("/form", func(c echo.Context) error {
		var s seen
		s.Title = web.FormString(c, "title")
		s.Missing = web.FormString(c, "missing")
		return c.JSON(http.StatusOK, s)
	})

	// "missing" is only in the query string, never in the multipart body;
	// "title" is in both, with different values, so a pass-through bug
	// (reading request.Form instead of request.PostForm) would be caught
	// either way.
	rec := apitest.New(e).Multipart(t, http.MethodPost, "/api/v1/form?missing=fromquery&title=fromquery",
		map[string]string{"title": "frombody"})
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	got := apitest.Decode[seen](t, rec)
	require.NotNil(t, got.Title)
	assert.Equal(t, "frombody", *got.Title, "a field present in the body must use the body value, not the query string")
	assert.Nil(t, got.Missing, "a field absent from the body must be nil even when the query string has it")
}

func TestFormBoolAndDateValidation(t *testing.T) {
	e := apitest.NewEcho()
	api := web.NewAPI(e, false)
	api.POST("/bool", func(c echo.Context) error { _, err := web.FormBool(c, "flag"); return err })
	api.POST("/date", func(c echo.Context) error { _, err := web.FormDate(c, "date"); return err })

	rec := apitest.New(e).Multipart(t, http.MethodPost, "/api/v1/bool", map[string]string{"flag": "maybe"})
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	assert.Contains(t, apitest.ErrorOf(t, rec).Fields, "flag")

	rec = apitest.New(e).Multipart(t, http.MethodPost, "/api/v1/date", map[string]string{"date": "01/09/2026"})
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	assert.Contains(t, apitest.ErrorOf(t, rec).Fields, "date")
}

func TestBindJSONRejectsUnknownFieldsAndBadIDs(t *testing.T) {
	e := apitest.NewEcho()
	api := web.NewAPI(e, false)
	api.POST("/bind", func(c echo.Context) error {
		var in struct {
			Name string `json:"name"`
		}
		return web.BindJSON(c, &in)
	})
	api.GET("/items/:id", func(c echo.Context) error { _, err := web.ParamID(c, "id"); return err })

	rec := apitest.New(e).JSON(t, http.MethodPost, "/api/v1/bind", map[string]string{"nmae": "typo"}) //nolint:misspell // deliberate typo to test unknown-field rejection
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)

	for _, id := range []string{"abc", "0", "-3"} {
		rec = apitest.New(e).Get(t, "/api/v1/items/"+id)
		assert.Equal(t, http.StatusBadRequest, rec.Code, id)
	}
}
