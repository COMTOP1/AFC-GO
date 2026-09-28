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

// TestListHidesPhotoOnUnknownTeam pins Review Focus #1: a player whose team
// can't be found (e.g. team_id 0 after DetachTeam) must not leak a photo via
// List just because "not found" was treated as "not youth".
func TestListHidesPhotoOnUnknownTeam(t *testing.T) {
	store := newFakeStore(player.Player{
		ID: 4, Name: "Orphan", FileName: null.StringFrom("player/orphan.png"), DateOfBirth: yearsAgo(30), TeamID: 7,
	})
	objects := uploadtest.New()
	objects.Objects["player/orphan.png"] = "IMG"
	svc := player.NewService(store, teams, upload.New(objects))

	list, err := svc.List(ctx)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Empty(t, list[0].ImageURL, "team not found: fail closed")
	assert.Nil(t, list[0].Team)
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

// TestSquadIgnoresCallerIsYouth pins Review Focus #1: Squad must not trust
// the caller's team.Team.IsYouth; it re-reads the team itself.
func TestSquadIgnoresCallerIsYouth(t *testing.T) {
	store := newFakeStore(player.Player{
		ID: 5, Name: "Adult on Youth Team", FileName: null.StringFrom("player/adult2.png"), DateOfBirth: yearsAgo(30), TeamID: 2,
	})
	objects := uploadtest.New()
	objects.Objects["player/adult2.png"] = "IMG"
	svc := player.NewService(store, teams, upload.New(objects))

	// Team 2 is youth, but the caller (deliberately, or by a stale value)
	// says IsYouth: false.
	members, err := svc.Squad(ctx, team.Team{ID: 2, IsYouth: false})
	require.NoError(t, err)
	require.Len(t, members, 1)
	assert.Empty(t, members[0].ImageURL, "team 2 is youth regardless of what the caller claims")
}

// TestSquadHidesPhotosWhenTeamLookupFails pins Review Focus #1: if Squad
// can't re-read the team, it must fail closed rather than show photos.
func TestSquadHidesPhotosWhenTeamLookupFails(t *testing.T) {
	store := newFakeStore(player.Player{
		ID: 6, Name: "Adult Unknown Team", FileName: null.StringFrom("player/adult3.png"), DateOfBirth: yearsAgo(30), TeamID: 7,
	})
	objects := uploadtest.New()
	objects.Objects["player/adult3.png"] = "IMG"
	svc := player.NewService(store, teams, upload.New(objects))

	members, err := svc.Squad(ctx, team.Team{ID: 7, IsYouth: false})
	require.NoError(t, err)
	require.Len(t, members, 1, "players are still listed")
	assert.Empty(t, members[0].ImageURL, "team lookup failed: fail closed")
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
		assert.Equal(t, "player photo not found", se.Message, "missing and hidden players give the same error, id %d", id)
	}

	// An adult with a photo on an unknown team must also be NotFound (fail
	// closed), via the identical error a missing player would give.
	unknownTeamStore := newFakeStore(player.Player{
		ID: 4, Name: "Orphan", FileName: null.StringFrom("player/orphan.png"), DateOfBirth: yearsAgo(30), TeamID: 7,
	})
	objects := uploadtest.New()
	objects.Objects["player/orphan.png"] = "IMG"
	orphanSvc := player.NewService(unknownTeamStore, teams, upload.New(objects))
	_, err = orphanSvc.PhotoKey(ctx, 4)
	se, ok := svcerr.As(err)
	require.True(t, ok)
	assert.Equal(t, svcerr.KindNotFound, se.Kind)
	assert.Equal(t, "player photo not found", se.Message)
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
