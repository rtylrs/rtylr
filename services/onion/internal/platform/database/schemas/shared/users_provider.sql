-- Track which identity provider minted each user row.
-- Existing rows default to 'password' so the backfill is implicit.
ALTER TABLE users
    ADD COLUMN auth_provider VARCHAR(16) NOT NULL DEFAULT 'password';

CREATE INDEX idx_users_provider ON users (auth_provider);
