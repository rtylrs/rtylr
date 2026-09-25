-- Audit snapshots were added to the writer before this shared table's schema
-- was updated.  Use guarded ALTER statements so the migration is safe on both
-- fresh databases and databases created from the older schema.
SET @sql := IF(
    (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
     WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'activity_log'
       AND COLUMN_NAME = 'old_values') = 0,
    'ALTER TABLE activity_log ADD COLUMN old_values JSON NULL AFTER user_id',
    'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @sql := IF(
    (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
     WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'activity_log'
       AND COLUMN_NAME = 'new_values') = 0,
    'ALTER TABLE activity_log ADD COLUMN new_values JSON NULL AFTER old_values',
    'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @sql := IF(
    (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
     WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'activity_log'
       AND COLUMN_NAME = 'ip_address') = 0,
    'ALTER TABLE activity_log ADD COLUMN ip_address VARCHAR(45) NULL AFTER new_values',
    'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
