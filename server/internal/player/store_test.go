package player_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/server/internal/player"
	"github.com/COMTOP1/AFC-GO/server/internal/testdb"
)

func TestAddPlayerReturnsID(t *testing.T) {
	db, _ := testdb.Open(t)
	added, err := player.NewPlayerRepo(db).AddPlayer(context.Background(), player.Player{
		Name: "New", DateOfBirth: null.TimeFrom(time.Date(1995, 1, 1, 0, 0, 0, 0, time.UTC)), TeamID: 1})
	require.NoError(t, err)
	require.Positive(t, added.ID)
}

func TestEditPlayerClearsOptionalFields(t *testing.T) {
	db, _ := testdb.Open(t)
	store := player.NewPlayerRepo(db)
	ctx := context.Background()
	p, err := store.GetPlayer(ctx, player.Player{ID: 1})
	require.NoError(t, err)
	require.True(t, p.FileName.Valid)
	require.True(t, p.Position.Valid)
	require.True(t, p.IsCaptain)

	p.FileName, p.Position, p.IsCaptain = null.String{}, null.String{}, false
	_, err = store.EditPlayer(ctx, p)
	require.NoError(t, err)

	got, err := store.GetPlayer(ctx, player.Player{ID: 1})
	require.NoError(t, err)
	assert.False(t, got.FileName.Valid)
	assert.False(t, got.Position.Valid)
	assert.False(t, got.IsCaptain)
}
