package team_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/COMTOP1/AFC-GO/server/internal/team"
	"github.com/COMTOP1/AFC-GO/server/internal/testdb"
)

func TestAddTeamReturnsID(t *testing.T) {
	db, _ := testdb.Open(t)
	added, err := team.NewTeamRepo(db).AddTeam(context.Background(), team.Team{Name: "Reserves", IsActive: true, Ages: 99})
	require.NoError(t, err)
	require.Positive(t, added.ID)
}
