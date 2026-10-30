-- name: GetLayout :one
SELECT layout FROM portal_preferences WHERE user_id = ?;

-- name: SaveLayout :exec
INSERT INTO portal_preferences (user_id, layout) VALUES (?, ?)
ON DUPLICATE KEY UPDATE layout = VALUES(layout);

-- name: DeleteLayout :exec
DELETE FROM portal_preferences WHERE user_id = ?;
