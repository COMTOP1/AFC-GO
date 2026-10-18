package files_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/server/internal/affiliation"
	"github.com/COMTOP1/AFC-GO/server/internal/document"
	"github.com/COMTOP1/AFC-GO/server/internal/files"
	"github.com/COMTOP1/AFC-GO/server/internal/image"
	"github.com/COMTOP1/AFC-GO/server/internal/news"
	"github.com/COMTOP1/AFC-GO/server/internal/programme"
	"github.com/COMTOP1/AFC-GO/server/internal/sponsor"
	"github.com/COMTOP1/AFC-GO/server/internal/svcerr"
	"github.com/COMTOP1/AFC-GO/server/internal/team"
	"github.com/COMTOP1/AFC-GO/server/internal/user"
	"github.com/COMTOP1/AFC-GO/server/internal/whatson"
)

// The fakes below are small map-backed stand-ins for each domain store's
// single-row getter, so TestResolversMapEveryKind exercises the real
// files.Resolvers() wiring (field mapping, null.String vs string FileName,
// and the "not found" translation) rather than hand-rolled Resolver stubs.

type fakeAffiliations map[int]affiliation.Affiliation

func (f fakeAffiliations) GetAffiliation(_ context.Context, a affiliation.Affiliation) (affiliation.Affiliation, error) {
	if x, ok := f[a.ID]; ok {
		return x, nil
	}
	return affiliation.Affiliation{}, fmt.Errorf("failed to get affiliation: %w", sql.ErrNoRows)
}

type fakeDocuments map[int]document.Document

func (f fakeDocuments) GetDocument(_ context.Context, d document.Document) (document.Document, error) {
	if x, ok := f[d.ID]; ok {
		return x, nil
	}
	return document.Document{}, fmt.Errorf("failed to get document: %w", sql.ErrNoRows)
}

type fakeImages map[int]image.Image

func (f fakeImages) GetImage(_ context.Context, i image.Image) (image.Image, error) {
	if x, ok := f[i.ID]; ok {
		return x, nil
	}
	return image.Image{}, fmt.Errorf("failed to get image: %w", sql.ErrNoRows)
}

type fakeNews map[int]news.News

func (f fakeNews) GetNewsArticle(_ context.Context, n news.News) (news.News, error) {
	if x, ok := f[n.ID]; ok {
		return x, nil
	}
	return news.News{}, fmt.Errorf("failed to get news article: %w", sql.ErrNoRows)
}

type fakeProgrammes map[int]programme.Programme

func (f fakeProgrammes) GetProgramme(_ context.Context, p programme.Programme) (programme.Programme, error) {
	if x, ok := f[p.ID]; ok {
		return x, nil
	}
	return programme.Programme{}, fmt.Errorf("failed to get programme: %w", sql.ErrNoRows)
}

type fakeSponsors map[int]sponsor.Sponsor

func (f fakeSponsors) GetSponsor(_ context.Context, s sponsor.Sponsor) (sponsor.Sponsor, error) {
	if x, ok := f[s.ID]; ok {
		return x, nil
	}
	return sponsor.Sponsor{}, fmt.Errorf("failed to get sponsor: %w", sql.ErrNoRows)
}

type fakeTeams map[int]team.Team

func (f fakeTeams) GetTeam(_ context.Context, t team.Team) (team.Team, error) {
	if x, ok := f[t.ID]; ok {
		return x, nil
	}
	return team.Team{}, fmt.Errorf("failed to get team: %w", sql.ErrNoRows)
}

type fakeUsers map[string]user.User

func (f fakeUsers) GetUserByUUID(_ context.Context, id string) (user.User, error) {
	if u, ok := f[id]; ok {
		return u, nil
	}
	return user.User{}, fmt.Errorf("failed to get user: %w", sql.ErrNoRows)
}

type fakeWhatsOn map[int]whatson.WhatsOn

func (f fakeWhatsOn) GetWhatsOnArticle(_ context.Context, w whatson.WhatsOn) (whatson.WhatsOn, error) {
	if x, ok := f[w.ID]; ok {
		return x, nil
	}
	return whatson.WhatsOn{}, fmt.Errorf("failed to get whats on event: %w", sql.ErrNoRows)
}

// fakePlayerPhotos stands in for player.Service, which returns NotFound for
// missing, youth-team and under-18 players (tested in the player package
// itself; here we only need Resolvers() to delegate to it untouched).
type fakePlayerPhotos map[string]string

func (f fakePlayerPhotos) PhotoKey(_ context.Context, id string) (string, error) {
	if key, ok := f[id]; ok {
		return key, nil
	}
	return "", svcerr.NotFound("player photo not found", nil)
}

const (
	oneUUID   = "00000000-0000-4000-8000-000000000001"
	otherUUID = "00000000-0000-4000-8000-000000000099"
)

// TestResolversMapEveryKind builds files.Resolvers() from Sources backed by
// small in-test fakes for all ten kinds, so a broken field mapping (e.g. the
// wrong FileName field, or a null.String read without .String) or a broken
// "user" ID guard fails here instead of passing silently.
func TestResolversMapEveryKind(t *testing.T) {
	resolvers := files.Resolvers(files.Sources{
		Affiliation: fakeAffiliations{1: {ID: 1, FileName: null.StringFrom("affiliation/one.png")}},
		Document:    fakeDocuments{1: {ID: 1, FileName: "document/one.pdf"}},
		Image:       fakeImages{1: {ID: 1, FileName: "gallery/one.jpg"}},
		News:        fakeNews{1: {ID: 1, FileName: null.StringFrom("news/one.jpg")}},
		Programme:   fakeProgrammes{1: {ID: 1, FileName: "programme/one.pdf"}},
		Sponsor:     fakeSponsors{1: {ID: 1, FileName: null.StringFrom("sponsor/one.png")}},
		Team:        fakeTeams{1: {ID: 1, FileName: null.StringFrom("team/one.png")}},
		User:        fakeUsers{oneUUID: {ID: 1, FileName: null.StringFrom("user/one.png")}},
		WhatsOn:     fakeWhatsOn{1: {ID: 1, FileName: null.StringFrom("whatson/one.jpg")}},
		Players:     fakePlayerPhotos{oneUUID: "player/one.png"},
	})

	cases := []struct {
		kind    string
		id      string
		missing string
		wantKey string
	}{
		{"affiliation", "1", "99", "affiliation/one.png"},
		{"document", "1", "99", "document/one.pdf"},
		{"gallery", "1", "99", "gallery/one.jpg"},
		{"news", "1", "99", "news/one.jpg"},
		{"programme", "1", "99", "programme/one.pdf"},
		{"sponsor", "1", "99", "sponsor/one.png"},
		{"team", "1", "99", "team/one.png"},
		{"user", oneUUID, otherUUID, "user/one.png"},
		{"player", oneUUID, otherUUID, "player/one.png"},
		{"whatson", "1", "99", "whatson/one.jpg"},
	}
	for _, tc := range cases {
		t.Run(tc.kind+"/found", func(t *testing.T) {
			key, err := resolvers[tc.kind](ctx, tc.id)
			require.NoError(t, err)
			assert.Equal(t, tc.wantKey, key)
		})
		t.Run(tc.kind+"/missing row", func(t *testing.T) {
			_, err := resolvers[tc.kind](ctx, tc.missing)
			se, ok := svcerr.As(err)
			require.True(t, ok, "want svcerr, got %v", err)
			assert.Equal(t, svcerr.KindNotFound, se.Kind)
		})
	}

	for _, kind := range []string{"user", "player"} {
		t.Run(kind+"/integer id is not found", func(t *testing.T) {
			_, err := resolvers[kind](ctx, "1")
			se, ok := svcerr.As(err)
			require.True(t, ok, "want svcerr, got %v", err)
			assert.Equal(t, svcerr.KindNotFound, se.Kind)
		})
	}
}
