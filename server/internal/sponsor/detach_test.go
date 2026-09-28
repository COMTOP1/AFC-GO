package sponsor_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/COMTOP1/AFC-GO/server/internal/sponsor"
	"github.com/COMTOP1/AFC-GO/server/internal/team"
	"github.com/COMTOP1/AFC-GO/server/internal/testdb"
)

func TestDetachTeam(t *testing.T) {
	db, _ := testdb.Open(t)
	store := sponsor.NewSponsorRepo(db)
	ctx := context.Background()
	require.NoError(t, store.DetachTeam(ctx, 1))
	left, err := store.GetSponsorsTeam(ctx, team.Team{ID: 1})
	require.NoError(t, err)
	assert.Empty(t, left)
	s, err := store.GetSponsor(ctx, sponsor.Sponsor{ID: 2})
	require.NoError(t, err)
	assert.Equal(t, "A", s.TeamID, "team sponsors become club-wide, as legacy TeamDeleteFunc did")
}
