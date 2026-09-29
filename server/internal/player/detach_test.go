package player_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/COMTOP1/AFC-GO/server/internal/player"
	"github.com/COMTOP1/AFC-GO/server/internal/team"
	"github.com/COMTOP1/AFC-GO/server/internal/testdb"
)

func TestDetachTeam(t *testing.T) {
	db, _ := testdb.Open(t)
	store := player.NewPlayerRepo(db)
	ctx := context.Background()
	require.NoError(t, store.DetachTeam(ctx, 1))
	left, err := store.GetPlayersTeam(ctx, team.Team{ID: 1})
	require.NoError(t, err)
	assert.Empty(t, left)
	youth, err := store.GetPlayersTeam(ctx, team.Team{ID: 2})
	require.NoError(t, err)
	assert.Len(t, youth, 1, "other teams untouched")
}
