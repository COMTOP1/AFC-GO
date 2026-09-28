package user_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/COMTOP1/AFC-GO/server/internal/auth/authtest"
	"github.com/COMTOP1/AFC-GO/server/internal/user"
	"github.com/COMTOP1/AFC-GO/server/internal/web"
	"github.com/COMTOP1/AFC-GO/server/internal/web/apitest"
)

func TestUserRoutes(t *testing.T) {
	h := newHarness(nil)
	sessions := authtest.Everyone()
	e := apitest.NewEcho()
	user.NewHandlers(h.svc).Register(web.NewAPI(e, false), sessions.Guards())
	anon := apitest.New(e)
	treasurer := anon.As(authtest.Cookie(t, sessions, authtest.Treasurer))
	admin := anon.As(authtest.Cookie(t, sessions, authtest.Webmaster))

	for _, path := range []string{"/api/v1/users", "/api/v1/users/3"} {
		assert.Equal(t, http.StatusUnauthorized, anon.Get(t, path).Code, path)
		assert.Equal(t, http.StatusForbidden, treasurer.Get(t, path).Code, path)
	}

	rec := admin.Get(t, "/api/v1/users")
	require.Equal(t, http.StatusOK, rec.Code)
	for _, secret := range []string{"keep-me", "hash", "salt", "password"} {
		assert.NotContains(t, rec.Body.String(), secret)
	}

	rec = admin.Multipart(t, http.MethodPost, "/api/v1/users",
		map[string]string{"name": "New", "email": "new@example.test", "role": "treasurer"})
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	assert.True(t, apitest.Decode[user.Created](t, rec).EmailSent)

	rec = admin.Multipart(t, http.MethodPost, "/api/v1/users",
		map[string]string{"name": "Dup", "email": "secretary@example.test", "role": "treasurer"})
	assert.Equal(t, http.StatusConflict, rec.Code)

	rec = admin.Multipart(t, http.MethodPatch, "/api/v1/users/3", map[string]string{"phone": ""})
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	rec = admin.JSON(t, http.MethodPost, "/api/v1/users/3/reset", nil)
	require.Equal(t, http.StatusOK, rec.Code)
	assert.True(t, apitest.Decode[user.ResetResult](t, rec).EmailSent)

	assert.Equal(t, http.StatusNoContent, admin.JSON(t, http.MethodDelete, "/api/v1/users/3", nil).Code)
}
