package player_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"sync"

	"github.com/COMTOP1/AFC-GO/server/internal/player"
	"github.com/COMTOP1/AFC-GO/server/internal/team"
)

type fakeStore struct {
	mu     sync.Mutex
	rows   map[int]player.Player
	nextID int
}

func newFakeStore(rows ...player.Player) *fakeStore {
	f := &fakeStore{rows: map[int]player.Player{}, nextID: 100}
	for _, r := range rows {
		f.rows[r.ID] = r
	}
	return f
}

func (f *fakeStore) sorted(keep func(player.Player) bool) []player.Player {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := []player.Player{}
	for _, r := range f.rows {
		if keep(r) {
			out = append(out, r)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func (f *fakeStore) GetPlayers(context.Context) ([]player.Player, error) {
	return f.sorted(func(player.Player) bool { return true }), nil
}

func (f *fakeStore) GetPlayersTeam(_ context.Context, t team.Team) ([]player.Player, error) {
	return f.sorted(func(p player.Player) bool { return p.TeamID == t.ID }), nil
}

func (f *fakeStore) GetPlayer(_ context.Context, p player.Player) (player.Player, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	r, ok := f.rows[p.ID]
	if !ok {
		return player.Player{}, fmt.Errorf("failed to get player: %w", sql.ErrNoRows)
	}
	return r, nil
}

func (f *fakeStore) AddPlayer(_ context.Context, p player.Player) (player.Player, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.nextID++
	p.ID = f.nextID
	f.rows[p.ID] = p
	return p, nil
}

func (f *fakeStore) EditPlayer(_ context.Context, p player.Player) (player.Player, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.rows[p.ID]; !ok {
		return player.Player{}, errors.New("no such row")
	}
	f.rows[p.ID] = p
	return p, nil
}

func (f *fakeStore) DeletePlayer(_ context.Context, p player.Player) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.rows, p.ID)
	return nil
}

func (f *fakeStore) row(id int) player.Player {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.rows[id]
}

type fakeTeams map[int]team.Team

func (f fakeTeams) GetTeam(_ context.Context, t team.Team) (team.Team, error) {
	if x, ok := f[t.ID]; ok {
		return x, nil
	}
	return team.Team{}, fmt.Errorf("failed to get team: %w", sql.ErrNoRows)
}

func (f fakeTeams) GetTeams(context.Context) ([]team.Team, error) {
	out := make([]team.Team, 0, len(f))
	for _, t := range f {
		out = append(out, t)
	}
	return out, nil
}
