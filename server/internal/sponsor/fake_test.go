package sponsor_test

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strconv"
	"sync"

	"github.com/COMTOP1/AFC-GO/server/internal/sponsor"
	"github.com/COMTOP1/AFC-GO/server/internal/team"
)

type fakeStore struct {
	mu     sync.Mutex
	rows   map[int]sponsor.Sponsor
	nextID int
}

func newFakeStore(rows ...sponsor.Sponsor) *fakeStore {
	f := &fakeStore{rows: map[int]sponsor.Sponsor{}, nextID: 100}
	for _, r := range rows {
		f.rows[r.ID] = r
	}
	return f
}

func (f *fakeStore) all() []sponsor.Sponsor {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]sponsor.Sponsor, 0, len(f.rows))
	for _, r := range f.rows {
		out = append(out, r)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func (f *fakeStore) GetSponsors(context.Context) ([]sponsor.Sponsor, error) { return f.all(), nil }
func (f *fakeStore) GetSponsorsMinimal(context.Context) ([]sponsor.Sponsor, error) {
	return f.all(), nil
}

func (f *fakeStore) GetSponsorsTeam(_ context.Context, t team.Team) ([]sponsor.Sponsor, error) {
	out := []sponsor.Sponsor{}
	for _, r := range f.all() {
		if r.TeamID == strconv.Itoa(t.ID) {
			out = append(out, r)
		}
	}
	return out, nil
}

func (f *fakeStore) GetSponsor(_ context.Context, s sponsor.Sponsor) (sponsor.Sponsor, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	r, ok := f.rows[s.ID]
	if !ok {
		return sponsor.Sponsor{}, fmt.Errorf("failed to get sponsor: %w", sql.ErrNoRows)
	}
	return r, nil
}

func (f *fakeStore) AddSponsor(_ context.Context, s sponsor.Sponsor) (sponsor.Sponsor, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.nextID++
	s.ID = f.nextID
	f.rows[s.ID] = s
	return s, nil
}

func (f *fakeStore) DeleteSponsor(_ context.Context, s sponsor.Sponsor) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.rows, s.ID)
	return nil
}

type fakeTeams map[int]team.Team

func (f fakeTeams) GetTeam(_ context.Context, t team.Team) (team.Team, error) {
	if x, ok := f[t.ID]; ok {
		return x, nil
	}
	return team.Team{}, fmt.Errorf("failed to get team: %w", sql.ErrNoRows)
}
