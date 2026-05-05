-- Track which identity provider minted each user row. Idempotent —
-- rerunning migrate.sh in a loop is safe. Existing rows default to
-- 'password' so the backfill is implicit.
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS auth_provider VARCHAR(16) NOT NULL DEFAULT 'password';

CREATE INDEX IF NOT EXISTS idx_users_provider ON users (auth_provider);
