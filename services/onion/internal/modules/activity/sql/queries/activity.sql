-- name: ListActivity :many
SELECT a.id, a.owner_id, a.business_id, a.app, a.entity_type, a.entity_id, a.action, a.description, a.user_id, a.created_at
FROM activity_log a
WHERE a.business_id = ?
  AND (? = '' OR a.app = ?)
ORDER BY a.created_at DESC
LIMIT ? OFFSET ?;

-- name: CountActivity :one
SELECT COUNT(*) AS total
FROM activity_log a
WHERE a.business_id = ?
  AND (? = '' OR a.app = ?);

-- name: RecordMutation :exec
INSERT INTO activity_log (owner_id, business_id, app, entity_type, entity_id, action, description, user_id, old_values, new_values, ip_address, created_at, source_event_id)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE id = id;
