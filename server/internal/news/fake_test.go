package news_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/COMTOP1/AFC-GO/server/internal/news"
)

type fakeStore struct {
	mu      sync.Mutex
	rows    map[int]news.News
	nextID  int
	editErr error
}

func newFakeStore(rows ...news.News) *fakeStore {
	f := &fakeStore{rows: map[int]news.News{}, nextID: 100}
	for _, r := range rows {
		f.rows[r.ID] = r
	}
	return f
}

func (f *fakeStore) GetNews(context.Context) ([]news.News, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]news.News, 0, len(f.rows))
	for _, r := range f.rows {
		out = append(out, r)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Date.After(out[j].Date) })
	return out, nil
}

func (f *fakeStore) GetNewsLatest(ctx context.Context) (news.News, error) {
	all, _ := f.GetNews(ctx)
	if len(all) == 0 {
		return news.News{}, fmt.Errorf("failed to get news latest: %w", sql.ErrNoRows)
	}
	return all[0], nil
}

func (f *fakeStore) GetNewsArticle(_ context.Context, n news.News) (news.News, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	r, ok := f.rows[n.ID]
	if !ok {
		return news.News{}, fmt.Errorf("failed to get news article: %w", sql.ErrNoRows)
	}
	return r, nil
}

func (f *fakeStore) AddNews(_ context.Context, n news.News) (news.News, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.nextID++
	n.ID = f.nextID
	n.Date = time.Now()
	f.rows[n.ID] = n
	return n, nil
}

func (f *fakeStore) EditNews(_ context.Context, n news.News) (news.News, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.editErr != nil {
		return news.News{}, f.editErr
	}
	if _, ok := f.rows[n.ID]; !ok {
		return news.News{}, errors.New("no such row")
	}
	f.rows[n.ID] = n
	return n, nil
}

func (f *fakeStore) DeleteNews(_ context.Context, n news.News) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.rows, n.ID)
	return nil
}

func (f *fakeStore) row(id int) news.News {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.rows[id]
}
