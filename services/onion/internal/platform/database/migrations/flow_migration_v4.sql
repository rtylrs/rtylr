-- Flow schema migration v4
-- Adds: project columns (custom kanban columns), section_id on tasks

CREATE TABLE IF NOT EXISTS flow_project_columns (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    project_id BIGINT NOT NULL,
    workspace_id BIGINT NOT NULL,
    name VARCHAR(255) NOT NULL,
    color VARCHAR(20) DEFAULT '#6b7280',
    sort_order INT NOT NULL DEFAULT 0,
    created_by BIGINT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_project_columns_project (project_id),
    INDEX idx_project_columns_workspace (workspace_id)
);

-- Add section_id to flow_tasks (nullable — tasks without a section float to first column)
ALTER TABLE flow_tasks
    ADD COLUMN section_id BIGINT DEFAULT NULL AFTER parent_task_id,
    ADD INDEX idx_tasks_section (section_id);
