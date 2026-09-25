-- Business-leading Flow Insights indexes. They cover completion windows,
-- workload grouping, and roadmap project/task joins without replacing the
-- existing workspace-oriented interactive-board indexes. Roll back individual
-- keys only after checking live plans with DROP INDEX <name> ON the table.
SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'flow_projects' AND INDEX_NAME = 'idx_flow_projects_business_created') = 0,
  'ALTER TABLE flow_projects ADD INDEX idx_flow_projects_business_created (business_id, created_at, id)',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'flow_tasks' AND INDEX_NAME = 'idx_flow_tasks_business_status_completed') = 0,
  'ALTER TABLE flow_tasks ADD INDEX idx_flow_tasks_business_status_completed (business_id, status, completed_at)',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'flow_tasks' AND INDEX_NAME = 'idx_flow_tasks_business_assignee') = 0,
  'ALTER TABLE flow_tasks ADD INDEX idx_flow_tasks_business_assignee (business_id, assignee_id)',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'flow_tasks' AND INDEX_NAME = 'idx_flow_tasks_project_parent') = 0,
  'ALTER TABLE flow_tasks ADD INDEX idx_flow_tasks_project_parent (project_id, parent_task_id)',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
