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

type fakeUsers map[int]user.User

func (f fakeUsers) GetUser(_ context.Context, u user.User) (user.User, error) {
	if x, ok := f[u.ID]; ok {
		return x, nil
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
type fakePlayerPhotos map[int]string

func (f fakePlayerPhotos) PhotoKey(_ context.Context, id int) (string, error) {
	if key, ok := f[id]; ok {
		return key, nil
	}
	return "", svcerr.NotFound("player photo not found", nil)
}

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
		User: fakeUsers{
			1: {ID: 1, FileName: null.StringFrom("user/one.png")},
			// GetUser matches "email = ? OR id = ?"; a blank-email row can
			// collide with an unrelated id, so id 2 resolves to someone
			// else's row (ID 5) here to pin that the resolver must reject it.
			2: {ID: 5, FileName: null.StringFrom("user/five.png")},
		},
		WhatsOn: fakeWhatsOn{1: {ID: 1, FileName: null.StringFrom("whatson/one.jpg")}},
		Players: fakePlayerPhotos{1: "player/one.png"},
	})

	cases := []struct {
		kind    string
		wantKey string
	}{
		{"affiliation", "affiliation/one.png"},
		{"document", "document/one.pdf"},
		{"gallery", "gallery/one.jpg"},
		{"news", "news/one.jpg"},
		{"programme", "programme/one.pdf"},
		{"sponsor", "sponsor/one.png"},
		{"team", "team/one.png"},
		{"user", "user/one.png"},
		{"whatson", "whatson/one.jpg"},
	}
	for _, tc := range cases {
		t.Run(tc.kind+"/found", func(t *testing.T) {
			key, err := resolvers[tc.kind](ctx, 1)
			require.NoError(t, err)
			assert.Equal(t, tc.wantKey, key)
		})
		t.Run(tc.kind+"/missing row", func(t *testing.T) {
			_, err := resolvers[tc.kind](ctx, 99)
			se, ok := svcerr.As(err)
			require.True(t, ok, "want svcerr, got %v", err)
			assert.Equal(t, svcerr.KindNotFound, se.Kind)
		})
	}

	t.Run("user/id-email collision is rejected", func(t *testing.T) {
		_, err := resolvers["user"](ctx, 2)
		se, ok := svcerr.As(err)
		require.True(t, ok, "want svcerr, got %v", err)
		assert.Equal(t, svcerr.KindNotFound, se.Kind)
	})

	t.Run("player/delegates to PhotoKey", func(t *testing.T) {
		key, err := resolvers["player"](ctx, 1)
		require.NoError(t, err)
		assert.Equal(t, "player/one.png", key)
	})

	t.Run("player/PhotoKey NotFound passes through", func(t *testing.T) {
		_, err := resolvers["player"](ctx, 2)
		se, ok := svcerr.As(err)
		require.True(t, ok, "want svcerr, got %v", err)
		assert.Equal(t, svcerr.KindNotFound, se.Kind)
	})
}
