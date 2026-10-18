-- Lets a team have its own shared contact address (e.g. under12s@...),
-- shown on its team page instead of its managers' personal emails when the
-- teamEmailMode setting is "team".
--
-- Safe to run more than once. Run it before deploying the server version
-- that reads this column:
--   psql "$DATABASE_URL" -f server/migrations/004_team_contact_email.sql
ALTER TABLE teams ADD COLUMN IF NOT EXISTS contact_email TEXT;
