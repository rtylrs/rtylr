-- Existing installations already have flow_user_settings. Keep this additive
-- field in a separately named migration so rerunning the schema applies it.
SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
   WHERE TABLE_SCHEMA = DATABASE()
     AND TABLE_NAME = 'flow_user_settings'
     AND COLUMN_NAME = 'sidebar_preferences_json') = 0,
  'ALTER TABLE flow_user_settings ADD COLUMN sidebar_preferences_json TEXT DEFAULT NULL AFTER email_notifications',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
