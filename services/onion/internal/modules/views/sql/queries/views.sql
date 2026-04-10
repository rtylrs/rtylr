-- name: ListViewsByWorkspace :many
SELECT id, owner_id, workspace_id, name, filters, sort_by, group_by, is_default, created_at, updated_at
FROM flow_views WHERE owner_id = ? AND workspace_id = ? ORDER BY is_default DESC, name ASC;

-- name: CreateView :execresult
INSERT INTO flow_views (owner_id, workspace_id, name, filters, sort_by, group_by, is_default)
VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: UpdateView :execresult
UPDATE flow_views SET name = ?, filters = ?, sort_by = ?, group_by = ?, is_default = ?
WHERE id = ? AND owner_id = ?;

-- name: DeleteView :execresult
DELETE FROM flow_views WHERE id = ? AND owner_id = ?;
