package player_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/server/internal/player"
	"github.com/COMTOP1/AFC-GO/server/internal/svcerr"
	"github.com/COMTOP1/AFC-GO/server/internal/team"
	"github.com/COMTOP1/AFC-GO/server/internal/upload"
	"github.com/COMTOP1/AFC-GO/server/internal/upload/uploadtest"
)

var ctx = context.Background()

func yearsAgo(n int) null.Time { return null.TimeFrom(time.Now().AddDate(-n, 0, -1)) }

var teams = fakeTeams{
	1: {ID: 1, Name: "First Team", IsActive: true, Ages: 99},
	2: {ID: 2, Name: "Under 12s", IsActive: true, IsYouth: true, Ages: 12},
}

func squad() []player.Player {
	return []player.Player{
		{ID: 1, Name: "Adult", FileName: null.StringFrom("player/adult.png"), DateOfBirth: yearsAgo(30),
			Position: null.StringFrom("Striker"), IsCaptain: true, TeamID: 1},
		{ID: 2, Name: "Youth", FileName: null.StringFrom("player/youth.png"), DateOfBirth: yearsAgo(11), TeamID: 2},
		{ID: 3, Name: "Young Senior", FileName: null.StringFrom("player/young.png"), DateOfBirth: yearsAgo(16), TeamID: 1},
	}
}

func newService() (*player.Service, *fakeStore, *uploadtest.Storage) {
	store := newFakeStore(squad()...)
	objects := uploadtest.New()
	for _, p := range squad() {
		objects.Objects[p.FileName.String] = "IMG"
	}
	return player.NewService(store, teams, upload.New(objects)), store, objects
}

func fieldsOf(t *testing.T, err error) map[string]string {
	t.Helper()
	se, ok := svcerr.As(err)
	require.True(t, ok, "want svcerr, got %v", err)
	return se.Fields
}

// TestListHidesMinorPhotos pins Review Focus #1 for /players.
func TestListHidesMinorPhotos(t *testing.T) {
	svc, _, _ := newService()
	list, err := svc.List(ctx)
	require.NoError(t, err)
	byID := map[int]player.Public{}
	for _, p := range list {
		byID[p.ID] = p
	}
	assert.NotEmpty(t, byID[1].ImageURL)
	assert.Empty(t, byID[2].ImageURL, "youth team")
	assert.Empty(t, byID[3].ImageURL, "under 18")
	require.NotNil(t, byID[1].Team)
	assert.Equal(t, "First Team", byID[1].Team.Name)
	require.NotNil(t, byID[1].Age)
	assert.Equal(t, 30, *byID[1].Age)
}

func TestSquadHidesMinorPhotos(t *testing.T) {
	svc, _, _ := newService()
	members, err := svc.Squad(ctx, team.Team{ID: 1, IsYouth: false})
	require.NoError(t, err)
	require.Len(t, members, 2)
	for _, m := range members {
		if m.ID == 3 {
			assert.Empty(t, m.ImageURL)
		} else {
			assert.NotEmpty(t, m.ImageURL)
		}
	}
}

func TestPhotoKey(t *testing.T) {
	svc, _, _ := newService()
	key, err := svc.PhotoKey(ctx, 1)
	require.NoError(t, err)
	assert.Equal(t, "player/adult.png", key)
	for _, id := range []int{2, 3, 99} {
		_, err = svc.PhotoKey(ctx, id)
		se, ok := svcerr.As(err)
		require.True(t, ok, id)
		assert.Equal(t, svcerr.KindNotFound, se.Kind, id)
	}
}

func TestCreateValidates(t *testing.T) {
	svc, _, _ := newService()
	dob := time.Date(1995, 1, 1, 0, 0, 0, 0, time.UTC)
	assert.Contains(t, fieldsOf(t, errOf(svc.Create(ctx, player.CreateInput{TeamID: 1, DateOfBirth: dob}, nil))), "name")
	assert.Contains(t, fieldsOf(t, errOf(svc.Create(ctx, player.CreateInput{Name: "x", TeamID: 9, DateOfBirth: dob}, nil))), "teamId")
	assert.Contains(t, fieldsOf(t, errOf(svc.Create(ctx, player.CreateInput{Name: "x", TeamID: 1}, nil))), "dateOfBirth")
	assert.Contains(t, fieldsOf(t, errOf(svc.Create(ctx, player.CreateInput{Name: "x", TeamID: 1, DateOfBirth: time.Now().AddDate(0, 0, 1)}, nil))), "dateOfBirth")
}

func errOf(_ player.Public, err error) error { return err }

func TestUpdateCanUncaptainAndRemoveImage(t *testing.T) {
	svc, store, objects := newService()
	no := false
	_, err := svc.Update(ctx, 1, player.UpdateInput{IsCaptain: &no, RemoveImage: true}, nil)
	require.NoError(t, err)
	row := store.row(1)
	assert.False(t, row.IsCaptain, "legacy edit could never un-captain")
	assert.False(t, row.FileName.Valid)
	assert.Equal(t, "Striker", row.Position.String, "omitted fields unchanged")
	assert.Equal(t, []string{"player/adult.png"}, objects.Deleted)
}

func TestDelete(t *testing.T) {
	svc, store, objects := newService()
	deleted, err := svc.Delete(ctx, 2)
	require.NoError(t, err)
	assert.Equal(t, "Youth", deleted.Name)
	assert.Empty(t, deleted.ImageURL, "the deleted record never exposes a minor's photo either")
	_, still := store.rows[2]
	assert.False(t, still)
	assert.Equal(t, []string{"player/youth.png"}, objects.Deleted)
}
