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
SELECT id, email, password_hash, full_name, signup_ip_address, signup_device_info, created_at, updated_at, is_console_admin
FROM users WHERE id = ?;

-- name: GetUserByEmail :one
SELECT id, email, password_hash, full_name, signup_ip_address, signup_device_info, created_at, updated_at, is_console_admin
FROM users WHERE email = ?;

-- name: EmailExists :one
SELECT EXISTS(SELECT 1 FROM users WHERE email = ?);

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

-- name: GetUserBusinessID :one
SELECT business_id FROM users WHERE id = ?;

-- name: SetUserBusinessID :exec
UPDATE users SET business_id = ? WHERE id = ?;

-- name: GetBusinessOwnerID :one
SELECT owner_id FROM businesses WHERE id = ?;

-- name: GetUserAppRole :one
SELECT role FROM user_app_permissions WHERE business_id = ? AND user_id = ?;

-- name: CreateBusiness :execresult
INSERT INTO businesses (name, owner_id) VALUES (?, ?);

-- name: GetActivePlanForSignup :one
SELECT id, trial_days FROM plans WHERE slug = ? AND is_active = TRUE;

-- name: UpsertSignupSubscription :exec
INSERT INTO subscriptions (owner_id, plan_id, status, trial_ends_at)
VALUES (?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
  plan_id = VALUES(plan_id),
  status = VALUES(status),
  trial_ends_at = VALUES(trial_ends_at),
  updated_at = CURRENT_TIMESTAMP;
