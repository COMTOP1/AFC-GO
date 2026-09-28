package team_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/server/internal/svcerr"
	"github.com/COMTOP1/AFC-GO/server/internal/team"
	"github.com/COMTOP1/AFC-GO/server/internal/upload"
	"github.com/COMTOP1/AFC-GO/server/internal/upload/uploadtest"
)

var ctx = context.Background()

func firstTeam() team.Team {
	return team.Team{ID: 1, Name: "First Team", League: null.StringFrom("Hellenic"),
		FileName: null.StringFrom("team/first.jpg"), IsActive: true, Ages: 99}
}

func oldTeam() team.Team {
	return team.Team{ID: 3, Name: "Old Team", IsActive: false, Ages: 99}
}

func newService(detachers ...team.Detacher) (*team.Service, *fakeStore, *uploadtest.Storage) {
	store := newFakeStore(firstTeam(), oldTeam())
	objects := uploadtest.New()
	objects.Objects["team/first.jpg"] = "IMG"
	return team.NewService(store, upload.New(objects), detachers...), store, objects
}

func fields(t *testing.T, err error) map[string]string {
	t.Helper()
	se, ok := svcerr.As(err)
	require.True(t, ok, "want svcerr, got %v", err)
	return se.Fields
}

func TestListHidesInactiveFromPublic(t *testing.T) {
	svc, _, _ := newService()
	public, err := svc.List(ctx, false)
	require.NoError(t, err)
	assert.Len(t, public, 1)
	all, err := svc.List(ctx, true)
	require.NoError(t, err)
	assert.Len(t, all, 2)
}

func TestCreateValidates(t *testing.T) {
	svc, _, _ := newService()
	assert.Contains(t, fields(t, mustErr(svc.Create(ctx, team.CreateInput{Ages: 99}, nil))), "name")
	assert.Contains(t, fields(t, mustErr(svc.Create(ctx, team.CreateInput{Name: "x", LeagueTable: "table"}, nil))), "leagueTable")
	assert.Contains(t, fields(t, mustErr(svc.Create(ctx, team.CreateInput{Name: "x", Fixtures: "fixtures"}, nil))), "fixtures")
	assert.Contains(t, fields(t, mustErr(svc.Create(ctx, team.CreateInput{Name: "x", Ages: -1}, nil))), "ages")
}

func mustErr(_ team.Public, err error) error { return err }

func TestCreateYouthRuleAndFlags(t *testing.T) {
	svc, store, _ := newService()
	u12, err := svc.Create(ctx, team.CreateInput{Name: "U12", Ages: 12}, nil)
	require.NoError(t, err)
	assert.True(t, u12.IsYouth, "ages < 19 is always youth")
	assert.False(t, u12.IsActive)

	// Legacy quirk fixed: ticking youth used to set isActive instead.
	senior, err := svc.Create(ctx, team.CreateInput{Name: "Vets", Ages: 99, IsYouth: true}, nil)
	require.NoError(t, err)
	assert.True(t, store.row(senior.ID).IsYouth)
	assert.False(t, store.row(senior.ID).IsActive)
}

func TestUpdateCanClearFlags(t *testing.T) {
	svc, store, _ := newService()
	no := false
	_, err := svc.Update(ctx, 1, team.UpdateInput{IsActive: &no}, nil)
	require.NoError(t, err)
	assert.False(t, store.row(1).IsActive, "legacy edit could never untick isActive")
	assert.Equal(t, "Hellenic", store.row(1).League.String, "omitted fields unchanged")
}

func TestUpdateAgesForcesYouth(t *testing.T) {
	svc, store, _ := newService()
	ages := 16
	_, err := svc.Update(ctx, 1, team.UpdateInput{Ages: &ages}, nil)
	require.NoError(t, err)
	assert.True(t, store.row(1).IsYouth)
}

func TestUpdateRemoveImage(t *testing.T) {
	svc, store, objects := newService()
	_, err := svc.Update(ctx, 1, team.UpdateInput{RemoveImage: true}, nil)
	require.NoError(t, err)
	assert.False(t, store.row(1).FileName.Valid)
	assert.Equal(t, []string{"team/first.jpg"}, objects.Deleted)
}

func TestDeleteDetachesThenDeletes(t *testing.T) {
	var calls []string
	svc, store, objects := newService(
		recordingDetacher{name: "players", calls: &calls},
		recordingDetacher{name: "sponsors", calls: &calls},
		recordingDetacher{name: "users", calls: &calls},
	)
	deleted, err := svc.Delete(ctx, 1)
	require.NoError(t, err)
	assert.Equal(t, "First Team", deleted.Name)
	assert.Equal(t, []string{"players:1", "sponsors:1", "users:1"}, calls)
	_, still := store.rows[1]
	assert.False(t, still)
	assert.Equal(t, []string{"team/first.jpg"}, objects.Deleted)
}

func TestDeleteStopsWhenDetachFails(t *testing.T) {
	var calls []string
	svc, store, _ := newService(recordingDetacher{name: "players", calls: &calls, failOn: 1})
	_, err := svc.Delete(ctx, 1)
	require.Error(t, err)
	_, still := store.rows[1]
	assert.True(t, still, "team must not be deleted if its players could not be detached")
}
