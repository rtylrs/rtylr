-- F-A: per-project custom statuses.
--
-- Every statement is scoped by business_id AND the resource id. There are no
-- flow_workspaces or flow_workspace_members EXISTS clauses anywhere in this
-- file, deliberately: those clauses are what made SetTaskLabels impossible on
-- any v2-created task, because a v2 task has workspace_id = 0 so neither EXISTS
-- could ever be true. Authorization is the caller's job through
-- RequireProjectRole and RequireTaskRole, which resolve through the project.

-- name: ListStatusesByProject :many
SELECT id, business_id, project_id, name, bucket, legacy_status, color, sort_order, is_default, created_by, created_at
FROM flow_statuses
WHERE business_id = ? AND project_id = ?
ORDER BY sort_order ASC, id ASC;

-- name: GetStatusByID :one
SELECT id, business_id, project_id, name, bucket, legacy_status, color, sort_order, is_default, created_by, created_at
FROM flow_statuses
WHERE id = ? AND business_id = ?;

-- name: CreateStatus :execresult
INSERT INTO flow_statuses
  (business_id, project_id, name, bucket, legacy_status, color, sort_order, is_default, created_by)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: UpdateStatus :execresult
UPDATE flow_statuses
SET name = ?, bucket = ?, legacy_status = ?, color = ?, sort_order = ?, is_default = ?
WHERE id = ? AND business_id = ?;

-- name: DeleteStatus :execresult
DELETE FROM flow_statuses
WHERE id = ? AND business_id = ?;

-- name: ClearProjectDefaultStatus :execresult
-- is_default is at most one per project. MySQL cannot express that as a partial
-- unique index, so the write path clears the flag across the project before
-- setting it, inside the same transaction.
UPDATE flow_statuses
SET is_default = FALSE
WHERE business_id = ? AND project_id = ?;

-- name: NextStatusSortOrder :one
-- Appends when the caller does not supply a position. COALESCE because MAX over
-- an empty set is NULL, and the first status in a project has to land at 0.
-- CAST so sqlc types the result. Without it COALESCE over MAX yields an
-- untyped expression and the generated method returns interface{}, which
-- pushes a type assertion into Go for no reason.
SELECT CAST(COALESCE(MAX(sort_order) + 1, 0) AS SIGNED) FROM flow_statuses
WHERE business_id = ? AND project_id = ?;

-- name: ClearTasksOfStatus :execresult
-- Deleting a status never blocks on its tasks and never deletes them. Their
-- status_id is cleared and the built-in string status they already carry keeps
-- them exactly where the user left them.
UPDATE flow_tasks
SET status_id = NULL
WHERE business_id = ? AND status_id = ?;

-- name: SetTaskStatusID :execresult
-- Writes both columns together. The string column is the one every existing
-- query, filter and index reads, so letting the two drift is how a task would
-- show one status on the board and another in a report.
UPDATE flow_tasks
SET status_id = ?, status = ?
WHERE id = ? AND business_id = ?;

-- name: CountTasksOnStatus :one
SELECT COUNT(*) FROM flow_tasks
WHERE business_id = ? AND status_id = ?;
