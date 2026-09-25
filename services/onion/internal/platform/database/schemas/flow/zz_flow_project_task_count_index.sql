-- Supports page-scoped project task and completed-task aggregates.
SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'flow_tasks' AND INDEX_NAME = 'idx_flow_tasks_project_status') = 0,
  'ALTER TABLE flow_tasks ADD INDEX idx_flow_tasks_project_status (project_id, status)',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
