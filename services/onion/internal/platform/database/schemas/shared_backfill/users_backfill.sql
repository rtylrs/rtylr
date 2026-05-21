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
