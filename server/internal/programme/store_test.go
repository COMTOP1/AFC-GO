package programme_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/COMTOP1/AFC-GO/server/internal/programme"
	"github.com/COMTOP1/AFC-GO/server/internal/testdb"
)

func TestInsertsReturnIDs(t *testing.T) {
	db, _ := testdb.Open(t)
	store := programme.NewProgrammeRepo(db)
	ctx := context.Background()

	season, err := store.AddSeason(ctx, programme.Season{Season: "2026-27"})
	require.NoError(t, err)
	require.Positive(t, season.ID)

	p, err := store.AddProgramme(ctx, programme.Programme{Name: "Home v Away", FileName: "programme/p.pdf",
		DateOfProgramme: time.Now(), SeasonID: season.ID})
	require.NoError(t, err)
	require.Positive(t, p.ID)
}
