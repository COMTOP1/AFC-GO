package site_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/server/internal/affiliation"
	"github.com/COMTOP1/AFC-GO/server/internal/auth/authtest"
	"github.com/COMTOP1/AFC-GO/server/internal/news"
	"github.com/COMTOP1/AFC-GO/server/internal/player"
	"github.com/COMTOP1/AFC-GO/server/internal/role"
	"github.com/COMTOP1/AFC-GO/server/internal/site"
	"github.com/COMTOP1/AFC-GO/server/internal/sponsor"
	"github.com/COMTOP1/AFC-GO/server/internal/svcerr"
	"github.com/COMTOP1/AFC-GO/server/internal/team"
	"github.com/COMTOP1/AFC-GO/server/internal/upload"
	"github.com/COMTOP1/AFC-GO/server/internal/upload/uploadtest"
	"github.com/COMTOP1/AFC-GO/server/internal/user"
	"github.com/COMTOP1/AFC-GO/server/internal/web"
	"github.com/COMTOP1/AFC-GO/server/internal/web/apitest"
	"github.com/COMTOP1/AFC-GO/server/internal/whatson"
)

var ctx = context.Background()

type fakes struct {
	newsErr      error
	sponsorsErr  error
	squadForTeam map[int][]player.Member
}

func (f *fakes) Latest(context.Context) (news.Article, bool, error) {
	return news.Article{ID: 1, Title: "Opener"}, true, f.newsErr
}
func (f *fakes) Next(context.Context) (whatson.Event, bool, error) {
	return whatson.Event{}, false, nil
}
func (f *fakes) ListMinimal(context.Context) ([]sponsor.Public, error) {
	return []sponsor.Public{{ID: 1, Name: "Club Sponsor"}}, f.sponsorsErr
}
func (f *fakes) ForTeam(_ context.Context, _ int) ([]sponsor.Public, error) {
	return []sponsor.Public{{ID: 2, Name: "Team Sponsor", Team: "1"}}, nil
}

type affiliations struct{}

func (affiliations) ListMinimal(context.Context) ([]affiliation.Public, error) {
	return []affiliation.Public{{ID: 1, Name: "County FA"}}, nil
}

type teams struct{}

func (teams) List(_ context.Context, _ bool) ([]team.Public, error) {
	return []team.Public{{ID: 1, Name: "First Team", IsActive: true}}, nil
}

func (teams) Get(_ context.Context, id int) (team.Public, error) {
	switch id {
	case 1:
		return team.Public{ID: 1, Name: "First Team", IsActive: true}, nil
	case 2:
		return team.Public{ID: 2, Name: "Under 12s", IsYouth: true}, nil
	}
	return team.Public{}, svcerr.NotFound("team not found", nil)
}

func (f *fakes) Squad(_ context.Context, t team.Team) ([]player.Member, error) {
	return f.squadForTeam[t.ID], nil
}

type settings struct{}

func (settings) DisplayEmail(context.Context) (string, error) { return "hello@example.test", nil }

type users struct{}

func (users) GetUsersContact(context.Context) ([]user.User, error) {
	return []user.User{{ID: 3, Name: "Club Secretary", Email: "secretary@example.test", Role: role.ClubSecretary,
		FileName: null.StringFrom("user/sec.png"), Hash: null.StringFrom("secret-hash")}}, nil
}

func (users) GetUsersManagersTeam(_ context.Context, t team.Team) ([]user.User, error) {
	if t.ID == 1 {
		return []user.User{{ID: 2, Name: "Team Manager", Email: "manager@example.test", Role: role.Manager}}, nil
	}
	return nil, nil
}

type visitors struct{}

func (visitors) Count() int { return 42 }

func newService(f *fakes) *site.Service {
	return site.NewService(site.Deps{
		News: f, WhatsOn: f, Sponsors: f, Affiliations: affiliations{}, Teams: teams{}, Players: f,
		Settings: settings{}, Users: users{}, Visitors: visitors{},
		Files: upload.New(uploadtest.New()), Version: "test",
	})
}

func TestSite(t *testing.T) {
	info, err := newService(&fakes{}).Site(ctx)
	require.NoError(t, err)
	assert.Equal(t, 42, info.VisitorCount)
	assert.Equal(t, "hello@example.test", info.DisplayEmail)
	assert.Len(t, info.Teams, 1)
	assert.Positive(t, info.Year)
}

func TestHomeIsFailSoft(t *testing.T) {
	home := newService(&fakes{sponsorsErr: errors.New("db down")}).Home(ctx)
	require.NotNil(t, home.LatestNews)
	assert.Equal(t, "Opener", home.LatestNews.Title)
	assert.Nil(t, home.NextEvent)
	assert.Empty(t, home.Sponsors, "failed panel is omitted")
	assert.Len(t, home.Affiliations, 1)

	home = newService(&fakes{newsErr: errors.New("db down")}).Home(ctx)
	assert.Nil(t, home.LatestNews)
}

func TestContactNeverLeaksSecrets(t *testing.T) {
	c, err := newService(&fakes{}).Contact(ctx)
	require.NoError(t, err)
	require.Len(t, c.People, 1)
	assert.Equal(t, "Club Secretary", c.People[0].Role)
	assert.Equal(t, "https://cdn.test/user/sec.png", c.People[0].ImageURL)
}

// TestTeamDetailHidesMinorPhotos pins Review Focus #1 for /teams/{id}: the
// squad is built by player.Service.Squad, which applies PhotoVisible with the
// team's youth flag. This test proves site passes that flag through.
func TestTeamDetailHidesMinorPhotos(t *testing.T) {
	f := &fakes{squadForTeam: map[int][]player.Member{2: {{ID: 9, Name: "Youth"}}}}
	var seen []team.Team
	svc := site.NewService(site.Deps{
		News: f, WhatsOn: f, Sponsors: f, Affiliations: affiliations{}, Teams: teams{},
		Players: squadSpy{f: f, seen: &seen}, Settings: settings{}, Users: users{}, Visitors: visitors{},
		Files: upload.New(uploadtest.New()),
	})
	detail, err := svc.Team(ctx, 2)
	require.NoError(t, err)
	require.Len(t, seen, 1)
	assert.True(t, seen[0].IsYouth, "youth flag must reach player.Squad")
	assert.Len(t, detail.Players, 1)
}

type squadSpy struct {
	f    *fakes
	seen *[]team.Team
}

func (s squadSpy) Squad(ctx context.Context, t team.Team) ([]player.Member, error) {
	*s.seen = append(*s.seen, t)
	return s.f.Squad(ctx, t)
}

func TestTeamDetail(t *testing.T) {
	detail, err := newService(&fakes{}).Team(ctx, 1)
	require.NoError(t, err)
	assert.Equal(t, "First Team", detail.Team.Name)
	require.Len(t, detail.Managers, 1)
	assert.Equal(t, "manager@example.test", detail.Managers[0].Email)
	assert.Len(t, detail.Sponsors, 1)

	_, err = newService(&fakes{}).Team(ctx, 99)
	se, ok := svcerr.As(err)
	require.True(t, ok)
	assert.Equal(t, svcerr.KindNotFound, se.Kind)
}

func TestSiteRoutes(t *testing.T) {
	e := apitest.NewEcho()
	site.NewHandlers(newService(&fakes{})).Register(web.NewAPI(e, false), authtest.Everyone().Guards())
	c := apitest.New(e)
	for _, path := range []string{"/api/v1/site", "/api/v1/home", "/api/v1/contact", "/api/v1/teams/1"} {
		rec := c.Get(t, path)
		assert.Equal(t, http.StatusOK, rec.Code, path)
		assert.NotContains(t, rec.Body.String(), "secret-hash", path)
	}
	assert.Equal(t, http.StatusNotFound, c.Get(t, "/api/v1/teams/99").Code)
}
