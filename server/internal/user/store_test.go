package user_test

import (
	"context"
	"testing"

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
