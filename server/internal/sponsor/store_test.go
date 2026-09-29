package sponsor_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/COMTOP1/AFC-GO/server/internal/sponsor"
	"github.com/COMTOP1/AFC-GO/server/internal/testdb"
)

func TestAddSponsorReturnsID(t *testing.T) {
	db, _ := testdb.Open(t)
	added, err := sponsor.NewSponsorRepo(db).AddSponsor(context.Background(), sponsor.Sponsor{Name: "New", TeamID: "A"})
	require.NoError(t, err)
	require.Positive(t, added.ID)
}
