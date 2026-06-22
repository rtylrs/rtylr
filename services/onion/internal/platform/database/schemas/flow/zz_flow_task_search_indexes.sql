-- Adds indexed Flow task search support.
-- Duplicate-index errors are safe under the migration runner's --force mode.

ALTER TABLE flow_tasks ADD COLUMN archived BOOLEAN DEFAULT FALSE;
ALTER TABLE flow_tasks ADD INDEX idx_flow_tasks_workspace_updated (workspace_id, archived, updated_at);
ALTER TABLE flow_tasks ADD FULLTEXT INDEX idx_flow_tasks_search (title, description);
