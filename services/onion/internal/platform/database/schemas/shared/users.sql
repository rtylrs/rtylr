CREATE TABLE IF NOT EXISTS users (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    business_id BIGINT NULL,
    email VARCHAR(320) UNIQUE NOT NULL,
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
    INDEX idx_users_business (business_id),
    INDEX idx_users_provider (auth_provider),
    INDEX idx_users_console_admin (is_console_admin),
    INDEX idx_users_banned (is_banned)
);

-- Backfill columns/indexes on pre-existing prod tables (CREATE TABLE IF NOT EXISTS is a no-op on existing tables).
DROP PROCEDURE IF EXISTS add_users_column_if_missing;
DELIMITER $$
CREATE PROCEDURE add_users_column_if_missing(IN col_name VARCHAR(64), IN col_def TEXT)
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM INFORMATION_SCHEMA.COLUMNS
        WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'users' AND COLUMN_NAME = col_name
    ) THEN
        SET @sql := CONCAT('ALTER TABLE users ADD COLUMN ', col_name, ' ', col_def);
        PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
    END IF;
END$$
DELIMITER ;

CALL add_users_column_if_missing('signup_ip_address',  'VARCHAR(45) NULL');
CALL add_users_column_if_missing('signup_device_info', 'TEXT NULL');
CALL add_users_column_if_missing('onboarded',          'BOOLEAN DEFAULT FALSE');
CALL add_users_column_if_missing('is_console_admin',   'BOOLEAN NOT NULL DEFAULT FALSE');
CALL add_users_column_if_missing('is_banned',          'BOOLEAN NOT NULL DEFAULT FALSE');
CALL add_users_column_if_missing('ban_reason',         'TEXT NULL');
CALL add_users_column_if_missing('banned_at',          'TIMESTAMP NULL');
CALL add_users_column_if_missing('auth_provider',      "VARCHAR(16) NOT NULL DEFAULT 'password'");

DROP PROCEDURE IF EXISTS add_users_column_if_missing;

DROP PROCEDURE IF EXISTS add_users_index_if_missing;
DELIMITER $$
CREATE PROCEDURE add_users_index_if_missing(IN idx_name VARCHAR(64), IN idx_cols TEXT)
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM INFORMATION_SCHEMA.STATISTICS
        WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'users' AND INDEX_NAME = idx_name
    ) THEN
        SET @sql := CONCAT('ALTER TABLE users ADD INDEX ', idx_name, ' ', idx_cols);
        PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
    END IF;
END$$
DELIMITER ;

CALL add_users_index_if_missing('idx_users_business',      '(business_id)');
CALL add_users_index_if_missing('idx_users_provider',      '(auth_provider)');
CALL add_users_index_if_missing('idx_users_console_admin', '(is_console_admin)');
CALL add_users_index_if_missing('idx_users_banned',        '(is_banned)');

DROP PROCEDURE IF EXISTS add_users_index_if_missing;

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
