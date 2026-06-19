-- name: ListNotifications :many
SELECT id, owner_id, business_id, user_id, title, COALESCE(body,'') AS body, type, is_read, COALESCE(link,'') AS link, created_at
FROM hub_notifications
WHERE business_id = ? AND user_id = ?
ORDER BY created_at DESC
LIMIT 100;

-- name: MarkNotificationRead :exec
UPDATE hub_notifications SET is_read = TRUE WHERE id = ? AND user_id = ?;

-- name: MarkAllNotificationsRead :exec
UPDATE hub_notifications SET is_read = TRUE WHERE user_id = ? AND is_read = FALSE;
