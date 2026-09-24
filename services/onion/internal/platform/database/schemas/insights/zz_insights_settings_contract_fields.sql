-- Persist the settings fields declared by insights.setting.v1. MySQL does not
-- support ADD COLUMN IF NOT EXISTS, so keep this migration replay-safe.

SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
   WHERE TABLE_SCHEMA = DATABASE()
     AND TABLE_NAME = 'insights_settings'
     AND COLUMN_NAME = 'default_period') = 0,
  'ALTER TABLE insights_settings ADD COLUMN default_period VARCHAR(20) NOT NULL DEFAULT ''this_month'' AFTER timezone',
  'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
   WHERE TABLE_SCHEMA = DATABASE()
     AND TABLE_NAME = 'insights_settings'
     AND COLUMN_NAME = 'compare_previous') = 0,
  'ALTER TABLE insights_settings ADD COLUMN compare_previous BOOLEAN NOT NULL DEFAULT FALSE AFTER default_period',
  'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
   WHERE TABLE_SCHEMA = DATABASE()
     AND TABLE_NAME = 'insights_settings'
     AND COLUMN_NAME = 'dashboard_widgets') = 0,
  'ALTER TABLE insights_settings ADD COLUMN dashboard_widgets JSON NOT NULL DEFAULT (JSON_ARRAY()) AFTER compare_previous',
  'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
   WHERE TABLE_SCHEMA = DATABASE()
     AND TABLE_NAME = 'insights_settings'
     AND COLUMN_NAME = 'favorite_reports') = 0,
  'ALTER TABLE insights_settings ADD COLUMN favorite_reports JSON NOT NULL DEFAULT (JSON_ARRAY()) AFTER dashboard_widgets',
  'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
