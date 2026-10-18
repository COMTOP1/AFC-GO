package team_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/server/internal/team"
	"github.com/COMTOP1/AFC-GO/server/internal/testdb"
)

func TestAddTeamReturnsID(t *testing.T) {
	db, _ := testdb.Open(t)
	added, err := team.NewTeamRepo(db).AddTeam(context.Background(), team.Team{Name: "Reserves", IsActive: true, Ages: 99})
	require.NoError(t, err)
	require.Positive(t, added.ID)
}

func TestTeamContactEmailRoundTrips(t *testing.T) {
	db, _ := testdb.Open(t)
	store := team.NewTeamRepo(db)
	ctx := context.Background()
	added, err := store.AddTeam(ctx, team.Team{Name: "Under 12s", IsYouth: true, Ages: 12, ContactEmail: null.StringFrom("u12s@example.test")})
	require.NoError(t, err)
	got, err := store.GetTeam(ctx, team.Team{ID: added.ID})
	require.NoError(t, err)
	assert.Equal(t, "u12s@example.test", got.ContactEmail.String)

	got.ContactEmail = null.String{}
	_, err = store.EditTeam(ctx, got)
	require.NoError(t, err)
	got, err = store.GetTeam(ctx, team.Team{ID: added.ID})
	require.NoError(t, err)
	assert.False(t, got.ContactEmail.Valid)
}
