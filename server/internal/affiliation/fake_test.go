package affiliation_test

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"sync"

	"github.com/COMTOP1/AFC-GO/server/internal/affiliation"
)

type fakeStore struct {
	mu     sync.Mutex
	rows   map[int]affiliation.Affiliation
	nextID int
}

func newFakeStore(rows ...affiliation.Affiliation) *fakeStore {
	f := &fakeStore{rows: map[int]affiliation.Affiliation{}, nextID: 100}
	for _, r := range rows {
		f.rows[r.ID] = r
	}
	return f
}

func (f *fakeStore) GetAffiliations(context.Context) ([]affiliation.Affiliation, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]affiliation.Affiliation, 0, len(f.rows))
	for _, r := range f.rows {
		out = append(out, r)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

func (f *fakeStore) GetAffiliationsMinimal(ctx context.Context) ([]affiliation.Affiliation, error) {
	return f.GetAffiliations(ctx)
}

func (f *fakeStore) GetAffiliation(_ context.Context, a affiliation.Affiliation) (affiliation.Affiliation, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	r, ok := f.rows[a.ID]
	if !ok {
		return affiliation.Affiliation{}, fmt.Errorf("failed to get affiliation: %w", sql.ErrNoRows)
	}
	return r, nil
}

func (f *fakeStore) AddAffiliation(_ context.Context, a affiliation.Affiliation) (affiliation.Affiliation, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.nextID++
	a.ID = f.nextID
	f.rows[a.ID] = a
	return a, nil
}

func (f *fakeStore) DeleteAffiliation(_ context.Context, a affiliation.Affiliation) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.rows, a.ID)
	return nil
}
