-- The read model of Flow tasks that the AI query module (Caret analyst mode)
-- may query, published by Flow because Flow owns flow_tasks. It used to be
-- defined in aiquery's own schema, which read another module's table. Same
-- columns as before, scoped by business_id (see aiquery's
-- zz_ai_views_business_scope.sql), so existing queries see no change.
CREATE OR REPLACE VIEW ai_tasks AS
SELECT id, owner_id, business_id, project_id, workspace_id, title, status, priority,
       assignee_id, due_date, start_date, completed_at,
       estimated_hours, actual_hours, created_at, updated_at
FROM flow_tasks;
