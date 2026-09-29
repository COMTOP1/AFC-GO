package whatson_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/COMTOP1/AFC-GO/server/internal/testdb"
	"github.com/COMTOP1/AFC-GO/server/internal/whatson"
)

func TestAddWhatsOnReturnsID(t *testing.T) {
	db, _ := testdb.Open(t)
	store := whatson.NewWhatsOnRepo(db)
	ctx := context.Background()

	added, err := store.AddWhatsOn(ctx, whatson.WhatsOn{Title: "Quiz", DateOfEvent: time.Now().AddDate(0, 0, 7)})
	require.NoError(t, err)
	require.Positive(t, added.ID)
	got, err := store.GetWhatsOnArticle(ctx, whatson.WhatsOn{ID: added.ID})
	require.NoError(t, err)
	assert.Equal(t, "Quiz", got.Title)
}
