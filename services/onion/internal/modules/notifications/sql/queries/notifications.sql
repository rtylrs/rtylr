-- name: ListNotifications :many
SELECT id, owner_id, business_id, user_id, title, COALESCE(body,'') AS body, type, is_read, COALESCE(link,'') AS link, created_at
FROM hub_notifications
WHERE business_id = ? AND user_id = ?
  AND (
    sqlc.narg('cursor_created_at') IS NULL
    OR created_at < sqlc.narg('cursor_created_at')
    OR (created_at = sqlc.narg('cursor_created_at') AND id < sqlc.narg('cursor_id'))
  )
ORDER BY created_at DESC, id DESC
LIMIT ?;

-- name: MarkNotificationRead :exec
UPDATE hub_notifications
SET is_read = TRUE
WHERE id = ? AND business_id = ? AND user_id = ?;

-- name: MarkAllNotificationsRead :exec
UPDATE hub_notifications
SET is_read = TRUE
WHERE business_id = ? AND user_id = ? AND is_read = FALSE;
