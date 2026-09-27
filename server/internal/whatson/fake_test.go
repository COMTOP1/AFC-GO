package whatson_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/COMTOP1/AFC-GO/server/internal/whatson"
)

var today = time.Now().Truncate(24 * time.Hour)

type fakeStore struct {
	mu      sync.Mutex
	rows    map[int]whatson.WhatsOn
	nextID  int
	editErr error
}

func newFakeStore(rows ...whatson.WhatsOn) *fakeStore {
	f := &fakeStore{rows: map[int]whatson.WhatsOn{}, nextID: 100}
	for _, r := range rows {
		f.rows[r.ID] = r
	}
	return f
}

func (f *fakeStore) filter(keep func(whatson.WhatsOn) bool, asc bool) []whatson.WhatsOn {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := []whatson.WhatsOn{}
	for _, r := range f.rows {
		if keep(r) {
			out = append(out, r)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if asc {
			return out[i].DateOfEvent.Before(out[j].DateOfEvent)
		}
		return out[i].DateOfEvent.After(out[j].DateOfEvent)
	})
	return out
}

func (f *fakeStore) GetWhatsOn(context.Context) ([]whatson.WhatsOn, error) {
	return f.filter(func(whatson.WhatsOn) bool { return true }, true), nil
}

func (f *fakeStore) GetWhatsOnFuture(context.Context) ([]whatson.WhatsOn, error) {
	return f.filter(func(w whatson.WhatsOn) bool { return !w.DateOfEvent.Before(today) }, true), nil
}

func (f *fakeStore) GetWhatsOnPast(context.Context) ([]whatson.WhatsOn, error) {
	return f.filter(func(w whatson.WhatsOn) bool { return w.DateOfEvent.Before(today) }, false), nil
}

func (f *fakeStore) GetWhatsOnLatest(ctx context.Context) (whatson.WhatsOn, error) {
	future, _ := f.GetWhatsOnFuture(ctx)
	if len(future) == 0 {
		return whatson.WhatsOn{}, fmt.Errorf("failed to get whats on latest: %w", sql.ErrNoRows)
	}
	return future[0], nil
}

func (f *fakeStore) GetWhatsOnArticle(_ context.Context, w whatson.WhatsOn) (whatson.WhatsOn, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	r, ok := f.rows[w.ID]
	if !ok {
		return whatson.WhatsOn{}, fmt.Errorf("failed to get whats on: %w", sql.ErrNoRows)
	}
	return r, nil
}

func (f *fakeStore) AddWhatsOn(_ context.Context, w whatson.WhatsOn) (whatson.WhatsOn, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.nextID++
	w.ID = f.nextID
	w.Date = time.Now()
	f.rows[w.ID] = w
	return w, nil
}

func (f *fakeStore) EditWhatsOn(_ context.Context, w whatson.WhatsOn) (whatson.WhatsOn, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.editErr != nil {
		return whatson.WhatsOn{}, f.editErr
	}
	if _, ok := f.rows[w.ID]; !ok {
		return whatson.WhatsOn{}, errors.New("no such row")
	}
	f.rows[w.ID] = w
	return w, nil
}

func (f *fakeStore) DeleteWhatsOn(_ context.Context, w whatson.WhatsOn) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.rows, w.ID)
	return nil
}

func (f *fakeStore) row(id int) whatson.WhatsOn {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.rows[id]
}
