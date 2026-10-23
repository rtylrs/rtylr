-- name: GetUserByEmail :one
SELECT id, email, full_name, business_id, is_console_admin,
       is_banned, ban_reason, banned_at, created_at, password_hash
FROM users
WHERE email = ?
ORDER BY email ASC
LIMIT 1;

-- name: GetUserByID :one
SELECT id, email, full_name, business_id, is_console_admin,
       is_banned, ban_reason, banned_at, created_at, password_hash
FROM users
WHERE id = ?
ORDER BY id ASC
LIMIT 1;

-- name: CreateUser :execresult
INSERT INTO users (email, password_hash, full_name, is_console_admin)
VALUES (?, ?, ?, ?);

-- name: UpdateUser :exec
UPDATE users
SET full_name = ?, email = ?, updated_at = NOW()
WHERE id = ?;

-- name: BanUser :exec
UPDATE users
SET is_banned = TRUE, ban_reason = ?, banned_at = NOW()
WHERE id = ?;

-- name: UnbanUser :exec
UPDATE users
SET is_banned = FALSE, ban_reason = NULL, banned_at = NULL
WHERE id = ?;

-- name: SetConsoleAdmin :exec
UPDATE users SET is_console_admin = ? WHERE id = ?;
