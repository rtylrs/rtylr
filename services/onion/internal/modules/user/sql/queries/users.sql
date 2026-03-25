-- name: CreateUser :one
INSERT INTO users (
    email, password_hash, full_name, signup_ip_address, signup_device_info
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: UpdateUserProfile :one
UPDATE users
SET full_name = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateUserPassword :exec
UPDATE users
SET password_hash = $2, updated_at = NOW()
WHERE id = $1;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1;
