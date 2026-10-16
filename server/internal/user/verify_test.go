package user_test

import (
	"context"
	"crypto/sha512"
	"encoding/hex"
	"strings"
	"testing"

	"github.com/jmoiron/sqlx"
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

// addLegacyUser stores a user the way the old site did: only the password
// column, no hash or salt.
func addLegacyUser(t *testing.T, db *sqlx.DB, email, stored string) {
	t.Helper()
	_, err := db.ExecContext(context.Background(), `INSERT INTO users (name, email, role, password) VALUES ('Legacy', $1, 'TREASURER', $2)`, email, stored)
	require.NoError(t, err)
}

func passwordColumn(t *testing.T, db *sqlx.DB, email string) (password, salt, hash null.String) {
	t.Helper()
	require.NoError(t, db.QueryRowContext(context.Background(), `SELECT password, salt, hash FROM users WHERE email = $1`, email).Scan(&password, &salt, &hash))
	return password, salt, hash
}

func TestWrapLegacyPasswordsWrapsEachDigestOnce(t *testing.T) {
	db, _ := testdb.Open(t)
	store := user.NewUserRepo(db)
	digest := sha512.Sum512([]byte("Legacy-Pa55!"))
	addLegacyUser(t, db, "legacy@example.test", hex.EncodeToString(digest[:]))
	newUserLikeSignup(t, store, "new@example.test", "Temp-Pa55!")

	n, err := store.WrapLegacyPasswords(context.Background(), workFactor, blockSize, parallelism, keyLen)
	require.NoError(t, err)
	assert.Equal(t, 1, n, "only the legacy row is wrapped")

	password, salt, _ := passwordColumn(t, db, "legacy@example.test")
	assert.True(t, strings.HasPrefix(password.String, "scrypt-sha512:"))
	assert.NotContains(t, password.String, hex.EncodeToString(digest[:]), "the bare digest must be gone")
	assert.True(t, salt.Valid)

	n, err = store.WrapLegacyPasswords(context.Background(), workFactor, blockSize, parallelism, keyLen)
	require.NoError(t, err)
	assert.Zero(t, n, "running it again does nothing")
}

func TestWrappedLegacyPasswordStillLogsInAndUpgrades(t *testing.T) {
	db, _ := testdb.Open(t)
	store := user.NewUserRepo(db)

	// What WrapLegacyPasswords stores for a legacy digest.
	digest := sha512.Sum512([]byte("Legacy-Pa55!"))
	salt, err := utils.GenerateRandom(utils.GenerateSalt)
	require.NoError(t, err)
	wrapped, err := utils.HashPassScrypt(digest[:], []byte(salt), workFactor, blockSize, parallelism, keyLen)
	require.NoError(t, err)
	addLegacyUser(t, db, "legacy@example.test", "scrypt-sha512:"+wrapped)
	_, err = db.ExecContext(context.Background(), `UPDATE users SET salt = $1 WHERE email = 'legacy@example.test'`, hex.EncodeToString([]byte(salt)))
	require.NoError(t, err)

	_, _, err = verify(t, store, "legacy@example.test", "wrong")
	require.Error(t, err)

	u, reset, err := verify(t, store, "legacy@example.test", "Legacy-Pa55!")
	require.NoError(t, err)
	assert.False(t, reset)
	assert.Positive(t, u.ID)

	password, _, hash := passwordColumn(t, db, "legacy@example.test")
	assert.False(t, password.Valid, "the legacy column is cleared once upgraded")
	assert.True(t, hash.Valid)

	_, _, err = verify(t, store, "legacy@example.test", "Legacy-Pa55!")
	require.NoError(t, err, "the upgraded hash keeps working")
}
