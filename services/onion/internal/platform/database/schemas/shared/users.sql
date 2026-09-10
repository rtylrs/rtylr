CREATE TABLE IF NOT EXISTS users (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    business_id BIGINT NULL,
    email VARCHAR(320) UNIQUE NOT NULL,
    username VARCHAR(64) NULL UNIQUE,
    password_hash VARCHAR(512) NOT NULL,
    full_name VARCHAR(255) NULL,
    signup_ip_address VARCHAR(45) NULL,
    signup_device_info TEXT NULL,
    onboarded BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_console_admin BOOLEAN NOT NULL DEFAULT FALSE,
    is_banned BOOLEAN NOT NULL DEFAULT FALSE,
    ban_reason TEXT NULL,
    banned_at TIMESTAMP NULL,
    auth_provider VARCHAR(16) NOT NULL DEFAULT 'password',
    token_epoch BIGINT NOT NULL DEFAULT 0,
    two_factor_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    INDEX idx_users_business (business_id),
    INDEX idx_users_provider (auth_provider),
    INDEX idx_users_console_admin (is_console_admin),
    INDEX idx_users_banned (is_banned)
);



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

-- A single, transactionally maintained row backs the console's global stats.
-- It avoids re-scanning every global table when an admin opens the dashboard.
CREATE TABLE IF NOT EXISTS console_stats (
    id TINYINT NOT NULL PRIMARY KEY,
    total_users BIGINT NOT NULL DEFAULT 0,
    banned_users BIGINT NOT NULL DEFAULT 0,
    total_businesses BIGINT NOT NULL DEFAULT 0,
    active_subs BIGINT NOT NULL DEFAULT 0,
    trial_subs BIGINT NOT NULL DEFAULT 0,
    manual_grants BIGINT NOT NULL DEFAULT 0,
    manual_addon_grants BIGINT NOT NULL DEFAULT 0,
    console_admins BIGINT NOT NULL DEFAULT 0,
    CHECK (id = 1)
);
