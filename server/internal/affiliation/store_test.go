package affiliation_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/COMTOP1/AFC-GO/server/internal/affiliation"
	"github.com/COMTOP1/AFC-GO/server/internal/testdb"
)

func TestAddAffiliationReturnsID(t *testing.T) {
	db, _ := testdb.Open(t)
	added, err := affiliation.NewAffiliationRepo(db).AddAffiliation(context.Background(), affiliation.Affiliation{Name: "League"})
	require.NoError(t, err)
	require.Positive(t, added.ID)
}
