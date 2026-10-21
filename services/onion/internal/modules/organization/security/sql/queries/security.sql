-- name: GetTwoFactorEnabled :one
SELECT COALESCE(two_factor_enabled, FALSE) FROM users WHERE id = ?;

-- name: SetTwoFactor :exec
UPDATE users SET two_factor_enabled = ? WHERE id = ?;

-- name: ListLoginHistory :many
SELECT id, email, method, COALESCE(ip_address,''), COALESCE(user_agent,''), success, created_at
FROM login_history
WHERE user_id = ?
ORDER BY created_at DESC, id DESC
LIMIT ? OFFSET ?;

-- name: CountLoginHistory :one
SELECT COUNT(*) FROM login_history WHERE user_id = ?;

-- name: InvalidateAllSessions :exec
UPDATE user_sessions SET revoked_at = NOW() WHERE user_id = ? AND revoked_at IS NULL;

-- name: RecordLogin :exec
INSERT INTO login_history (user_id, email, method, ip_address, user_agent, success)
VALUES (?, ?, ?, ?, ?, ?);
