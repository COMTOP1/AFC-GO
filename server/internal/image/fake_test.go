package image_test

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"sync"

	"github.com/COMTOP1/AFC-GO/server/internal/image"
)

type fakeStore struct {
	mu     sync.Mutex
	rows   map[int]image.Image
	nextID int
}

func newFakeStore(rows ...image.Image) *fakeStore {
	f := &fakeStore{rows: map[int]image.Image{}, nextID: 100}
	for _, r := range rows {
		f.rows[r.ID] = r
	}
	return f
}

func (f *fakeStore) GetImages(context.Context) ([]image.Image, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]image.Image, 0, len(f.rows))
	for _, r := range f.rows {
		out = append(out, r)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func (f *fakeStore) GetImage(_ context.Context, i image.Image) (image.Image, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	r, ok := f.rows[i.ID]
	if !ok {
		return image.Image{}, fmt.Errorf("failed to get image: %w", sql.ErrNoRows)
	}
	return r, nil
}

func (f *fakeStore) AddImage(_ context.Context, i image.Image) (image.Image, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.nextID++
	i.ID = f.nextID
	f.rows[i.ID] = i
	return i, nil
}

func (f *fakeStore) DeleteImage(_ context.Context, i image.Image) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.rows, i.ID)
	return nil
}
