package player_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/COMTOP1/AFC-GO/server/internal/auth/authtest"
	"github.com/COMTOP1/AFC-GO/server/internal/player"
	"github.com/COMTOP1/AFC-GO/server/internal/web"
	"github.com/COMTOP1/AFC-GO/server/internal/web/apitest"
)

func TestPlayerRoutes(t *testing.T) {
	svc, store, _ := newService()
	sessions := authtest.Everyone()
	e := apitest.NewEcho()
	player.NewHandlers(svc).Register(web.NewAPI(e, false), sessions.Guards())
	anon := apitest.New(e)
	editor := anon.As(authtest.Cookie(t, sessions, authtest.Webmaster))
	manager := anon.As(authtest.Cookie(t, sessions, authtest.Manager))

	assert.Equal(t, http.StatusUnauthorized, anon.Get(t, "/api/v1/players").Code, "players list is login-only")
	rec := manager.Get(t, "/api/v1/players")
	require.Equal(t, http.StatusOK, rec.Code)
	list := apitest.Decode[[]player.Public](t, rec)
	assert.Len(t, list, 3)
	for _, p := range list {
		assert.Equal(t, p.Team != nil && p.Team.ID == authtest.Manager.TeamID, p.DateOfBirth != nil,
			"the manager sees dates of birth for their own team only (player %d)", p.ID)
	}
	photographer := anon.As(authtest.Cookie(t, sessions, authtest.Photographer))
	for _, p := range apitest.Decode[[]player.Public](t, photographer.Get(t, "/api/v1/players")) {
		assert.Nil(t, p.DateOfBirth)
	}

	create := map[string]string{"name": "Signing", "teamId": "1", "dateOfBirth": "1999-04-01", "isCaptain": "false"}
	assert.Equal(t, http.StatusForbidden, manager.Multipart(t, http.MethodPost, "/api/v1/players", create).Code)
	rec = editor.Multipart(t, http.MethodPost, "/api/v1/players", create)
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())

	rec = editor.Multipart(t, http.MethodPatch, "/api/v1/players/1", map[string]string{"isCaptain": "false"})
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	assert.False(t, store.row(1).IsCaptain)

	assert.Equal(t, http.StatusNoContent, editor.JSON(t, http.MethodDelete, "/api/v1/players/3", nil).Code)
}
