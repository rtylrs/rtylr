-- Composite workload index for workspace task batching and roadmap reads.
SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'flow_tasks' AND INDEX_NAME = 'idx_flow_tasks_workspace_assignee_status_parent') = 0,
  'ALTER TABLE flow_tasks ADD INDEX idx_flow_tasks_workspace_assignee_status_parent (workspace_id, assignee_id, status, parent_task_id)',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
