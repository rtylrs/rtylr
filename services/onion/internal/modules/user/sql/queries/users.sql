-- name: CreateUser :execresult
INSERT INTO users (
    email, password_hash, full_name, signup_ip_address, signup_device_info
) VALUES (
    ?, ?, ?, ?, ?
);

-- Explicit column lists (never SELECT *): the query selects only the columns
-- the application reads, so adding a column to the users schema can never break
-- these queries against a database that has not been migrated yet.
-- name: GetUserByID :one
SELECT id, email, password_hash, full_name, signup_ip_address, signup_device_info, created_at, updated_at
FROM users WHERE id = ?;

-- name: GetUserByEmail :one
SELECT id, email, password_hash, full_name, signup_ip_address, signup_device_info, created_at, updated_at
FROM users WHERE email = ?;

-- name: UpdateUserProfile :exec
UPDATE users
SET full_name = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: UpdateUserPassword :exec
UPDATE users
SET password_hash = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = ?;

-- name: RecordLoginHistory :exec
INSERT INTO login_history (user_id, email, method, ip_address, success) VALUES (?, ?, ?, ?, ?);
