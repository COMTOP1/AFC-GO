-- Lets a password change sign the user out everywhere else. Sessions are
-- cookies, so the server can't delete them; instead each cookie carries the
-- version it was issued at, and bumping the user's version makes every older
-- cookie stop working.
--
-- Safe to run more than once. Run it before deploying the server version
-- that reads this column:
--   psql "$DATABASE_URL" -f server/migrations/002_session_version.sql
ALTER TABLE users ADD COLUMN IF NOT EXISTS session_version INTEGER NOT NULL DEFAULT 0;
