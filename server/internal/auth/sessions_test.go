package auth_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/server/internal/auth"
	"github.com/COMTOP1/AFC-GO/server/internal/auth/authtest"
	"github.com/COMTOP1/AFC-GO/server/internal/user"
	"github.com/COMTOP1/AFC-GO/server/internal/web"
	"github.com/COMTOP1/AFC-GO/server/internal/web/apitest"
)

func newMeAPI(s *auth.Sessions) *apitest.Client {
	e := apitest.NewEcho()
	api := web.NewAPI(e, false)
	auth.NewHandlers(s, authtest.Files()).Register(api, s.Guards())
	return apitest.New(e)
}

func TestMeRequiresLogin(t *testing.T) {
	rec := newMeAPI(authtest.Everyone()).Get(t, "/api/v1/auth/me")
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Equal(t, 401, apitest.ErrorOf(t, rec).Code)
}

func TestMeReturnsUserAndPermissions(t *testing.T) {
	s := authtest.Everyone()
	rec := newMeAPI(s).As(authtest.Cookie(t, s, authtest.Treasurer)).Get(t, "/api/v1/auth/me")
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	me := apitest.Decode[auth.CurrentUser](t, rec)
	assert.Equal(t, authtest.Treasurer.ID, me.ID)
	assert.Equal(t, "Treasurer", me.Role)
	assert.True(t, me.Permissions.CanEdit)
	assert.False(t, me.Permissions.CanManageUsers)
}

func TestMeNeverLeaksSecrets(t *testing.T) {
	u := authtest.Webmaster
	u.Hash = null.StringFrom("deadbeef")
	u.Salt = null.StringFrom("cafe")
	u.Password = null.StringFrom("hunter2")
	s := authtest.New(u)
	rec := newMeAPI(s).As(authtest.Cookie(t, s, u)).Get(t, "/api/v1/auth/me")
	require.Equal(t, http.StatusOK, rec.Code)
	for _, secret := range []string{"deadbeef", "cafe", "hunter2", "hash", "salt", "password"} {
		assert.NotContains(t, rec.Body.String(), secret)
	}
}

// TestLegacySessionCookieIsRecognised pins Review Focus #2: a cookie written
// exactly as legacy LoginFunc writes it must authenticate the API, so nobody
// is logged out by the deploy.
func TestLegacySessionCookieIsRecognised(t *testing.T) {
	s := authtest.Everyone()

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	sess, _ := s.CookieStore().Get(req, s.Name())
	legacyUser := authtest.Webmaster
	legacyUser.Authenticated = true
	sess.Values["user"] = legacyUser
	require.NoError(t, sess.Save(req, rec))
	cookie := rec.Result().Cookies()[0]

	got := newMeAPI(s).As(cookie).Get(t, "/api/v1/auth/me")
	require.Equal(t, http.StatusOK, got.Code, got.Body.String())
	assert.Equal(t, authtest.Webmaster.ID, apitest.Decode[auth.CurrentUser](t, got).ID)
}

func TestLogoutClearsSession(t *testing.T) {
	s := authtest.Everyone()
	cookie := authtest.Cookie(t, s, authtest.Webmaster)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/", nil)
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()
	require.NoError(t, s.Logout(rec, req))
	cleared := rec.Result().Cookies()[0]
	assert.Negative(t, cleared.MaxAge)

	check := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
	check.AddCookie(cleared)
	_, ok := s.User(check)
	assert.False(t, ok)
}

func TestSessionCookieAttributes(t *testing.T) {
	s := auth.NewSessions(auth.Config{CookieName: "session", Secure: true}, authtest.NewUsers())
	rec := httptest.NewRecorder()
	require.NoError(t, s.Login(rec, httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/", nil), user.User{ID: 1}, true))
	c := rec.Result().Cookies()[0]
	assert.Equal(t, "session", c.Name)
	assert.True(t, c.HttpOnly)
	assert.True(t, c.Secure)
	assert.Equal(t, http.SameSiteLaxMode, c.SameSite)
	assert.Equal(t, 86400*31, c.MaxAge, "remember=true keeps the session for 31 days")
}
