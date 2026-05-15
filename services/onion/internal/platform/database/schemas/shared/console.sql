-- Console admin panel schema
-- Adds internal admin capabilities to the users table and a grants table for manual access deals.

ALTER TABLE users
    ADD COLUMN IF NOT EXISTS is_console_admin BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS is_banned        BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS ban_reason       TEXT NULL,
    ADD COLUMN IF NOT EXISTS banned_at        TIMESTAMP NULL,
    ADD INDEX IF NOT EXISTS idx_users_console_admin (is_console_admin),
    ADD INDEX IF NOT EXISTS idx_users_banned (is_banned);

-- Manual access grants: used when a deal is made outside of Paddle (e.g. cash payment).
CREATE TABLE IF NOT EXISTS manual_access_grants (
    id           BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    user_id      BIGINT NOT NULL,
    plan_id      BIGINT NOT NULL,
    granted_by   BIGINT NOT NULL COMMENT 'console admin user_id who created the grant',
    note         TEXT NULL       COMMENT 'deal notes, e.g. "cash payment $500/yr agreed 2026-05-14"',
    expires_at   TIMESTAMP NULL  COMMENT 'NULL = never expires',
    revoked_at   TIMESTAMP NULL,
    created_at   TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at   TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_mag_user (user_id),
    INDEX idx_mag_plan (plan_id)
);
