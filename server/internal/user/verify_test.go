package user_test

import (
	"context"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/server/internal/role"
	"github.com/COMTOP1/AFC-GO/server/internal/testdb"
	"github.com/COMTOP1/AFC-GO/server/internal/user"
	"github.com/COMTOP1/AFC-GO/server/internal/utils"
)

const (
	iter, workFactor, blockSize, parallelism, keyLen = 1, 2, 1, 1, 32
)

func verify(t *testing.T, store *user.Store, email, password string) (user.User, bool, error) {
	t.Helper()
	return store.VerifyUser(context.Background(), user.User{Email: email, Password: null.StringFrom(password)},
		iter, workFactor, blockSize, parallelism, keyLen)
}

// newUserLikeSignup stores a user exactly as legacy UserAddFunc and
// user.Service.Create do: raw hex-salt bytes, reset_password = true.
func newUserLikeSignup(t *testing.T, store *user.Store, email, password string) {
	t.Helper()
	salt, err := utils.GenerateRandom(utils.GenerateSalt)
	require.NoError(t, err)
	hash, err := utils.HashPassScrypt([]byte(password), []byte(salt), workFactor, blockSize, parallelism, keyLen)
	require.NoError(t, err)
	_, err = store.AddUser(context.Background(), user.User{Name: "New", Email: email, Role: role.Treasurer,
		ResetPassword: true, Hash: null.StringFrom(hash), Salt: null.StringFrom(salt)})
	require.NoError(t, err)
}

func TestResetFlaggedUserNeedsTheRightPassword(t *testing.T) {
	db, _ := testdb.Open(t)
	store := user.NewUserRepo(db)
	newUserLikeSignup(t, store, "new@example.test", "Temp-Pa55!")

	_, reset, err := verify(t, store, "new@example.test", "a guess")
	require.Error(t, err)
	assert.False(t, reset, "a wrong password must never be told 'reset required'")

	u, reset, err := verify(t, store, "new@example.test", "Temp-Pa55!")
	require.Error(t, err, "reset-required is still reported as an error to callers")
	assert.True(t, reset, "the emailed temporary password must now work")
	assert.Positive(t, u.ID)
}

func TestAdminResetUserKeepsWorkingPassword(t *testing.T) {
	db, _ := testdb.Open(t)
	store := user.NewUserRepo(db)
	ctx := context.Background()

	// An established account: hex-decoded salt, as EditUserPassword writes.
	saltHex, err := utils.GenerateRandom(utils.GenerateSalt)
	require.NoError(t, err)
	saltBytes, err := hex.DecodeString(saltHex)
	require.NoError(t, err)
	hash, err := utils.HashPassScrypt([]byte("Known-Pa55!"), saltBytes, workFactor, blockSize, parallelism, keyLen)
	require.NoError(t, err)
	_, err = store.AddUser(ctx, user.User{Name: "Old", Email: "old@example.test", Role: role.Treasurer,
		Hash: null.StringFrom(hash), Salt: null.StringFrom(saltHex)})
	require.NoError(t, err)

	_, reset, err := verify(t, store, "old@example.test", "Known-Pa55!")
	require.NoError(t, err)
	assert.False(t, reset)

	u, err := store.GetUser(ctx, user.User{Email: "old@example.test"})
	require.NoError(t, err)
	u.ResetPassword = true // admin clicked "reset password"
	_, err = store.EditUser(ctx, u)
	require.NoError(t, err)

	_, reset, _ = verify(t, store, "old@example.test", "wrong")
	assert.False(t, reset)
	_, reset, _ = verify(t, store, "old@example.test", "Known-Pa55!")
	assert.True(t, reset)
}
