-- Business-scoped projection of this module's tables for the AI analyst (aiquery).
-- Published here so the owner controls the columns aiquery may read.
-- owner_id is shared by every business of an account, so scope by business_id.
-- Kept as its own zz_ file: migrate.sh records applied files by name.

CREATE OR REPLACE VIEW ai_tasks AS
SELECT id, owner_id, business_id, project_id, workspace_id, title, status, priority,
       assignee_id, due_date, start_date, completed_at,
       estimated_hours, actual_hours, created_at, updated_at
FROM flow_tasks;
