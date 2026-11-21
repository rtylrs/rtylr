-- Ensure hr_shifts carries the template columns and nullable staff_id/date on
-- databases created before the unified hr_shifts schema. Safe to re-run.

SET @sql := IF(
    EXISTS (
        SELECT 1 FROM INFORMATION_SCHEMA.COLUMNS
        WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'hr_shifts' AND COLUMN_NAME = 'location_id'
    ),
    'SELECT 1',
    'ALTER TABLE hr_shifts ADD COLUMN location_id BIGINT DEFAULT NULL'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @sql := IF(
    EXISTS (
        SELECT 1 FROM INFORMATION_SCHEMA.COLUMNS
        WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'hr_shifts' AND COLUMN_NAME = 'name'
    ),
    'SELECT 1',
    'ALTER TABLE hr_shifts ADD COLUMN name VARCHAR(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @sql := IF(
    EXISTS (
        SELECT 1 FROM INFORMATION_SCHEMA.COLUMNS
        WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'hr_shifts' AND COLUMN_NAME = 'color'
    ),
    'SELECT 1',
    'ALTER TABLE hr_shifts ADD COLUMN color VARCHAR(10) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT ''#4CAF50'''
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

ALTER TABLE hr_shifts
    MODIFY COLUMN staff_id BIGINT DEFAULT NULL,
    MODIFY COLUMN date DATE DEFAULT NULL,
    MODIFY COLUMN start_time TIME NOT NULL DEFAULT '00:00:00',
    MODIFY COLUMN end_time TIME NOT NULL DEFAULT '00:00:00';
