-- name: GetEvent :one
SELECT id, business_id, event_type, subject_type, subject_id, payload, occurred_at
FROM platform_events
WHERE id = ? AND business_id = ?;
