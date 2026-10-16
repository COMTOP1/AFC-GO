-- Records differences between the database and the file store, found by the
-- daily file check (and the checkfiles command):
--   unreferenced: a file in the bucket that no row points to. Deleted once it
--                 has stayed unreferenced for 30 days.
--   missing:      a row pointing at a file that isn't in the bucket. Never
--                 deleted automatically; detail says which rows.
-- A row disappears as soon as a check no longer finds that difference.
--
-- Safe to run more than once. Run it before deploying the server version
-- that uses this table:
--   psql "$DATABASE_URL" -f server/migrations/005_file_discrepancies.sql
CREATE TABLE IF NOT EXISTS file_discrepancies (
    key        TEXT        NOT NULL,
    kind       TEXT        NOT NULL CHECK (kind IN ('unreferenced', 'missing')),
    detail     TEXT        NOT NULL DEFAULT '',
    first_seen TIMESTAMPTZ NOT NULL,
    last_seen  TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (key, kind)
);
