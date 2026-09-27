package news_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/server/internal/news"
	"github.com/COMTOP1/AFC-GO/server/internal/svcerr"
	"github.com/COMTOP1/AFC-GO/server/internal/upload"
	"github.com/COMTOP1/AFC-GO/server/internal/upload/uploadtest"
)

var ctx = context.Background()

func seeded() news.News {
	return news.News{
		ID: 1, Title: "Opener", Content: null.StringFrom("<p>hi</p>"),
		FileName: null.StringFrom("news/old.jpg"), Date: time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC),
	}
}

func newService(rows ...news.News) (*news.Service, *fakeStore, *uploadtest.Storage) {
	store := newFakeStore(rows...)
	objects := uploadtest.New()
	objects.Objects["news/old.jpg"] = "OLD"
	return news.NewService(store, upload.New(objects)), store, objects
}

func kind(t *testing.T, err error) svcerr.Kind {
	t.Helper()
	se, ok := svcerr.As(err)
	require.True(t, ok, "want svcerr, got %v", err)
	return se.Kind
}

func TestCreateSanitisesAndStoresImage(t *testing.T) {
	svc, store, objects := newService()
	a, err := svc.Create(ctx, news.CreateInput{Title: "  Win  ", Content: `<p>ok</p><script>x()</script>`},
		uploadtest.File("a.png", "image/png", "IMG"))
	require.NoError(t, err)

	assert.Positive(t, a.ID)
	assert.Equal(t, "Win", a.Title)
	assert.Equal(t, "<p>ok</p>", a.Content)
	row := store.row(a.ID)
	require.True(t, row.FileName.Valid)
	assert.Equal(t, "IMG", objects.Objects[row.FileName.String])
	assert.Equal(t, "https://cdn.test/"+row.FileName.String, a.ImageURL)
}

func TestCreateRequiresTitle(t *testing.T) {
	svc, _, _ := newService()
	_, err := svc.Create(ctx, news.CreateInput{Title: "   "}, nil)
	assert.Equal(t, svcerr.KindInvalid, kind(t, err))
}

func TestCreateRejectsHTMLUpload(t *testing.T) {
	svc, store, objects := newService()
	_, err := svc.Create(ctx, news.CreateInput{Title: "x"}, uploadtest.File("x.html", "text/html", "<script>"))
	se, _ := svcerr.As(err)
	require.NotNil(t, se)
	assert.Contains(t, se.Fields, "file")
	assert.Len(t, objects.Objects, 1, "only the pre-existing object")
	assert.Empty(t, store.rows)
}

func TestUpdateOmittedFieldsUnchanged(t *testing.T) {
	svc, store, objects := newService(seeded())
	title := "Renamed"
	_, err := svc.Update(ctx, 1, news.UpdateInput{Title: &title}, nil)
	require.NoError(t, err)

	row := store.row(1)
	assert.Equal(t, "Renamed", row.Title)
	assert.Equal(t, "<p>hi</p>", row.Content.String, "content was not sent, so it must be untouched")
	assert.Equal(t, "news/old.jpg", row.FileName.String)
	assert.Empty(t, objects.Deleted)
}

func TestUpdateEmptyContentClearsIt(t *testing.T) {
	svc, store, _ := newService(seeded())
	empty := ""
	_, err := svc.Update(ctx, 1, news.UpdateInput{Content: &empty}, nil)
	require.NoError(t, err)
	assert.False(t, store.row(1).Content.Valid)
}

func TestUpdateRemoveImage(t *testing.T) {
	svc, store, objects := newService(seeded())
	a, err := svc.Update(ctx, 1, news.UpdateInput{RemoveImage: true}, nil)
	require.NoError(t, err)
	assert.False(t, store.row(1).FileName.Valid)
	assert.Empty(t, a.ImageURL)
	assert.Equal(t, []string{"news/old.jpg"}, objects.Deleted)
}

func TestUpdateReplaceImage(t *testing.T) {
	svc, store, objects := newService(seeded())
	_, err := svc.Update(ctx, 1, news.UpdateInput{}, uploadtest.File("n.webp", "image/webp", "NEW"))
	require.NoError(t, err)
	key := store.row(1).FileName.String
	assert.NotEqual(t, "news/old.jpg", key)
	assert.Equal(t, "NEW", objects.Objects[key])
	assert.Equal(t, []string{"news/old.jpg"}, objects.Deleted)
}

func TestUpdateDBFailureKeepsOldImage(t *testing.T) {
	svc, store, objects := newService(seeded())
	store.editErr = errors.New("db down")
	_, err := svc.Update(ctx, 1, news.UpdateInput{}, uploadtest.File("n.png", "image/png", "NEW"))
	require.Error(t, err)
	assert.Equal(t, "OLD", objects.Objects["news/old.jpg"], "old image must survive a failed edit")
	assert.Len(t, objects.Objects, 1, "the new upload is cleaned up")
}

func TestGetMissingIsNotFound(t *testing.T) {
	svc, _, _ := newService()
	_, err := svc.Get(ctx, 42)
	assert.Equal(t, svcerr.KindNotFound, kind(t, err))
}

func TestDeleteRemovesImage(t *testing.T) {
	svc, store, objects := newService(seeded())
	a, err := svc.Delete(ctx, 1)
	require.NoError(t, err)
	assert.Equal(t, "Opener", a.Title)
	assert.Empty(t, store.rows)
	assert.Equal(t, []string{"news/old.jpg"}, objects.Deleted)
}

func TestLatest(t *testing.T) {
	svc, _, _ := newService()
	_, ok, err := svc.Latest(ctx)
	require.NoError(t, err)
	assert.False(t, ok)

	svc, _, _ = newService(seeded())
	a, ok, err := svc.Latest(ctx)
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, 1, a.ID)
}
