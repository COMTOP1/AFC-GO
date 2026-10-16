package filecheck

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

// Kind is a type of difference between the database and the file store.
type Kind string

const (
	// Unreferenced is a file in the bucket that no row points to.
	Unreferenced Kind = "unreferenced"
	// Missing is a row pointing at a file that isn't in the bucket.
	Missing Kind = "missing"
)

// Discrepancy is one recorded difference.
type Discrepancy struct {
	Key  string `db:"key"`
	Kind Kind   `db:"kind"`
	// Detail names the rows that point at a missing file, e.g. "players 12".
	Detail    string    `db:"detail"`
	FirstSeen time.Time `db:"first_seen"`
	LastSeen  time.Time `db:"last_seen"`
}

// Store reads file references from the database and keeps the
// file_discrepancies table (migration 005).
type Store struct {
	db *sqlx.DB
}

func NewStore(db *sqlx.DB) *Store { return &Store{db: db} }

// fileColumns is every column that holds a storage key. A table added with
// a file column must be added here, or its files will look unreferenced and
// be deleted after 30 days.
var fileColumns = []string{"affiliations", "documents", "images", "news", "players", "programmes", "sponsors", "teams", "users", "whatson"}

// References maps each storage key a row points at to the rows pointing at
// it ("players 12").
func (s *Store) References(ctx context.Context) (map[string][]string, error) {
	ctx, span := tracer.Start(ctx, "filecheck.References")
	defer span.End()
	parts := make([]string, len(fileColumns))
	for i, table := range fileColumns {
		parts[i] = fmt.Sprintf("SELECT '%s' AS tbl, id, file_name FROM %s WHERE file_name IS NOT NULL AND file_name <> ''", table, table)
	}
	query := strings.Join(parts, " UNION ALL ")
	var rows []struct {
		Table    string `db:"tbl"`
		ID       int    `db:"id"`
		FileName string `db:"file_name"`
	}
	if err := s.db.SelectContext(ctx, &rows, query); err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to read file references: %w", err)
	}
	refs := make(map[string][]string, len(rows))
	for _, r := range rows {
		refs[r.FileName] = append(refs[r.FileName], fmt.Sprintf("%s %d", r.Table, r.ID))
	}
	return refs, nil
}

// Texts returns all stored rich text (articles, events, the info page), in
// which a file may be linked directly.
func (s *Store) Texts(ctx context.Context) ([]string, error) {
	ctx, span := tracer.Start(ctx, "filecheck.Texts")
	defer span.End()
	var texts []string
	err := s.db.SelectContext(ctx, &texts, `SELECT content FROM news WHERE content IS NOT NULL
		UNION ALL SELECT content FROM whatson WHERE content IS NOT NULL
		UNION ALL SELECT setting_text FROM settings`)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to read stored text: %w", err)
	}
	return texts, nil
}

// Record saves what a check found at now: new differences are added, ones
// seen before keep their first_seen, and ones no longer found are removed.
func (s *Store) Record(ctx context.Context, now time.Time, found []Discrepancy) error {
	ctx, span := tracer.Start(ctx, "filecheck.Record")
	defer span.End()
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to start recording file check: %w", err)
	}
	defer func() { _ = tx.Rollback() }() //nolint:errcheck // a no-op after Commit
	for _, d := range found {
		if _, err = tx.ExecContext(ctx, `INSERT INTO file_discrepancies (key, kind, detail, first_seen, last_seen)
			VALUES ($1, $2, $3, $4, $4)
			ON CONFLICT (key, kind) DO UPDATE SET detail = EXCLUDED.detail, last_seen = EXCLUDED.last_seen`,
			d.Key, d.Kind, d.Detail, now); err != nil {
			span.RecordError(err)
			return fmt.Errorf("failed to record %s file %q: %w", d.Kind, d.Key, err)
		}
	}
	keys, kinds := make([]string, len(found)), make([]string, len(found))
	for i, d := range found {
		keys[i], kinds[i] = d.Key, string(d.Kind)
	}
	if _, err = tx.ExecContext(ctx, `DELETE FROM file_discrepancies
		WHERE (key, kind) NOT IN (SELECT * FROM unnest($1::text[], $2::text[]))`, pq.Array(keys), pq.Array(kinds)); err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to clear resolved file differences: %w", err)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to record file check: %w", err)
	}
	return nil
}

// List returns every recorded difference, oldest first.
func (s *Store) List(ctx context.Context) ([]Discrepancy, error) {
	var out []Discrepancy
	if err := s.db.SelectContext(ctx, &out, `SELECT key, kind, detail, first_seen, last_seen
		FROM file_discrepancies ORDER BY first_seen, key`); err != nil {
		return nil, fmt.Errorf("failed to list file differences: %w", err)
	}
	return out, nil
}

// Forget removes one recorded difference (after its file is deleted).
func (s *Store) Forget(ctx context.Context, key string, kind Kind) error {
	if _, err := s.db.ExecContext(ctx, `DELETE FROM file_discrepancies WHERE key = $1 AND kind = $2`, key, kind); err != nil {
		return fmt.Errorf("failed to forget %s file %q: %w", kind, key, err)
	}
	return nil
}

// lockID is the Postgres advisory lock that lets only one server instance
// run the daily check at a time.
const lockID int64 = 0x4146_4346_494c_4553 // "AFCFILES"

// WithLock runs f only if no other instance holds the file check lock, and
// reports whether it ran.
func (s *Store) WithLock(ctx context.Context, f func(context.Context) error) (bool, error) {
	conn, err := s.db.Connx(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to get a connection for the file check lock: %w", err)
	}
	defer conn.Close()
	var got bool
	if err = conn.GetContext(ctx, &got, `SELECT pg_try_advisory_lock($1)`, lockID); err != nil {
		return false, fmt.Errorf("failed to take the file check lock: %w", err)
	}
	if !got {
		return false, nil
	}
	defer func() {
		_, _ = conn.ExecContext(context.WithoutCancel(ctx), `SELECT pg_advisory_unlock($1)`, lockID) //nolint:errcheck // closing the connection frees it anyway
	}()
	return true, f(ctx)
}
