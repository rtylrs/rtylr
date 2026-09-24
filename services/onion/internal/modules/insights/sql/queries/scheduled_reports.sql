-- name: ListScheduledReports :many
SELECT id, owner_id, business_id, template_id, cadence, channel, recipient, is_active, created_at, updated_at
FROM insights_scheduled_reports
WHERE business_id = ?
ORDER BY
  CASE WHEN sqlc.arg(sort_by) = 'created_at' AND sqlc.arg(sort_dir) = 'asc' THEN created_at END ASC,
  CASE WHEN sqlc.arg(sort_by) = 'created_at' AND sqlc.arg(sort_dir) = 'desc' THEN created_at END DESC,
  CASE WHEN sqlc.arg(sort_by) = 'updated_at' AND sqlc.arg(sort_dir) = 'asc' THEN updated_at END ASC,
  CASE WHEN sqlc.arg(sort_by) = 'updated_at' AND sqlc.arg(sort_dir) = 'desc' THEN updated_at END DESC,
  CASE WHEN sqlc.arg(sort_by) = 'recipient' AND sqlc.arg(sort_dir) = 'asc' THEN LOWER(recipient) END ASC,
  CASE WHEN sqlc.arg(sort_by) = 'recipient' AND sqlc.arg(sort_dir) = 'desc' THEN LOWER(recipient) END DESC,
  CASE WHEN sqlc.arg(sort_by) = 'cadence' AND sqlc.arg(sort_dir) = 'asc' THEN cadence END ASC,
  CASE WHEN sqlc.arg(sort_by) = 'cadence' AND sqlc.arg(sort_dir) = 'desc' THEN cadence END DESC,
  CASE WHEN sqlc.arg(sort_by) = 'channel' AND sqlc.arg(sort_dir) = 'asc' THEN channel END ASC,
  CASE WHEN sqlc.arg(sort_by) = 'channel' AND sqlc.arg(sort_dir) = 'desc' THEN channel END DESC,
  CASE WHEN sqlc.arg(sort_by) = 'is_active' AND sqlc.arg(sort_dir) = 'asc' THEN is_active END ASC,
  CASE WHEN sqlc.arg(sort_by) = 'is_active' AND sqlc.arg(sort_dir) = 'desc' THEN is_active END DESC,
  created_at DESC,
  id DESC
LIMIT ? OFFSET ?;

-- name: CountScheduledReports :one
SELECT COUNT(*) AS count
FROM insights_scheduled_reports
WHERE business_id = ?;

-- name: GetScheduledReport :one
SELECT id, owner_id, business_id, template_id, cadence, channel, recipient, is_active, created_at, updated_at
FROM insights_scheduled_reports
WHERE id = ? AND business_id = ?;

-- name: CreateScheduledReport :execresult
INSERT INTO insights_scheduled_reports
  (owner_id, business_id, template_id, cadence, channel, recipient, is_active)
VALUES (?, ?, ?, ?, ?, ?, COALESCE(sqlc.narg(is_active), TRUE));

-- name: UpdateScheduledReport :exec
UPDATE insights_scheduled_reports
SET template_id = ?, cadence = ?, channel = ?, recipient = ?,
    is_active = COALESCE(sqlc.narg(is_active), is_active)
WHERE id = ? AND business_id = ?;

-- name: DeleteScheduledReport :exec
DELETE FROM insights_scheduled_reports WHERE id = ? AND business_id = ?;
