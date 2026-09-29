package team_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/COMTOP1/AFC-GO/server/internal/auth/authtest"
	"github.com/COMTOP1/AFC-GO/server/internal/team"
	"github.com/COMTOP1/AFC-GO/server/internal/web"
	"github.com/COMTOP1/AFC-GO/server/internal/web/apitest"
)

func TestTeamRoutes(t *testing.T) {
	svc, store, _ := newService()
	sessions := authtest.Everyone()
	e := apitest.NewEcho()
	team.NewHandlers(svc).Register(web.NewAPI(e, false), sessions.Guards())
	anon := apitest.New(e)
	editor := anon.As(authtest.Cookie(t, sessions, authtest.Webmaster))
	manager := anon.As(authtest.Cookie(t, sessions, authtest.Manager))

	assert.Len(t, apitest.Decode[[]team.Public](t, anon.Get(t, "/api/v1/teams")), 1, "anonymous: active only")
	assert.Len(t, apitest.Decode[[]team.Public](t, manager.Get(t, "/api/v1/teams")), 2, "logged in: all teams")

	create := map[string]string{"name": "Reserves", "ages": "99", "isActive": "true", "leagueTable": "https://x.example.test"}
	assert.Equal(t, http.StatusForbidden, manager.Multipart(t, http.MethodPost, "/api/v1/teams", create).Code)
	rec := editor.Multipart(t, http.MethodPost, "/api/v1/teams", create)
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	assert.True(t, apitest.Decode[team.Public](t, rec).IsActive)

	rec = editor.Multipart(t, http.MethodPost, "/api/v1/teams", map[string]string{"name": "No ages"})
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	assert.Contains(t, apitest.ErrorOf(t, rec).Fields, "ages")

	rec = editor.Multipart(t, http.MethodPatch, "/api/v1/teams/1", map[string]string{"isActive": "false"})
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	assert.False(t, store.row(1).IsActive)
	assert.Equal(t, "First Team", store.row(1).Name)

	assert.Equal(t, http.StatusNoContent, editor.JSON(t, http.MethodDelete, "/api/v1/teams/3", nil).Code)
}
