package programme_test

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"sync"

	"github.com/COMTOP1/AFC-GO/server/internal/programme"
)

type fakeStore struct {
	mu         sync.Mutex
	programmes map[int]programme.Programme
	seasons    map[int]programme.Season
	nextID     int
}

func newFakeStore() *fakeStore {
	return &fakeStore{programmes: map[int]programme.Programme{}, seasons: map[int]programme.Season{}, nextID: 100}
}

func (f *fakeStore) list(keep func(programme.Programme) bool) []programme.Programme {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := []programme.Programme{}
	for _, p := range f.programmes {
		if keep(p) {
			out = append(out, p)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].DateOfProgramme.After(out[j].DateOfProgramme) })
	return out
}

func (f *fakeStore) GetProgrammes(context.Context) ([]programme.Programme, error) {
	return f.list(func(programme.Programme) bool { return true }), nil
}

func (f *fakeStore) GetProgrammesSeason(_ context.Context, s programme.Season) ([]programme.Programme, error) {
	return f.list(func(p programme.Programme) bool { return p.SeasonID == s.ID }), nil
}

func (f *fakeStore) GetProgramme(_ context.Context, p programme.Programme) (programme.Programme, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	r, ok := f.programmes[p.ID]
	if !ok {
		return programme.Programme{}, fmt.Errorf("failed to get programme: %w", sql.ErrNoRows)
	}
	return r, nil
}

func (f *fakeStore) AddProgramme(_ context.Context, p programme.Programme) (programme.Programme, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.nextID++
	p.ID = f.nextID
	f.programmes[p.ID] = p
	return p, nil
}

func (f *fakeStore) EditProgramme(_ context.Context, p programme.Programme) (programme.Programme, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.programmes[p.ID] = p
	return p, nil
}

func (f *fakeStore) DeleteProgramme(_ context.Context, p programme.Programme) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.programmes, p.ID)
	return nil
}

func (f *fakeStore) GetSeasons(context.Context) ([]programme.Season, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]programme.Season, 0, len(f.seasons))
	for _, s := range f.seasons {
		out = append(out, s)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func (f *fakeStore) GetSeason(_ context.Context, s programme.Season) (programme.Season, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	r, ok := f.seasons[s.ID]
	if !ok {
		return programme.Season{}, fmt.Errorf("failed to get season: %w", sql.ErrNoRows)
	}
	return r, nil
}

func (f *fakeStore) AddSeason(_ context.Context, s programme.Season) (programme.Season, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.nextID++
	s.ID = f.nextID
	f.seasons[s.ID] = s
	return s, nil
}

func (f *fakeStore) EditSeason(_ context.Context, s programme.Season) (programme.Season, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.seasons[s.ID] = s
	return s, nil
}

func (f *fakeStore) DeleteSeason(_ context.Context, s programme.Season) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.seasons, s.ID)
	return nil
}
