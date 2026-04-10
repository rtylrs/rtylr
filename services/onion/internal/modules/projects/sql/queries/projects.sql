-- name: ListProjectsByOwner :many
SELECT id, owner_id, workspace_id, name, description, status, color, created_at, updated_at
FROM flow_projects WHERE owner_id = ? ORDER BY created_at DESC;

-- name: CreateProject :execresult
INSERT INTO flow_projects (owner_id, workspace_id, name, description, status, color)
VALUES (?, ?, ?, ?, 'active', ?);

-- name: UpdateProject :execresult
UPDATE flow_projects SET workspace_id = ?, name = ?, description = ?, status = ?, color = ?
WHERE id = ? AND owner_id = ?;

-- name: DeleteProject :execresult
DELETE FROM flow_projects WHERE id = ? AND owner_id = ?;
