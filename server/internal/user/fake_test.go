package user_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/lib/pq"

	"github.com/COMTOP1/AFC-GO/server/internal/infrastructure/mail"
	"github.com/COMTOP1/AFC-GO/server/internal/team"
	"github.com/COMTOP1/AFC-GO/server/internal/user"
)

type fakeStore struct {
	mu     sync.Mutex
	rows   map[int]user.User
	nextID int
}

func newFakeStore(rows ...user.User) *fakeStore {
	f := &fakeStore{rows: map[int]user.User{}, nextID: 100}
	for _, r := range rows {
		f.rows[r.ID] = r
	}
	return f
}

func (f *fakeStore) GetUsers(context.Context) ([]user.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]user.User, 0, len(f.rows))
	for _, r := range f.rows {
		out = append(out, r)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func (f *fakeStore) GetUser(_ context.Context, u user.User) (user.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if r, ok := f.rows[u.ID]; ok {
		return r, nil
	}
	return user.User{}, fmt.Errorf("failed to get user: %w", sql.ErrNoRows)
}

func (f *fakeStore) AddUser(_ context.Context, u user.User) (user.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, r := range f.rows {
		if r.Email == u.Email {
			return user.User{}, fmt.Errorf("failed to add user: %w", &pq.Error{Code: "23505"})
		}
	}
	f.nextID++
	u.ID = f.nextID
	f.rows[u.ID] = u
	return u, nil
}

func (f *fakeStore) EditUser(_ context.Context, u user.User) (user.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.rows[u.ID]; !ok {
		return user.User{}, errors.New("no such row")
	}
	f.rows[u.ID] = u
	return u, nil
}

func (f *fakeStore) DeleteUser(_ context.Context, u user.User) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.rows, u.ID)
	return nil
}

func (f *fakeStore) row(id int) user.User {
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

type fakeSender struct {
	err  error
	sent []mail.Mail
}

func (f *fakeSender) Send(_ context.Context, m mail.Mail) error {
	if f.err != nil {
		return f.err
	}
	f.sent = append(f.sent, m)
	return nil
}

type fakeTokens struct {
	tokens map[string]int
	ttl    time.Duration
}

func (f *fakeTokens) Set(_ context.Context, token string, userID int, ttl time.Duration) error {
	f.tokens[token] = userID
	f.ttl = ttl
	return nil
}
