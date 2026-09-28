package programme_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/COMTOP1/AFC-GO/server/internal/programme"
	"github.com/COMTOP1/AFC-GO/server/internal/svcerr"
	"github.com/COMTOP1/AFC-GO/server/internal/upload"
	"github.com/COMTOP1/AFC-GO/server/internal/upload/uploadtest"
)

var ctx = context.Background()

func newService() (*programme.Service, *fakeStore, *uploadtest.Storage) {
	store := newFakeStore()
	store.seasons[1] = programme.Season{ID: 1, Season: "2025-26"}
	store.programmes[1] = programme.Programme{ID: 1, Name: "Opening day", FileName: "programme/opening.pdf",
		DateOfProgramme: time.Date(2025, 8, 9, 0, 0, 0, 0, time.UTC), SeasonID: 1}
	store.programmes[2] = programme.Programme{ID: 2, Name: "Friendly", FileName: "programme/friendly.pdf",
		DateOfProgramme: time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC)}
	objects := uploadtest.New()
	return programme.NewService(store, upload.New(objects)), store, objects
}

func pdf() *upload.File { return uploadtest.File("p.pdf", "application/pdf", "PDF") }

func kindOf(t *testing.T, err error) svcerr.Kind {
	t.Helper()
	se, ok := svcerr.As(err)
	require.True(t, ok, "want svcerr, got %v", err)
	return se.Kind
}

func TestListAttachesSeasons(t *testing.T) {
	svc, _, _ := newService()
	all, err := svc.List(ctx, 0)
	require.NoError(t, err)
	require.Len(t, all, 2)
	require.NotNil(t, all[0].Season)
	assert.Equal(t, "2025-26", all[0].Season.Name)
	assert.Nil(t, all[1].Season, "programme with no season")

	one, err := svc.List(ctx, 1)
	require.NoError(t, err)
	assert.Len(t, one, 1)

	_, err = svc.List(ctx, 99)
	assert.Equal(t, svcerr.KindNotFound, kindOf(t, err))
}

func TestCreateValidates(t *testing.T) {
	svc, _, _ := newService()
	date := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	for name, tc := range map[string]struct {
		in    programme.CreateInput
		file  *upload.File
		field string
	}{
		"name":   {programme.CreateInput{Date: date}, pdf(), "name"},
		"date":   {programme.CreateInput{Name: "x"}, pdf(), "date"},
		"file":   {programme.CreateInput{Name: "x", Date: date}, nil, "file"},
		"season": {programme.CreateInput{Name: "x", Date: date, SeasonID: 99}, pdf(), "seasonId"},
	} {
		_, err := svc.Create(ctx, tc.in, tc.file)
		se, ok := svcerr.As(err)
		require.True(t, ok, name)
		assert.Contains(t, se.Fields, tc.field, name)
	}
}

func TestCreateAndDelete(t *testing.T) {
	svc, store, objects := newService()
	p, err := svc.Create(ctx, programme.CreateInput{Name: "Cup tie", Date: time.Now(), SeasonID: 1}, pdf())
	require.NoError(t, err)
	assert.Equal(t, 1, store.programmes[p.ID].SeasonID)

	deleted, err := svc.Delete(ctx, 1)
	require.NoError(t, err)
	assert.Equal(t, "Opening day", deleted.Name)
	assert.Equal(t, []string{"programme/opening.pdf"}, objects.Deleted)
}

func TestSeasonLifecycle(t *testing.T) {
	svc, store, _ := newService()
	_, err := svc.CreateSeason(ctx, "  ")
	assert.Equal(t, svcerr.KindInvalid, kindOf(t, err))

	s, err := svc.CreateSeason(ctx, "2026-27")
	require.NoError(t, err)
	renamed, err := svc.RenameSeason(ctx, s.ID, "2026/27")
	require.NoError(t, err)
	assert.Equal(t, "2026/27", renamed.Name)

	_, err = svc.DeleteSeason(ctx, 1)
	require.NoError(t, err)
	assert.Zero(t, store.programmes[1].SeasonID, "programmes are unlinked, not deleted")
	_, stillThere := store.seasons[1]
	assert.False(t, stillThere)

	_, err = svc.RenameSeason(ctx, 1, "gone")
	assert.Equal(t, svcerr.KindNotFound, kindOf(t, err))
}
