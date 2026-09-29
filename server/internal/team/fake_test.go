package team_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"sync"

	"github.com/COMTOP1/AFC-GO/server/internal/team"
)

type fakeStore struct {
	mu     sync.Mutex
	rows   map[int]team.Team
	nextID int
}

func newFakeStore(rows ...team.Team) *fakeStore {
	f := &fakeStore{rows: map[int]team.Team{}, nextID: 100}
	for _, r := range rows {
		f.rows[r.ID] = r
	}
	return f
}

func (f *fakeStore) list(keep func(team.Team) bool) []team.Team {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := []team.Team{}
	for _, r := range f.rows {
		if keep(r) {
			out = append(out, r)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func (f *fakeStore) GetTeams(context.Context) ([]team.Team, error) {
	return f.list(func(team.Team) bool { return true }), nil
}

func (f *fakeStore) GetTeamsActive(context.Context) ([]team.Team, error) {
	return f.list(func(t team.Team) bool { return t.IsActive }), nil
}

func (f *fakeStore) GetTeam(_ context.Context, t team.Team) (team.Team, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	r, ok := f.rows[t.ID]
	if !ok {
		return team.Team{}, fmt.Errorf("failed to get team: %w", sql.ErrNoRows)
	}
	return r, nil
}

func (f *fakeStore) AddTeam(_ context.Context, t team.Team) (team.Team, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.nextID++
	t.ID = f.nextID
	f.rows[t.ID] = t
	return t, nil
}

func (f *fakeStore) EditTeam(_ context.Context, t team.Team) (team.Team, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.rows[t.ID]; !ok {
		return team.Team{}, errors.New("no such row")
	}
	f.rows[t.ID] = t
	return t, nil
}

func (f *fakeStore) DeleteTeam(_ context.Context, t team.Team) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.rows, t.ID)
	return nil
}

func (f *fakeStore) row(id int) team.Team {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.rows[id]
}

type recordingDetacher struct {
	name   string
	calls  *[]string
	failOn int
}

func (d recordingDetacher) DetachTeam(_ context.Context, teamID int) error {
	*d.calls = append(*d.calls, fmt.Sprintf("%s:%d", d.name, teamID))
	if teamID == d.failOn {
		return errors.New(d.name + " failed")
	}
	return nil
}
