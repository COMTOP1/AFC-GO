package user_test

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	"gopkg.in/guregu/null.v4"

	_ "github.com/lib/pq"

	"github.com/COMTOP1/AFC-GO/role"
	"github.com/COMTOP1/AFC-GO/user"
	"github.com/COMTOP1/AFC-GO/utils"
)

const (
	iter, workFactor, blockSize, parallelism, keyLen = 1, 2, 1, 1, 32
)

// openTestDB connects to the postgres URL given by AFC_TEST_DB, creates a
// uniquely named schema containing only the users table, and registers a
// cleanup that drops the schema. Tests using this helper are skipped when
// AFC_TEST_DB is not set.
func openTestDB(t *testing.T) *sqlx.DB {
	t.Helper()

	dsn := os.Getenv("AFC_TEST_DB")
	if dsn == "" {
		t.Skip("AFC_TEST_DB not set, skipping DB-backed test")
	}

	schema, err := randomSchemaName()
	if err != nil {
		t.Fatalf("failed to generate random schema name: %v", err)
	}

	adminDB, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		t.Fatalf("failed to connect to test db: %v", err)
	}
	defer func() {
		_ = adminDB.Close()
	}()

	if _, err = adminDB.Exec(fmt.Sprintf("CREATE SCHEMA %q", schema)); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	t.Cleanup(func() {
		cleanupDB, cleanupErr := sqlx.Connect("postgres", dsn)
		if cleanupErr != nil {
			t.Logf("failed to connect to test db for cleanup: %v", cleanupErr)
			return
		}
		defer func() {
			_ = cleanupDB.Close()
		}()
		if _, cleanupErr = cleanupDB.Exec(fmt.Sprintf("DROP SCHEMA IF EXISTS %q CASCADE", schema)); cleanupErr != nil {
			t.Logf("failed to drop schema: %v", cleanupErr)
		}
	})

	db, err := sqlx.Connect("postgres", withSearchPath(dsn, schema))
	if err != nil {
		t.Fatalf("failed to connect to test db with search_path: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	const createUsersTable = `
		CREATE TABLE users (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL,
			email TEXT NOT NULL UNIQUE,
			phone TEXT,
			team_id INTEGER NOT NULL DEFAULT 0,
			role TEXT NOT NULL,
			file_name TEXT,
			reset_password BOOLEAN NOT NULL DEFAULT FALSE,
			password TEXT,
			hash TEXT,
			salt TEXT
		)`
	if _, err = db.Exec(createUsersTable); err != nil {
		t.Fatalf("failed to create users table: %v", err)
	}

	return db
}

// randomSchemaName returns a unique, safe-to-quote schema name.
func randomSchemaName() (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "t_" + hex.EncodeToString(b), nil
}

// withSearchPath appends an options query parameter that sets search_path
// to the given schema for the connection.
func withSearchPath(dsn, schema string) string {
	sep := "?"
	if hasQuery(dsn) {
		sep = "&"
	}
	return fmt.Sprintf("%s%soptions=-c search_path=%s", dsn, sep, schema)
}

func hasQuery(dsn string) bool {
	for _, c := range dsn {
		if c == '?' {
			return true
		}
	}
	return false
}

func verify(t *testing.T, store *user.Store, email, password string) (user.User, bool, error) {
	t.Helper()
	return store.VerifyUser(context.Background(), user.User{Email: email, Password: null.StringFrom(password)},
		iter, workFactor, blockSize, parallelism, keyLen)
}

// newUserLikeSignup stores a user exactly as the signup flow does: raw
// hex-salt bytes used directly (not decoded) as the scrypt salt, and
// reset_password = true.
func newUserLikeSignup(t *testing.T, store *user.Store, email, password string) {
	t.Helper()
	salt, err := utils.GenerateRandom(utils.GenerateSalt)
	if err != nil {
		t.Fatalf("failed to generate salt: %v", err)
	}
	hash, err := utils.HashPassScrypt([]byte(password), []byte(salt), workFactor, blockSize, parallelism, keyLen)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}
	_, err = store.AddUser(context.Background(), user.User{Name: "New", Email: email, Role: role.Treasurer,
		ResetPassword: true, Hash: null.StringFrom(hash), Salt: null.StringFrom(salt)})
	if err != nil {
		t.Fatalf("failed to add user: %v", err)
	}
}

func TestResetFlaggedUserNeedsTheRightPassword(t *testing.T) {
	db := openTestDB(t)
	store := user.NewUserRepo(db)
	newUserLikeSignup(t, store, "new@example.test", "Temp-Pa55!")

	_, reset, err := verify(t, store, "new@example.test", "a guess")
	if err == nil {
		t.Fatal("expected an error for a wrong password")
	}
	if reset {
		t.Error("a wrong password must never be told 'reset required'")
	}

	u, reset, err := verify(t, store, "new@example.test", "Temp-Pa55!")
	if err == nil {
		t.Fatal("reset-required is still reported as an error to callers")
	}
	if !reset {
		t.Error("the emailed temporary password must now work")
	}
	if u.ID <= 0 {
		t.Errorf("expected a positive user ID, got %d", u.ID)
	}
}

func TestAdminResetUserKeepsWorkingPassword(t *testing.T) {
	db := openTestDB(t)
	store := user.NewUserRepo(db)
	ctx := context.Background()

	// An established account: hex-decoded salt, as EditUserPassword writes.
	saltHex, err := utils.GenerateRandom(utils.GenerateSalt)
	if err != nil {
		t.Fatalf("failed to generate salt: %v", err)
	}
	saltBytes, err := hex.DecodeString(saltHex)
	if err != nil {
		t.Fatalf("failed to decode salt: %v", err)
	}
	hash, err := utils.HashPassScrypt([]byte("Known-Pa55!"), saltBytes, workFactor, blockSize, parallelism, keyLen)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}
	_, err = store.AddUser(ctx, user.User{Name: "Old", Email: "old@example.test", Role: role.Treasurer,
		Hash: null.StringFrom(hash), Salt: null.StringFrom(saltHex)})
	if err != nil {
		t.Fatalf("failed to add user: %v", err)
	}

	_, reset, err := verify(t, store, "old@example.test", "Known-Pa55!")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if reset {
		t.Error("expected reset to be false")
	}

	u, err := store.GetUser(ctx, user.User{Email: "old@example.test"})
	if err != nil {
		t.Fatalf("failed to get user: %v", err)
	}
	u.ResetPassword = true // admin clicked "reset password"
	_, err = store.EditUser(ctx, u)
	if err != nil {
		t.Fatalf("failed to edit user: %v", err)
	}

	_, reset, _ = verify(t, store, "old@example.test", "wrong")
	if reset {
		t.Error("a wrong password must never be told 'reset required'")
	}
	_, reset, _ = verify(t, store, "old@example.test", "Known-Pa55!")
	if !reset {
		t.Error("expected reset to be true for the admin-reset account")
	}
}
