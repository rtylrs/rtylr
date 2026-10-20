-- Activities live in crm_touchpoints. A deal's timeline includes the work done
-- on the lead it came from.

-- name: CreateActivity :execresult
INSERT INTO crm_touchpoints (owner_id, business_id, lead_id, deal_id, customer_id, type, outcome, notes,
                             duration_seconds, created_by, occurred_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetActivity :one
SELECT * FROM crm_touchpoints WHERE id = ? AND business_id = ?;

-- name: DeleteActivity :execresult
DELETE FROM crm_touchpoints WHERE id = ? AND business_id = ?;

-- name: ListLeadActivities :many
SELECT * FROM crm_touchpoints
WHERE business_id = sqlc.arg('business_id') AND lead_id = sqlc.arg('lead_id')
  AND (sqlc.narg('cursor_id') IS NULL OR occurred_at < sqlc.narg('cursor_time')
       OR (occurred_at = sqlc.narg('cursor_time') AND id < sqlc.narg('cursor_id')))
ORDER BY occurred_at DESC, id DESC
LIMIT ?;

-- name: ListDealActivities :many
SELECT * FROM crm_touchpoints
WHERE business_id = sqlc.arg('business_id')
  AND (deal_id = sqlc.arg('deal_id') OR (sqlc.narg('lead_id') IS NOT NULL AND lead_id = sqlc.narg('lead_id')))
  AND (sqlc.narg('cursor_id') IS NULL OR occurred_at < sqlc.narg('cursor_time')
       OR (occurred_at = sqlc.narg('cursor_time') AND id < sqlc.narg('cursor_id')))
ORDER BY occurred_at DESC, id DESC
LIMIT ?;

-- name: CountCallsSince :one
SELECT COUNT(*) FROM crm_touchpoints
WHERE business_id = sqlc.arg('business_id') AND type = 'call' AND occurred_at >= sqlc.arg('since')
  AND (sqlc.narg('created_by') IS NULL OR created_by = sqlc.narg('created_by'));

-- name: CountConnectedCallsSince :one
SELECT COUNT(*) FROM crm_touchpoints
WHERE business_id = sqlc.arg('business_id') AND type = 'call' AND occurred_at >= sqlc.arg('since')
  AND outcome IN ('connected', 'callback', 'meeting_booked', 'not_interested')
  AND (sqlc.narg('created_by') IS NULL OR created_by = sqlc.narg('created_by'));

-- name: DeleteLeadActivities :exec
DELETE FROM crm_touchpoints WHERE business_id = sqlc.arg('business_id') AND lead_id IN (sqlc.slice('lead_ids'));

-- name: DeleteDealActivities :exec
DELETE FROM crm_touchpoints WHERE business_id = ? AND deal_id = ?;
