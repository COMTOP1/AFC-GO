package user_test

import (
	"context"
	"testing"

	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/server/internal/role"
	"github.com/COMTOP1/AFC-GO/server/internal/testdb"
	"github.com/COMTOP1/AFC-GO/server/internal/user"
)

func TestAddUserReturnsID(t *testing.T) {
	db, _ := testdb.Open(t)
	added, err := user.NewUserRepo(db).AddUser(context.Background(),
		user.User{Name: "New", Email: "new@example.test", Role: role.Treasurer, ResetPassword: true})
	require.NoError(t, err)
	require.Positive(t, added.ID)
}

func TestEditUserClearsPhone(t *testing.T) {
	db, _ := testdb.Open(t)
	store := user.NewUserRepo(db)
	ctx := context.Background()
	u, err := store.GetUser(ctx, user.User{ID: 1})
	require.NoError(t, err)
	require.True(t, u.Phone.Valid)

	u.Phone = null.String{}
	_, err = store.EditUser(ctx, u)
	require.NoError(t, err)
	got, err := store.GetUser(ctx, user.User{ID: 1})
	require.NoError(t, err)
	assert.False(t, got.Phone.Valid)
}

// TestEditUserRejectsEmailAlreadyUsedByAnotherUser guards against a
// regression where EditUser looked up the row to update with
// "WHERE email = $newEmail OR id = $id": changing user 3's email to one
// user 1 already had could match and silently overwrite user 1 instead of
// hitting the unique constraint on email.
func TestEditUserRejectsEmailAlreadyUsedByAnotherUser(t *testing.T) {
	db, _ := testdb.Open(t)
	store := user.NewUserRepo(db)
	ctx := context.Background()

	before, err := store.GetUser(ctx, user.User{ID: 1})
	require.NoError(t, err)

	target, err := store.GetUser(ctx, user.User{ID: 3})
	require.NoError(t, err)
	target.Email = before.Email
	target.Name = "Hijacked"

	_, err = store.EditUser(ctx, target)
	require.Error(t, err)
	var pqErr *pq.Error
	require.ErrorAs(t, err, &pqErr, "want a pq error, got %v", err)
	assert.Equal(t, "23505", string(pqErr.Code))

	after, err := store.GetUser(ctx, user.User{ID: 1})
	require.NoError(t, err)
	assert.Equal(t, before, after, "user 1 must be completely unchanged")
}
