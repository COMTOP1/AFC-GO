// Package auth owns login state: the encrypted session cookie, the route
// guards, password reset tokens and the login and password endpoints.
package auth

import (
	"context"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/server/internal/user"
)

// userKey is the session value holding the logged-in user.User. Cookies issued
// before the cutover use the same key, so its name must not change.
const userKey = "user"

func init() {
	gob.Register(user.User{})
}

// Config configures the session cookie. Keys are hex encoded; an empty or
// invalid (non-hex) key falls back to a random one (sessions then reset on
// restart, exactly as today). A valid-hex key of an unexpected length (e.g. a
// truncated AuthenticationKey) is still used as-is — production keeps using
// its existing keys — but a startup warning is logged naming the weak key.
type Config struct {
	CookieName        string
	AuthenticationKey string
	EncryptionKey     string
	Secure            bool
}

// UserGetter loads a user from the database (satisfied by *user.Store).
type UserGetter interface {
	GetUser(ctx context.Context, u user.User) (user.User, error)
}

// Sessions reads and writes the session cookie.
type Sessions struct {
	store *sessions.CookieStore
	name  string
	users UserGetter
}

func NewSessions(conf Config, users UserGetter) *Sessions {
	store := sessions.NewCookieStore(
		decodeKey(conf.AuthenticationKey, 64, "authentication"),
		decodeKey(conf.EncryptionKey, 32, "encryption"),
	)
	store.Options = &sessions.Options{
		MaxAge:   60 * 60 * 24,
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
		Secure:   conf.Secure,
	}
	name := conf.CookieName
	if name == "" {
		name = "session"
	}
	return &Sessions{store: store, name: name, users: users}
}

func decodeKey(hexKey string, size int, what string) []byte {
	key, err := hex.DecodeString(hexKey)
	if err != nil {
		slog.Warn(fmt.Sprintf("failed to decode %s key: %+v; generating a random key instead (sessions will reset on restart)", what, err))
	}
	if len(key) == 0 {
		if err == nil {
			slog.Warn(what + " key is empty; generating a random key instead (sessions will reset on restart)")
		}
		return securecookie.GenerateRandomKey(size)
	}
	if msg := keyWarning(what, key, size); msg != "" {
		slog.Warn(msg)
	}
	return key
}

// keyWarning reports a warning message when key is weak for its purpose
// ("authentication" or "encryption"), or "" when its length is fine.
// securecookie accepts an HMAC authentication key of any length, and a
// too-short key is still used (production must keep using its existing
// keys), but a short authentication key weakens the HMAC and an encryption
// key that isn't a valid AES size can't be used for encryption at all. size
// is the length decodeKey would generate for a missing key, quoted in the
// message as the recommended length.
func keyWarning(what string, key []byte, size int) string {
	switch what {
	case "authentication":
		if len(key) < 32 {
			return fmt.Sprintf("%s key is %d bytes, shorter than the recommended minimum of 32 bytes (generated keys use %d)", what, len(key), size)
		}
	case "encryption":
		switch len(key) {
		case 16, 24, 32:
		default:
			return fmt.Sprintf("%s key is %d bytes; AES requires a 16, 24 or 32 byte key (generated keys use %d)", what, len(key), size)
		}
	}
	return ""
}

// CookieStore exposes the underlying store (tests decode the cookie with it).
func (s *Sessions) CookieStore() *sessions.CookieStore { return s.store }

// Name is the session cookie name.
func (s *Sessions) Name() string { return s.name }

// User returns the user recorded in the session cookie, without touching the
// database. ok is false when nobody is logged in.
func (s *Sessions) User(r *http.Request) (user.User, bool) {
	sess, err := s.store.Get(r, s.name)
	if err != nil {
		return user.User{}, false
	}
	u, ok := sess.Values[userKey].(user.User)
	if !ok || !u.Authenticated || u.ID <= 0 {
		return user.User{}, false
	}
	return u, true
}

// Login records u in the session. remember extends the cookie to 31 days.
func (s *Sessions) Login(w http.ResponseWriter, r *http.Request, u user.User, remember bool) error {
	// A decode error (e.g. rotated keys) still yields a fresh session to overwrite.
	sess, _ := s.store.Get(r, s.name) //nolint:errcheck // see comment above
	u.Authenticated = true
	u.Password, u.Hash, u.Salt = null.String{}, null.String{}, null.String{}
	sess.Values[userKey] = u
	if remember {
		sess.Options.MaxAge = 86400 * 31
	}
	return sess.Save(r, w)
}

// Logout clears the session and expires the cookie.
func (s *Sessions) Logout(w http.ResponseWriter, r *http.Request) error {
	sess, _ := s.store.Get(r, s.name) //nolint:errcheck // a bad cookie is overwritten anyway
	sess.Values[userKey] = user.User{}
	sess.Options.MaxAge = -1
	return sess.Save(r, w)
}
