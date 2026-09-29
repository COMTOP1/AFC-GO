package auth_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/COMTOP1/AFC-GO/server/internal/auth"
	"github.com/COMTOP1/AFC-GO/server/internal/auth/authtest"
	"github.com/COMTOP1/AFC-GO/server/internal/web"
	"github.com/COMTOP1/AFC-GO/server/internal/web/apitest"
)

func TestAuthRoutes(t *testing.T) {
	svc, users, tokens := newAuthService()
	sessions := authtest.New(users.byEmail["ok@example.test"], users.byEmail["reset@example.test"])
	e := apitest.NewEcho()
	auth.NewHandlers(sessions, svc, authtest.Files()).Register(web.NewAPI(e, false), sessions.Guards())
	anon := apitest.New(e)

	rec := anon.JSON(t, http.MethodPost, "/api/v1/auth/login", auth.LoginInput{Email: "ok@example.test", Password: "wrong"})
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Equal(t, "invalid email or password", apitest.ErrorOf(t, rec).Message)

	rec = anon.JSON(t, http.MethodPost, "/api/v1/auth/login", auth.LoginInput{Email: "reset@example.test", Password: "Right-Pa55"})
	require.Equal(t, http.StatusOK, rec.Code)
	res := apitest.Decode[auth.LoginResponse](t, rec)
	assert.True(t, res.ResetRequired)
	assert.Nil(t, res.User)
	assert.Empty(t, rec.Result().Cookies(), "no session for a reset-required login")

	rec = anon.JSON(t, http.MethodPost, "/api/v1/auth/login", auth.LoginInput{Email: "ok@example.test", Password: "Right-Pa55", Remember: true})
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	require.NotNil(t, apitest.Decode[auth.LoginResponse](t, rec).User)
	var session *http.Cookie
	for _, c := range rec.Result().Cookies() {
		if c.Name == sessions.Name() {
			session = c
		}
	}
	require.NotNil(t, session)
	me := anon.As(session)
	assert.Equal(t, http.StatusOK, me.Get(t, "/api/v1/auth/me").Code)

	rec = me.JSON(t, http.MethodPost, "/api/v1/auth/password",
		auth.PasswordInput{OldPassword: "Right-Pa55", NewPassword: "weak", ConfirmationPassword: "weak"})
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	rec = me.JSON(t, http.MethodPost, "/api/v1/auth/password",
		auth.PasswordInput{OldPassword: "Right-Pa55", NewPassword: "Str0ng-Pass!", ConfirmationPassword: "Str0ng-Pass!"})
	assert.Equal(t, http.StatusNoContent, rec.Code)

	rec = me.JSON(t, http.MethodPost, "/api/v1/auth/logout", nil)
	assert.Equal(t, http.StatusNoContent, rec.Code)

	require.NoError(t, tokens.Set(ctx, "tok", 2, time.Hour))
	assert.Equal(t, http.StatusNoContent, anon.Get(t, "/api/v1/auth/reset/tok").Code)
	assert.Equal(t, http.StatusNotFound, anon.Get(t, "/api/v1/auth/reset/bad").Code)
	rec = anon.JSON(t, http.MethodPost, "/api/v1/auth/reset/tok",
		auth.ResetInput{NewPassword: "Str0ng-Pass!", ConfirmationPassword: "Str0ng-Pass!"})
	assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.Equal(t, http.StatusNotFound, anon.Get(t, "/api/v1/auth/reset/tok").Code, "single use")
}
