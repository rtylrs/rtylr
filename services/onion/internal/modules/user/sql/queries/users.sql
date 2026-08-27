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
SELECT id, business_id, email, password_hash, full_name, signup_ip_address, signup_device_info, created_at, updated_at, is_console_admin, is_banned, token_epoch
FROM users WHERE id = ?;

-- name: GetUserByEmail :one
SELECT id, business_id, email, password_hash, full_name, signup_ip_address, signup_device_info, created_at, updated_at, is_console_admin, is_banned, token_epoch
FROM users WHERE email = ?;

-- name: GetUserByIdentifier :one
SELECT id, email, password_hash, full_name, signup_ip_address, signup_device_info, created_at, updated_at, is_console_admin, token_epoch
FROM users
WHERE email = ? OR LOWER(username) = LOWER(?)
ORDER BY id ASC
LIMIT 1;

-- name: EmailExists :one
SELECT EXISTS(SELECT 1 FROM users WHERE email = ?);

-- name: UpdateUserProfile :exec
UPDATE users
SET full_name = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: UpdateUserPassword :exec
-- Setting a password bumps token_epoch so every JWT issued before it stops
-- validating (server-side revocation for stolen-token scenarios).
UPDATE users
SET password_hash = ?, token_epoch = token_epoch + 1, updated_at = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = ?;

-- name: RecordLoginHistory :exec
INSERT INTO login_history (user_id, email, method, ip_address, user_agent, success) VALUES (?, ?, ?, ?, ?, ?);

-- name: GetUserBusinessID :one
SELECT business_id FROM users WHERE id = ?;

-- name: SetUserBusinessID :exec
UPDATE users SET business_id = ? WHERE id = ?;

-- name: GetBusinessOwnerID :one
SELECT owner_id FROM businesses WHERE id = ? AND is_active = TRUE LIMIT 1;

-- name: GetUserAppRole :one
SELECT role FROM user_app_permissions
WHERE business_id = ? AND user_id = ? AND COALESCE(is_active, TRUE) = TRUE;

-- Read authentication state without the application cache so deactivation and
-- bans take effect across processes without waiting for a profile-cache TTL.
-- name: GetUserAccessState :one
SELECT
    EXISTS(
        SELECT 1 FROM user_app_permissions p
        WHERE p.user_id = u.id
    ) AS has_permissions,
    EXISTS(
        SELECT 1 FROM user_app_permissions p
        WHERE p.user_id = u.id AND COALESCE(p.is_active, TRUE) = TRUE
    ) OR EXISTS(
        SELECT 1 FROM businesses b
        WHERE b.owner_id = u.id AND COALESCE(b.is_active, TRUE) = TRUE
    ) AS has_active_access
FROM users u
WHERE u.id = ?;

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

-- cross-schema write to hr_core — every user is also an hr_employees row
-- name: EnsureEmployeeForUser :exec
INSERT INTO hr_employees (owner_id, business_id, user_id, first_name, last_name, email, status)
SELECT ?, COALESCE(u.business_id, 0), u.id,
       COALESCE(NULLIF(SUBSTRING_INDEX(u.full_name, ' ', 1), ''), u.email),
       CASE WHEN LOCATE(' ', u.full_name) > 0 THEN SUBSTRING(u.full_name, LOCATE(' ', u.full_name) + 1) ELSE '' END,
       u.email, 'active'
FROM users u
WHERE u.id = ?
  AND NOT EXISTS (SELECT 1 FROM hr_employees e WHERE e.user_id = u.id);

-- name: BusinessNameExists :one
SELECT EXISTS(SELECT 1 FROM businesses WHERE name = ? LIMIT 1) AS exists_flag;

-- cross-schema read from user_business_access — multi-business switch validation
-- name: CheckBusinessAccess :one
SELECT COUNT(*) AS cnt FROM (
    SELECT 1 FROM businesses b WHERE b.id = ? AND b.owner_id = ? AND b.is_active = TRUE
    UNION ALL
    SELECT 1 FROM user_business_access uba WHERE uba.user_id = ? AND uba.business_id = ?
) x;
