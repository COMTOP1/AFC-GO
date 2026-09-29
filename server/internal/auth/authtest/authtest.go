// Package authtest builds sessions and cookies for handler tests.
package authtest

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/COMTOP1/AFC-GO/server/internal/auth"
	"github.com/COMTOP1/AFC-GO/server/internal/role"
	"github.com/COMTOP1/AFC-GO/server/internal/upload"
	"github.com/COMTOP1/AFC-GO/server/internal/upload/uploadtest"
	"github.com/COMTOP1/AFC-GO/server/internal/user"
)

// Personas used across handler tests.
var (
	Webmaster    = user.User{ID: 1, Name: "Web Master", Email: "webmaster@example.test", Role: role.Webmaster}
	Manager      = user.User{ID: 2, Name: "Team Manager", Email: "manager@example.test", Role: role.Manager, TeamID: 1}
	Photographer = user.User{ID: 3, Name: "Photo Grapher", Email: "photo@example.test", Role: role.Photographer}
	Treasurer    = user.User{ID: 4, Name: "Trea Surer", Email: "treasurer@example.test", Role: role.Treasurer}
)

// Users is an in-memory auth.UserGetter.
type Users struct {
	mu   sync.Mutex
	byID map[int]user.User
}

func NewUsers(users ...user.User) *Users {
	u := &Users{byID: map[int]user.User{}}
	for _, x := range users {
		u.byID[x.ID] = x
	}
	return u
}

func (u *Users) GetUser(_ context.Context, p user.User) (user.User, error) {
	u.mu.Lock()
	defer u.mu.Unlock()
	if x, ok := u.byID[p.ID]; ok {
		return x, nil
	}
	return user.User{}, fmt.Errorf("failed to get user: %w", sql.ErrNoRows)
}

// New returns sessions with fixed keys backed by the given users.
func New(users ...user.User) *auth.Sessions {
	return auth.NewSessions(auth.Config{
		CookieName:        "session",
		AuthenticationKey: strings.Repeat("ab", 64),
		EncryptionKey:     strings.Repeat("cd", 32),
	}, NewUsers(users...))
}

// Everyone returns sessions that know all four personas.
func Everyone() *auth.Sessions {
	return New(Webmaster, Manager, Photographer, Treasurer)
}

// Cookie logs u in and returns the resulting session cookie.
func Cookie(t *testing.T, s *auth.Sessions, u user.User) *http.Cookie {
	t.Helper()
	rec := httptest.NewRecorder()
	require.NoError(t, s.Login(rec, httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/", nil), u, false))
	for _, c := range rec.Result().Cookies() {
		if c.Name == s.Name() {
			return c
		}
	}
	t.Fatal("no session cookie written")
	return nil
}

// Files returns an upload.Files over an in-memory store.
func Files() *upload.Files {
	return upload.New(uploadtest.New())
}
