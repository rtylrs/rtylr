-- name: ListScheduledReports :many
SELECT id, owner_id, business_id, template_id, cadence, channel, recipient, is_active, created_at, updated_at
FROM insights_scheduled_reports
WHERE business_id = ?
ORDER BY created_at DESC;

-- name: GetScheduledReport :one
SELECT id, owner_id, business_id, template_id, cadence, channel, recipient, is_active, created_at, updated_at
FROM insights_scheduled_reports
WHERE id = ? AND business_id = ?;

-- name: CreateScheduledReport :execresult
INSERT INTO insights_scheduled_reports
  (owner_id, business_id, template_id, cadence, channel, recipient, is_active)
VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: UpdateScheduledReport :exec
UPDATE insights_scheduled_reports
SET template_id = ?, cadence = ?, channel = ?, recipient = ?, is_active = ?
WHERE id = ? AND business_id = ?;

-- name: DeleteScheduledReport :exec
DELETE FROM insights_scheduled_reports WHERE id = ? AND business_id = ?;
