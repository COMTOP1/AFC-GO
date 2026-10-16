-- Gives users and players a random public ID. File links for people use it
-- instead of their integer ID, so photos can't be collected by counting
-- through /api/v1/files/user/1, /2, /3...
--
-- Needs PostgreSQL 13 or later for gen_random_uuid(). Safe to run more than
-- once. Run it before deploying the server version that reads these columns:
--   psql "$DATABASE_URL" -f server/migrations/001_public_uuids.sql
BEGIN;

ALTER TABLE users ADD COLUMN IF NOT EXISTS uuid UUID NOT NULL DEFAULT gen_random_uuid();
ALTER TABLE players ADD COLUMN IF NOT EXISTS uuid UUID NOT NULL DEFAULT gen_random_uuid();

CREATE UNIQUE INDEX IF NOT EXISTS users_uuid_key ON users (uuid);
CREATE UNIQUE INDEX IF NOT EXISTS players_uuid_key ON players (uuid);

COMMIT;
