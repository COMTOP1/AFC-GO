package app_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/COMTOP1/AFC-GO/server/internal/app"
	"github.com/COMTOP1/AFC-GO/server/internal/infrastructure/mail"
	"github.com/COMTOP1/AFC-GO/server/internal/site"
	"github.com/COMTOP1/AFC-GO/server/internal/testdb"
	"github.com/COMTOP1/AFC-GO/server/internal/upload/uploadtest"
	"github.com/COMTOP1/AFC-GO/server/internal/web/apitest"
)

// TestAnonymousTeamDetailHidesYouthSquad pins Review Focus #1, CRITICAL child
// safety: GET /api/v1/teams/{id} must not list a youth team's players (names
// or positions) to an anonymous caller, matching the legacy team page's
// `{{if and (not .Team.IsYouth) ...}}` gate. Seed data: team 1 "First Team"
// is a senior team, team 2 "Under 12s" is youth.
func TestAnonymousTeamDetailHidesYouthSquad(t *testing.T) {
	db, _ := testdb.Open(t)
	a := app.Build(testConfig(), app.NewStores(db), uploadtest.New(), mail.NewMailer(mail.Config{}))
	t.Cleanup(a.Stop)

	rec := apitest.New(a.Echo).Get(t, "/api/v1/teams/2")
	require.Equal(t, 200, rec.Code, "body: %s", rec.Body.String())
	detail := apitest.Decode[site.TeamDetail](t, rec)
	assert.Empty(t, detail.Players, "youth team players must not be listed to anonymous callers")

	rec = apitest.New(a.Echo).Get(t, "/api/v1/teams/1")
	require.Equal(t, 200, rec.Code, "body: %s", rec.Body.String())
	detail = apitest.Decode[site.TeamDetail](t, rec)
	require.Len(t, detail.Players, 2, "senior team still lists its players")
	byName := map[string]string{}
	for _, p := range detail.Players {
		byName[p.Name] = p.ImageURL
	}
	assert.NotEmpty(t, byName["Adult Player"], "adult player keeps a photo")
	assert.Empty(t, byName["Young Senior"], "under-18 player on a senior team has no photo")
}
