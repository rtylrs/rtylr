-- name: CreateTaskLink :execresult
INSERT INTO flow_task_links (task_id, business_id, url, kind, title, ref, state, added_by)
VALUES (?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetTaskLink :one
SELECT id, task_id, business_id, url, kind, title, ref, state, added_by, created_at
FROM flow_task_links WHERE id = ? AND business_id = ?;

-- name: FindTaskLinkByURL :one
SELECT id, task_id, business_id, url, kind, title, ref, state, added_by, created_at
FROM flow_task_links WHERE task_id = ? AND business_id = ? AND url = ?
LIMIT 1;

-- name: ListTaskLinks :many
SELECT id, task_id, business_id, url, kind, title, ref, state, added_by, created_at
FROM flow_task_links WHERE task_id = ? AND business_id = ?
ORDER BY created_at DESC, id DESC LIMIT 200;

-- name: DeleteTaskLink :exec
DELETE FROM flow_task_links WHERE id = ? AND business_id = ?;
