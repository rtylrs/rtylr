-- name: ListTasksByOwner :many
SELECT id, owner_id, project_id, workspace_id, title, description, status, priority, assignee_id, labels, due_date, created_at, updated_at
FROM flow_tasks WHERE owner_id = ? ORDER BY created_at DESC;

-- name: CreateTask :execresult
INSERT INTO flow_tasks (owner_id, project_id, workspace_id, title, description, status, priority, assignee_id, labels, due_date)
VALUES (?, ?, ?, ?, ?, 'todo', ?, ?, ?, ?);

-- name: UpdateTask :execresult
UPDATE flow_tasks SET project_id = ?, workspace_id = ?, title = ?, description = ?, status = ?, priority = ?, assignee_id = ?, labels = ?, due_date = ?
WHERE id = ? AND owner_id = ?;

-- name: DeleteTask :execresult
DELETE FROM flow_tasks WHERE id = ? AND owner_id = ?;

-- name: ListLabelsByOwner :many
SELECT id, workspace_id, name, color, created_at FROM flow_labels WHERE workspace_id = ? ORDER BY name;

-- name: CreateLabel :execresult
INSERT INTO flow_labels (workspace_id, name, color) VALUES (?, ?, ?);
