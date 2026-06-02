-- name: ListMilestonesByOwner :many
SELECT id, owner_id, project_id, name, description, target_date, status, created_at, updated_at
FROM flow_milestones WHERE owner_id = ? ORDER BY target_date ASC, created_at ASC;

-- name: ListMilestonesByProject :many
SELECT id, owner_id, project_id, name, description, target_date, status, created_at, updated_at
FROM flow_milestones WHERE owner_id = ? AND project_id = ? ORDER BY target_date ASC, created_at ASC;

-- name: GetProjectByIDAndOwner :one
SELECT id FROM flow_projects WHERE id = ? AND owner_id = ?;

-- name: CreateMilestone :execresult
INSERT INTO flow_milestones (owner_id, project_id, name, description, target_date)
VALUES (?, ?, ?, ?, ?);

-- name: UpdateMilestone :execresult
UPDATE flow_milestones SET name = ?, description = ?, target_date = ?, status = ?
WHERE id = ? AND owner_id = ?;
