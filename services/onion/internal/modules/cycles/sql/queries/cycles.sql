-- name: ListCyclesByOwner :many
SELECT id, owner_id, workspace_id, name, start_date, end_date, status, created_at, updated_at
FROM flow_cycles WHERE owner_id = ? ORDER BY start_date ASC, created_at ASC;

-- name: ListCyclesByWorkspace :many
SELECT id, owner_id, workspace_id, name, start_date, end_date, status, created_at, updated_at
FROM flow_cycles WHERE owner_id = ? AND workspace_id = ? ORDER BY start_date ASC, created_at ASC;

-- name: CreateCycle :execresult
INSERT INTO flow_cycles (owner_id, workspace_id, name, start_date, end_date)
VALUES (?, ?, ?, ?, ?);

-- name: UpdateCycle :execresult
UPDATE flow_cycles SET name = ?, start_date = ?, end_date = ?, status = ?
WHERE id = ? AND owner_id = ?;
