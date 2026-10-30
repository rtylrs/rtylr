-- name: GetLayout :one
SELECT layout FROM portal_preferences WHERE user_id = ?;

-- name: EnsureLayout :exec
INSERT IGNORE INTO portal_preferences (user_id, layout) VALUES (?, '{}');

-- name: LockLayout :one
SELECT layout FROM portal_preferences WHERE user_id = ? FOR UPDATE;

-- name: SetLayout :exec
UPDATE portal_preferences SET layout = ? WHERE user_id = ?;
