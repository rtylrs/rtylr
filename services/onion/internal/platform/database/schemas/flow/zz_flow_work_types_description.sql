-- flow_work_types already existed in every deployed environment before
-- description was added to the contract. Give it its own filename so
-- migrate.sh (which records schema files as applied by filename) treats
-- this as new and actually runs it, rather than silently no-op'ing on an
-- edit to the already-applied v8_time_tracking.sql.
SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'flow_work_types' AND COLUMN_NAME = 'description') = 0,
  'ALTER TABLE flow_work_types ADD COLUMN description TEXT AFTER name',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
