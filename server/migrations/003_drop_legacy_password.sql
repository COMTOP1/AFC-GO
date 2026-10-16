-- Drops the old site's unsalted SHA-512 password column. Nothing reads it
-- any more; resetlegacypasswords empties it.
--
-- Run this only AFTER resetlegacypasswords has run. It refuses (and changes
-- nothing) while any account still holds a legacy digest, so those users
-- are never left without a reset email. Safe to run more than once:
--   psql "$DATABASE_URL" -f server/migrations/003_drop_legacy_password.sql
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.columns
               WHERE table_schema = current_schema() AND table_name = 'users' AND column_name = 'password') THEN
        IF EXISTS (SELECT 1 FROM users WHERE password IS NOT NULL) THEN
            RAISE EXCEPTION 'users still hold legacy passwords: run resetlegacypasswords first';
        END IF;
        ALTER TABLE users DROP COLUMN password;
    END IF;
END
$$;
