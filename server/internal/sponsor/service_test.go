package sponsor_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/server/internal/sponsor"
	"github.com/COMTOP1/AFC-GO/server/internal/svcerr"
	"github.com/COMTOP1/AFC-GO/server/internal/upload"
	"github.com/COMTOP1/AFC-GO/server/internal/upload/uploadtest"
)

var ctx = context.Background()

func clubSponsor() sponsor.Sponsor {
	return sponsor.Sponsor{ID: 1, Name: "Club Sponsor", Website: null.StringFrom("https://s.example.test"),
		FileName: null.StringFrom("sponsor/club.png"), TeamID: "A"}
}

func newService(rows ...sponsor.Sponsor) (*sponsor.Service, *fakeStore, *uploadtest.Storage) {
	store := newFakeStore(rows...)
	objects := uploadtest.New()
	teams := fakeTeams{1: {ID: 1, Name: "First Team"}}
	return sponsor.NewService(store, teams, upload.New(objects)), store, objects
}

func png() *upload.File { return uploadtest.File("s.png", "image/png", "PNG") }

func fieldsOf(t *testing.T, err error) map[string]string {
	t.Helper()
	se, ok := svcerr.As(err)
	require.True(t, ok, "want svcerr, got %v", err)
	return se.Fields
}

func TestCreateValidates(t *testing.T) {
	svc, _, objects := newService()
	cases := map[string]struct {
		in    sponsor.CreateInput
		image *upload.File
		field string
	}{
		"name required":   {sponsor.CreateInput{Team: "A"}, png(), "name"},
		"image required":  {sponsor.CreateInput{Name: "x", Team: "A"}, nil, "image"},
		"bad website":     {sponsor.CreateInput{Name: "x", Website: "not a url", Team: "A"}, png(), "website"},
		"bad team code":   {sponsor.CreateInput{Name: "x", Team: "Z"}, png(), "team"},
		"unknown team id": {sponsor.CreateInput{Name: "x", Team: "42"}, png(), "team"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := svc.Create(ctx, tc.in, tc.image)
			assert.Contains(t, fieldsOf(t, err), tc.field)
		})
	}
	assert.Empty(t, objects.Objects, "nothing uploaded when validation fails")
}

func TestCreate(t *testing.T) {
	svc, store, _ := newService()
	for _, teamValue := range []string{"A", "O", "Y", "1", ""} {
		s, err := svc.Create(ctx, sponsor.CreateInput{Name: "Kit Co", Website: "https://kit.example.test", Team: teamValue}, png())
		require.NoError(t, err, teamValue)
		assert.Equal(t, teamValue, store.rows[s.ID].TeamID)
		assert.NotEmpty(t, s.ImageURL)
	}
}

func TestForTeam(t *testing.T) {
	team1 := clubSponsor()
	team1.ID, team1.TeamID = 2, "1"
	svc, _, _ := newService(clubSponsor(), team1)
	got, err := svc.ForTeam(ctx, 1)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, 2, got[0].ID)
}

func TestDelete(t *testing.T) {
	svc, store, objects := newService(clubSponsor())
	s, err := svc.Delete(ctx, 1)
	require.NoError(t, err)
	assert.Equal(t, "Club Sponsor", s.Name)
	assert.Empty(t, store.rows)
	assert.Equal(t, []string{"sponsor/club.png"}, objects.Deleted)

	_, err = svc.Delete(ctx, 1)
	se, _ := svcerr.As(err)
	require.NotNil(t, se)
	assert.Equal(t, svcerr.KindNotFound, se.Kind)
}
