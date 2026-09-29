package user_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/COMTOP1/AFC-GO/server/internal/team"
	"github.com/COMTOP1/AFC-GO/server/internal/testdb"
	"github.com/COMTOP1/AFC-GO/server/internal/user"
)

func TestDetachTeam(t *testing.T) {
	db, _ := testdb.Open(t)
	store := user.NewUserRepo(db)
	ctx := context.Background()
	require.NoError(t, store.DetachTeam(ctx, 1))
	managers, err := store.GetUsersManagersTeam(ctx, team.Team{ID: 1})
	require.NoError(t, err)
	assert.Empty(t, managers)
}
