-- Composite workload index for workspace task batching and roadmap reads.
SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'flow_tasks' AND INDEX_NAME = 'idx_flow_tasks_workspace_assignee_status_parent') = 0
  -- Guarded on the COLUMN as well as the index. The index guard alone passes
  -- once Phase 4 drops workspace_id (the index is gone with it) and the ALTER
  -- then fails on the missing column, which migrate.sh records as applied
  -- behind a green check.
  AND (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'flow_tasks' AND COLUMN_NAME = 'workspace_id') > 0,
  'ALTER TABLE flow_tasks ADD INDEX idx_flow_tasks_workspace_assignee_status_parent (workspace_id, assignee_id, status, parent_task_id)',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
