package auth_test

import (
	"net/http"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"

	"github.com/COMTOP1/AFC-GO/server/internal/auth/authtest"
	"github.com/COMTOP1/AFC-GO/server/internal/user"
	"github.com/COMTOP1/AFC-GO/server/internal/web"
	"github.com/COMTOP1/AFC-GO/server/internal/web/apitest"
)

func TestGuards(t *testing.T) {
	s := authtest.Everyone()
	e := apitest.NewEcho()
	api := web.NewAPI(e, false)
	g := s.Guards()
	ok := func(c echo.Context) error { return c.NoContent(http.StatusNoContent) }
	api.GET("/login", ok, g.Login)
	api.GET("/editor", ok, g.Editor)
	api.GET("/gallery", ok, g.NotManager)
	api.GET("/secretary", ok, g.ClubSecretaryHigher)

	ghost := user.User{ID: 99, Email: "gone@example.test"} // has a cookie but no longer exists
	type want map[string]int
	cases := []struct {
		name string
		who  *user.User
		want want
	}{
		{"anonymous", nil, want{"/login": 401, "/editor": 401, "/gallery": 401, "/secretary": 401}},
		{"deleted user", &ghost, want{"/login": 401, "/editor": 401, "/gallery": 401, "/secretary": 401}},
		{"manager", &authtest.Manager, want{"/login": 204, "/editor": 403, "/gallery": 403, "/secretary": 403}},
		{"photographer", &authtest.Photographer, want{"/login": 204, "/editor": 403, "/gallery": 204, "/secretary": 403}},
		{"treasurer", &authtest.Treasurer, want{"/login": 204, "/editor": 204, "/gallery": 204, "/secretary": 403}},
		{"webmaster", &authtest.Webmaster, want{"/login": 204, "/editor": 204, "/gallery": 204, "/secretary": 204}},
	}
	for _, tc := range cases {
		client := apitest.New(e)
		if tc.who != nil {
			client = client.As(authtest.Cookie(t, s, *tc.who))
		}
		for path, code := range tc.want {
			rec := client.Get(t, "/api/v1"+path)
			assert.Equal(t, code, rec.Code, "%s GET %s", tc.name, path)
		}
	}
}

func TestGuardUsesFreshRole(t *testing.T) {
	// The cookie says Webmaster, but the DB now says Manager: the DB wins.
	demoted := authtest.Webmaster
	s := authtest.New(user.User{ID: demoted.ID, Name: demoted.Name, Email: demoted.Email, Role: authtest.Manager.Role})
	e := apitest.NewEcho()
	api := web.NewAPI(e, false)
	api.GET("/editor", func(c echo.Context) error { return c.NoContent(http.StatusNoContent) }, s.Guards().Editor)

	rec := apitest.New(e).As(authtest.Cookie(t, s, demoted)).Get(t, "/api/v1/editor")
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestIdentifyNeverRejects(t *testing.T) {
	s := authtest.Everyone()
	e := apitest.NewEcho()
	api := web.NewAPI(e, false)
	api.GET("/who", func(c echo.Context) error {
		if web.LoggedIn(c) {
			return c.String(http.StatusOK, "member")
		}
		return c.String(http.StatusOK, "guest")
	}, s.Guards().Identify)

	assert.Equal(t, "guest", apitest.New(e).Get(t, "/api/v1/who").Body.String())
	ghost := user.User{ID: 99, Email: "gone@example.test"}
	assert.Equal(t, "guest", apitest.New(e).As(authtest.Cookie(t, s, ghost)).Get(t, "/api/v1/who").Body.String())
	assert.Equal(t, "member", apitest.New(e).As(authtest.Cookie(t, s, authtest.Manager)).Get(t, "/api/v1/who").Body.String())
}
