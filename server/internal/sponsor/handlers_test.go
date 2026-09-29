package sponsor_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/COMTOP1/AFC-GO/server/internal/auth/authtest"
	"github.com/COMTOP1/AFC-GO/server/internal/sponsor"
	"github.com/COMTOP1/AFC-GO/server/internal/web"
	"github.com/COMTOP1/AFC-GO/server/internal/web/apitest"
)

func TestSponsorRoutes(t *testing.T) {
	svc, store, _ := newService(clubSponsor())
	sessions := authtest.Everyone()
	e := apitest.NewEcho()
	sponsor.NewHandlers(svc).Register(web.NewAPI(e, false), sessions.Guards())
	anon := apitest.New(e)
	editor := anon.As(authtest.Cookie(t, sessions, authtest.Treasurer))
	manager := anon.As(authtest.Cookie(t, sessions, authtest.Manager))

	rec := anon.Get(t, "/api/v1/sponsors")
	require.Equal(t, http.StatusOK, rec.Code)
	assert.Len(t, apitest.Decode[[]sponsor.Public](t, rec), 1)

	logo := apitest.FilePart{Field: "image", Name: "l.png", ContentType: "image/png", Body: "PNG"}
	fields := map[string]string{"name": "Kit Co", "team": "O"}
	assert.Equal(t, http.StatusUnauthorized, anon.Multipart(t, http.MethodPost, "/api/v1/sponsors", fields, logo).Code)
	assert.Equal(t, http.StatusForbidden, manager.Multipart(t, http.MethodPost, "/api/v1/sponsors", fields, logo).Code)

	rec = editor.Multipart(t, http.MethodPost, "/api/v1/sponsors", fields, logo)
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	assert.Equal(t, "O", apitest.Decode[sponsor.Public](t, rec).Team)

	rec = editor.Multipart(t, http.MethodPost, "/api/v1/sponsors", fields) // no image
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	assert.Contains(t, apitest.ErrorOf(t, rec).Fields, "image")

	assert.Equal(t, http.StatusForbidden, manager.JSON(t, http.MethodDelete, "/api/v1/sponsors/1", nil).Code)
	assert.Equal(t, http.StatusNoContent, editor.JSON(t, http.MethodDelete, "/api/v1/sponsors/1", nil).Code)
	_, stillThere := store.rows[1]
	assert.False(t, stillThere)
}
