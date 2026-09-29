package news_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/server/internal/news"
	"github.com/COMTOP1/AFC-GO/server/internal/testdb"
)

func TestAddNewsReturnsID(t *testing.T) {
	db, _ := testdb.Open(t)
	store := news.NewNewsRepo(db)
	ctx := context.Background()

	added, err := store.AddNews(ctx, news.News{Title: "Fresh", Content: null.StringFrom("<p>x</p>")})
	require.NoError(t, err)
	require.Positive(t, added.ID)

	got, err := store.GetNewsArticle(ctx, news.News{ID: added.ID})
	require.NoError(t, err)
	assert.Equal(t, "Fresh", got.Title)
}

// TestEditNewsClearsFileName pins Review Focus #3 at the SQL level: before
// this change EditNews ignored an invalid FileName, so "remove image" left
// the dead key in the row.
func TestEditNewsClearsFileName(t *testing.T) {
	db, _ := testdb.Open(t)
	store := news.NewNewsRepo(db)
	ctx := context.Background()

	n, err := store.GetNewsArticle(ctx, news.News{ID: 1})
	require.NoError(t, err)
	require.True(t, n.FileName.Valid, "fixture row 1 has an image")

	n.FileName = null.String{}
	_, err = store.EditNews(ctx, n)
	require.NoError(t, err)

	got, err := store.GetNewsArticle(ctx, news.News{ID: 1})
	require.NoError(t, err)
	assert.False(t, got.FileName.Valid)
}
