-- name: ListActivity :many
SELECT a.id, a.owner_id, a.business_id, a.app, a.entity_type, a.entity_id, a.action, a.description, a.user_id, u.full_name AS user_name, a.created_at
FROM activity_log a
LEFT JOIN users u ON u.id = a.user_id
WHERE a.business_id = ?
  AND (? = '' OR a.app = ?)
ORDER BY a.created_at DESC
LIMIT ? OFFSET ?;

-- name: CountActivity :one
SELECT COUNT(*) AS total
FROM activity_log a
WHERE a.business_id = ?
  AND (? = '' OR a.app = ?);
