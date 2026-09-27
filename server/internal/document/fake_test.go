package document_test

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"sync"

	"github.com/COMTOP1/AFC-GO/server/internal/document"
)

type fakeStore struct {
	mu     sync.Mutex
	rows   map[int]document.Document
	nextID int
}

func newFakeStore(rows ...document.Document) *fakeStore {
	f := &fakeStore{rows: map[int]document.Document{}, nextID: 100}
	for _, r := range rows {
		f.rows[r.ID] = r
	}
	return f
}

func (f *fakeStore) GetDocuments(context.Context) ([]document.Document, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]document.Document, 0, len(f.rows))
	for _, r := range f.rows {
		out = append(out, r)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

func (f *fakeStore) GetDocument(_ context.Context, d document.Document) (document.Document, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	r, ok := f.rows[d.ID]
	if !ok {
		return document.Document{}, fmt.Errorf("failed to get document: %w", sql.ErrNoRows)
	}
	return r, nil
}

func (f *fakeStore) AddDocument(_ context.Context, d document.Document) (document.Document, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.nextID++
	d.ID = f.nextID
	f.rows[d.ID] = d
	return d, nil
}

func (f *fakeStore) DeleteDocument(_ context.Context, d document.Document) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.rows, d.ID)
	return nil
}
