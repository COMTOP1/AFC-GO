package programme_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/COMTOP1/AFC-GO/server/internal/auth/authtest"
	"github.com/COMTOP1/AFC-GO/server/internal/programme"
	"github.com/COMTOP1/AFC-GO/server/internal/web"
	"github.com/COMTOP1/AFC-GO/server/internal/web/apitest"
)

func TestProgrammeRoutes(t *testing.T) {
	svc, _, _ := newService()
	sessions := authtest.Everyone()
	e := apitest.NewEcho()
	programme.NewHandlers(svc).Register(web.NewAPI(e, false), sessions.Guards())
	anon := apitest.New(e)
	editor := anon.As(authtest.Cookie(t, sessions, authtest.Webmaster))
	manager := anon.As(authtest.Cookie(t, sessions, authtest.Manager))

	rec := anon.Get(t, "/api/v1/programmes?season=1")
	require.Equal(t, http.StatusOK, rec.Code)
	assert.Len(t, apitest.Decode[[]programme.Public](t, rec), 1)
	assert.Equal(t, http.StatusBadRequest, anon.Get(t, "/api/v1/programmes?season=abc").Code)
	assert.Equal(t, http.StatusNotFound, anon.Get(t, "/api/v1/programmes?season=99").Code)

	pdfPart := apitest.FilePart{Field: "file", Name: "p.pdf", ContentType: "application/pdf", Body: "PDF"}
	fields := map[string]string{"name": "Cup", "date": "2026-01-10", "seasonId": "1"}
	assert.Equal(t, http.StatusForbidden, manager.Multipart(t, http.MethodPost, "/api/v1/programmes", fields, pdfPart).Code)
	assert.Equal(t, http.StatusCreated, editor.Multipart(t, http.MethodPost, "/api/v1/programmes", fields, pdfPart).Code)
	assert.Equal(t, http.StatusNoContent, editor.JSON(t, http.MethodDelete, "/api/v1/programmes/2", nil).Code)

	rec = anon.Get(t, "/api/v1/seasons")
	require.Equal(t, http.StatusOK, rec.Code)
	assert.Len(t, apitest.Decode[[]programme.PublicSeason](t, rec), 1)

	rec = editor.JSON(t, http.MethodPost, "/api/v1/seasons", map[string]string{"season": "2026-27"})
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	created := apitest.Decode[programme.PublicSeason](t, rec)
	assert.Equal(t, http.StatusOK, editor.JSON(t, http.MethodPatch, "/api/v1/seasons/1", map[string]string{"season": "25/26"}).Code)
	assert.Equal(t, http.StatusUnprocessableEntity, editor.JSON(t, http.MethodPatch, "/api/v1/seasons/1", map[string]string{"season": ""}).Code)
	assert.Equal(t, http.StatusForbidden, manager.JSON(t, http.MethodDelete, "/api/v1/seasons/1", nil).Code)
	assert.Equal(t, http.StatusNoContent, editor.JSON(t, http.MethodDelete, "/api/v1/seasons/1", nil).Code)
	assert.Positive(t, created.ID)
}
