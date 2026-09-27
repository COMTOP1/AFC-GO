// Package testdb gives DB-backed tests a throwaway Postgres schema loaded
// with a synthetic fixture. Tests are skipped when AFC_TEST_DB is unset.
package testdb

import (
	"crypto/rand"
	_ "embed"
	"encoding/hex"
	"net/url"
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // postgres driver
	"github.com/stretchr/testify/require"
)

//go:embed schema.sql
var schemaSQL string

//go:embed seed.sql
var seedSQL string

// Open creates a uniquely named schema in the AFC_TEST_DB database, loads
// schema.sql and seed.sql into it, and returns a connection plus a DSN whose
// search_path points at that schema. The schema is dropped when the test ends.
func Open(t *testing.T) (*sqlx.DB, string) {
	t.Helper()
	base := os.Getenv("AFC_TEST_DB")
	if base == "" {
		t.Skip("AFC_TEST_DB not set; skipping DB-backed test")
	}

	suffix := make([]byte, 6)
	_, err := rand.Read(suffix)
	require.NoError(t, err)
	schema := "t_" + hex.EncodeToString(suffix)

	admin, err := sqlx.Connect("postgres", base)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = admin.Exec("DROP SCHEMA IF EXISTS " + schema + " CASCADE")
		_ = admin.Close()
	})
	_, err = admin.Exec("CREATE SCHEMA " + schema)
	require.NoError(t, err)

	u, err := url.Parse(base)
	require.NoError(t, err)
	q := u.Query()
	q.Set("search_path", schema)
	u.RawQuery = q.Encode()
	dsn := u.String()

	db, err := sqlx.Connect("postgres", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	_, err = db.Exec(schemaSQL)
	require.NoError(t, err, "loading schema.sql")
	_, err = db.Exec(seedSQL)
	require.NoError(t, err, "loading seed.sql")
	return db, dsn
}
