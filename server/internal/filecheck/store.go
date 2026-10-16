package filecheck

import (
	"context"
	"database/sql"
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

// recordTable is this package's own table. It is never searched for file
// links: it lists every key, so every file would look referenced.
const recordTable = "file_discrepancies"

// References maps each storage key a row points at to the rows pointing at
// it ("players 12"). It reads every table in the schema with a file_name
// column, found from the database itself, so a table the code doesn't know
// about (or one added later) can't have its files deleted by mistake.
func (s *Store) References(ctx context.Context) (map[string][]string, error) {
	ctx, span := tracer.Start(ctx, "filecheck.References")
	defer span.End()
	var tables []struct {
		Name  string `db:"table_name"`
		HasID bool   `db:"has_id"`
	}
	if err := s.db.SelectContext(ctx, &tables, `SELECT c.table_name,
			EXISTS (SELECT 1 FROM information_schema.columns i
			        WHERE i.table_schema = c.table_schema AND i.table_name = c.table_name AND i.column_name = 'id') AS has_id
		FROM information_schema.columns c
		JOIN information_schema.tables t ON t.table_schema = c.table_schema AND t.table_name = c.table_name
		WHERE c.table_schema = current_schema() AND c.column_name = 'file_name' AND t.table_type = 'BASE TABLE'
		ORDER BY c.table_name`); err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to find tables with files: %w", err)
	}
	refs := map[string][]string{}
	if len(tables) == 0 {
		return refs, nil
	}
	parts := make([]string, len(tables))
	for i, t := range tables {
		id := "NULL::text"
		if t.HasID {
			id = "id::text"
		}
		parts[i] = fmt.Sprintf("SELECT %s AS tbl, %s AS id, file_name::text AS file_name FROM %s WHERE file_name IS NOT NULL AND file_name::text <> ''",
			pq.QuoteLiteral(t.Name), id, pq.QuoteIdentifier(t.Name))
	}
	var rows []struct {
		Table    string         `db:"tbl"`
		ID       sql.NullString `db:"id"`
		FileName string         `db:"file_name"`
	}
	if err := s.db.SelectContext(ctx, &rows, strings.Join(parts, " UNION ALL ")); err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to read file references: %w", err)
	}
	for _, r := range rows {
		where := r.Table
		if r.ID.Valid {
			where += " " + r.ID.String
		}
		refs[r.FileName] = append(refs[r.FileName], where)
	}
	return refs, nil
}

// Texts returns every text value stored in the schema (article and event
// bodies, the info page, team descriptions...), in which a file may be
// linked directly. Columns are found from the database itself.
func (s *Store) Texts(ctx context.Context) ([]string, error) {
	ctx, span := tracer.Start(ctx, "filecheck.Texts")
	defer span.End()
	var columns []struct {
		Table  string `db:"table_name"`
		Column string `db:"column_name"`
	}
	if err := s.db.SelectContext(ctx, &columns, `SELECT c.table_name, c.column_name
		FROM information_schema.columns c
		JOIN information_schema.tables t ON t.table_schema = c.table_schema AND t.table_name = c.table_name
		WHERE c.table_schema = current_schema() AND t.table_type = 'BASE TABLE' AND c.table_name <> $1
		  AND c.data_type IN ('text', 'character varying', 'character', 'json', 'jsonb')
		ORDER BY c.table_name, c.column_name`, recordTable); err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to find text columns: %w", err)
	}
	if len(columns) == 0 {
		return nil, nil
	}
	parts := make([]string, len(columns))
	for i, c := range columns {
		parts[i] = fmt.Sprintf("SELECT %s::text AS v FROM %s WHERE %s IS NOT NULL",
			pq.QuoteIdentifier(c.Column), pq.QuoteIdentifier(c.Table), pq.QuoteIdentifier(c.Column))
	}
	var texts []string
	if err := s.db.SelectContext(ctx, &texts, strings.Join(parts, " UNION ALL ")); err != nil {
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
