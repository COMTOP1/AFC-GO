package account_test

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/server/internal/account"
	"github.com/COMTOP1/AFC-GO/server/internal/auth"
	"github.com/COMTOP1/AFC-GO/server/internal/auth/authtest"
	"github.com/COMTOP1/AFC-GO/server/internal/upload"
	"github.com/COMTOP1/AFC-GO/server/internal/upload/uploadtest"
	"github.com/COMTOP1/AFC-GO/server/internal/user"
	"github.com/COMTOP1/AFC-GO/server/internal/web"
	"github.com/COMTOP1/AFC-GO/server/internal/web/apitest"
)

type fakeStore struct{ rows map[int]user.User }

func (f *fakeStore) EditUserImage(_ context.Context, u user.User) error {
	if _, ok := f.rows[u.ID]; !ok {
		return fmt.Errorf("failed to get user: %w", sql.ErrNoRows)
	}
	row := f.rows[u.ID]
	row.FileName = u.FileName
	f.rows[u.ID] = row
	return nil
}

// GetUser lets fakeStore double as the session guard's UserGetter, so a
// guard's per-request reload (auth.Sessions.RequireLogin) sees the same
// user row the account service just edited, exactly as production does
// with one shared *user.Store.
func (f *fakeStore) GetUser(_ context.Context, u user.User) (user.User, error) {
	if x, ok := f.rows[u.ID]; ok {
		return x, nil
	}
	return user.User{}, fmt.Errorf("failed to get user: %w", sql.ErrNoRows)
}

func TestAccountImage(t *testing.T) {
	me := authtest.Treasurer
	me.FileName = null.StringFrom("user/old.png")
	store := &fakeStore{rows: map[int]user.User{me.ID: me}}
	objects := uploadtest.New()
	objects.Objects["user/old.png"] = "OLD"
	files := upload.New(objects)
	// Backed by store (not authtest.New) so the guard's per-request reload
	// reflects the account service's edits, as it would against production's
	// single shared *user.Store.
	sessions := auth.NewSessions(auth.Config{CookieName: "session"}, store)

	e := apitest.NewEcho()
	account.NewHandlers(account.NewService(store, files), files).Register(web.NewAPI(e, false), sessions.Guards())
	anon := apitest.New(e)
	client := anon.As(authtest.Cookie(t, sessions, me))

	assert.Equal(t, http.StatusUnauthorized, anon.Get(t, "/api/v1/account").Code)
	rec := client.Get(t, "/api/v1/account")
	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "https://cdn.test/user/old.png", apitest.Decode[auth.CurrentUser](t, rec).ImageURL)

	rec = client.Multipart(t, http.MethodPut, "/api/v1/account/image", nil,
		apitest.FilePart{Field: "image", Name: "me.png", ContentType: "image/png", Body: "NEW"})
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	newKey := store.rows[me.ID].FileName.String
	assert.NotEqual(t, "user/old.png", newKey)
	assert.Equal(t, []string{"user/old.png"}, objects.Deleted, "old image removed only after the new one is saved")

	rec = client.JSON(t, http.MethodDelete, "/api/v1/account/image", nil)
	require.Equal(t, http.StatusOK, rec.Code)
	assert.False(t, store.rows[me.ID].FileName.Valid)
	assert.Contains(t, objects.Deleted, newKey)

	rec = client.Multipart(t, http.MethodPut, "/api/v1/account/image", nil)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code, "image is required")
}
