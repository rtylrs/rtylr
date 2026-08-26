-- name: GetPermission :one
SELECT id, owner_id, user_id, apps, role, is_active, created_at, updated_at
FROM user_app_permissions
WHERE owner_id = ? AND user_id = ?;

-- name: GetPermissionByBusiness :one
SELECT id, owner_id, user_id, apps, role, is_active, created_at, updated_at
FROM user_app_permissions
WHERE business_id = ? AND user_id = ?;

-- name: GetPermissionByUserID :one
SELECT id, owner_id, user_id, apps, role, is_active, created_at, updated_at
FROM user_app_permissions
WHERE user_id = ?
LIMIT 1;

-- name: ListPermissions :many
SELECT id, owner_id, user_id, apps, role, is_active, created_at, updated_at
FROM user_app_permissions
WHERE owner_id = ?
ORDER BY created_at DESC LIMIT 100;

-- name: UpsertPermission :exec
INSERT INTO user_app_permissions (business_id, owner_id, user_id, apps, role)
VALUES (?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
    business_id = VALUES(business_id),
    apps = VALUES(apps),
    role = VALUES(role),
    updated_at = NOW();

-- name: DeactivateUser :exec
UPDATE user_app_permissions
SET is_active = FALSE, updated_at = NOW()
WHERE owner_id = ? AND user_id = ?;

-- name: GetUserByEmail :one
SELECT id, email, COALESCE(full_name,'') AS full_name FROM users WHERE email = ? LIMIT 1;

-- name: GetUserByID :one
SELECT email, COALESCE(full_name,'') AS full_name, is_console_admin FROM users WHERE id = ?;

-- name: GetUserBusinessID :one
SELECT business_id FROM users WHERE id = ?;

-- name: CreateUser :execresult
INSERT INTO users (email, password_hash, full_name) VALUES (?, ?, ?);

-- name: ListStaffUsers :many
SELECT u.id, u.email, COALESCE(u.full_name,'') AS full_name, u.created_at,
       p.apps, p.role, p.is_active
FROM user_app_permissions p
JOIN users u ON u.id = p.user_id
WHERE p.owner_id = ?
ORDER BY u.created_at DESC
LIMIT ? OFFSET ?;

-- name: ListStaffUsersByBusinessID :many
SELECT u.id, u.email, COALESCE(u.full_name,'') AS full_name, u.created_at,
       p.apps, p.role, p.is_active
FROM user_app_permissions p
JOIN users u ON u.id = p.user_id
WHERE p.business_id = ?
ORDER BY u.created_at DESC
LIMIT ? OFFSET ?;

-- name: GetStaffUserByID :one
SELECT u.id, u.email, COALESCE(u.full_name,'') AS full_name, u.created_at,
       p.apps, p.role, p.is_active
FROM user_app_permissions p
JOIN users u ON u.id = p.user_id
WHERE p.owner_id = ?
  AND p.business_id = ?
  AND p.user_id = ?
LIMIT 1;

-- name: CreateInviteToken :exec
INSERT INTO invite_tokens (user_id, owner_id, token, expires_at) VALUES (?, ?, ?, ?);

-- name: GetInviteToken :one
SELECT user_id, owner_id, expires_at, used_at FROM invite_tokens WHERE token = ?;

-- name: MarkInviteTokenUsed :exec
UPDATE invite_tokens SET used_at = NOW() WHERE token = ?;

-- name: ConsumeInviteToken :execresult
UPDATE invite_tokens
SET used_at = NOW()
WHERE token = ? AND used_at IS NULL AND expires_at > NOW();

-- name: SetPassword :exec
UPDATE users SET password_hash = ? WHERE id = ?;

-- name: GetBusinessByID :one
SELECT name, owner_id FROM businesses WHERE id = ?;

-- name: CreateAccessRequest :exec
INSERT INTO access_requests (business_id, user_id, app)
VALUES (?, ?, ?)
ON DUPLICATE KEY UPDATE status = 'pending', created_at = NOW();

-- name: IsOnboarded :one
SELECT COALESCE(onboarded, FALSE) FROM users WHERE id = ?;

-- name: MarkOnboarded :exec
UPDATE users SET onboarded = TRUE WHERE id = ?;

-- name: UpdateUserProfile :exec
UPDATE users SET full_name = ?, updated_at = NOW() WHERE id = ?;

-- name: GetBusinessIDByOwnerID :one
SELECT id FROM businesses WHERE owner_id = ? LIMIT 1;

-- name: CountActiveMembers :one
SELECT COUNT(DISTINCT user_id)
FROM (
    SELECT b.owner_id AS user_id
    FROM businesses b
    JOIN users u ON u.id = b.owner_id
    WHERE b.owner_id = ? AND COALESCE(b.is_active, TRUE) = TRUE AND COALESCE(u.is_banned, FALSE) = FALSE
    UNION
    SELECT p.user_id AS user_id
    FROM user_app_permissions p
    JOIN users u ON u.id = p.user_id
    WHERE p.owner_id = ? AND COALESCE(p.is_active, TRUE) = TRUE AND COALESCE(u.is_banned, FALSE) = FALSE
    UNION
    SELECT u.id AS user_id
    FROM users u
    JOIN businesses b ON b.id = u.business_id
    WHERE b.owner_id = ? AND COALESCE(b.is_active, TRUE) = TRUE AND COALESCE(u.is_banned, FALSE) = FALSE
) active_users;

-- cross-schema read from subscriptions + plans — approved for owner app resolution
-- name: GetOwnerApps :one
SELECT p.apps
FROM subscriptions s
JOIN plans p ON p.id = s.plan_id
WHERE s.owner_id = ?;

-- cross-schema read from subscriptions + plans — approved for owner subscription info
-- name: GetOwnerSubscriptionInfo :one
SELECT p.slug, s.status
FROM subscriptions s
JOIN plans p ON p.id = s.plan_id
WHERE s.owner_id = ?;

-- cross-schema read from subscriptions — approved for seat period management
-- name: LockSubscriptionForSeatChange :one
SELECT s.id, p.price_per_seat_cents, p.max_seats
FROM subscriptions s
JOIN plans p ON p.id = s.plan_id
WHERE s.owner_id = ?
FOR UPDATE;

-- cross-schema write to seat_periods — approved for seat billing accuracy
-- name: CloseActiveSeatPeriod :exec
UPDATE seat_periods SET period_end = ?
WHERE subscription_id = ? AND period_end IS NULL;

-- cross-schema write to seat_periods — approved for seat billing accuracy
-- name: OpenSeatPeriod :exec
INSERT INTO seat_periods (subscription_id, owner_id, seat_count, price_per_seat_cents, period_start, billing_month)
VALUES (?, ?, ?, ?, ?, ?);

-- cross-schema write to subscriptions — approved for seat billing accuracy
-- name: UpdateSubscriptionSeatCount :exec
UPDATE subscriptions SET seat_count = ?, updated_at = CURRENT_TIMESTAMP WHERE owner_id = ?;

-- cross-schema read from hub — approved for business-wide admin notification fan-out
-- name: ListAdminsAndOwnerUserIDs :many
SELECT b.owner_id AS user_id
FROM businesses b
WHERE b.id = ?
UNION
SELECT p.user_id
FROM user_app_permissions p
WHERE p.business_id = ?
  AND p.role IN ('admin', 'owner')
  AND p.is_active = TRUE;

-- name: ResolveAccessRequest :exec
UPDATE access_requests SET status = ? WHERE id = ? AND business_id = ?;

-- name: ListAccessRequestsByBusiness :many
SELECT ar.id, ar.user_id, ar.app, ar.status, ar.created_at,
       COALESCE(u.full_name,'') AS requester_name, COALESCE(u.email,'') AS requester_email
FROM access_requests ar
JOIN users u ON u.id = ar.user_id
WHERE ar.business_id = ?
  AND ar.status = 'pending'
ORDER BY ar.created_at DESC;

-- cross-schema write to hub — deliver team access notifications through the hub module.
-- name: CreateHubNotification :exec
INSERT INTO hub_notifications (owner_id, user_id, title, body, type, link)
VALUES (?, ?, ?, ?, ?, ?);

-- cross-schema write to hr_core — every invited user is also an hr_employees row
-- name: EnsureEmployeeForUser :exec
INSERT INTO hr_employees (owner_id, business_id, user_id, first_name, last_name, email, status)
SELECT ?, COALESCE(u.business_id, 0), u.id,
       COALESCE(NULLIF(SUBSTRING_INDEX(u.full_name, ' ', 1), ''), u.email),
       CASE WHEN LOCATE(' ', u.full_name) > 0 THEN SUBSTRING(u.full_name, LOCATE(' ', u.full_name) + 1) ELSE '' END,
       u.email, 'active'
FROM users u
WHERE u.id = ?
  AND NOT EXISTS (SELECT 1 FROM hr_employees e WHERE e.user_id = u.id);

-- cross-schema read from hr_core — revoke access when the employee is terminated.
-- name: RevokeBusinessUserPermissions :execrows
-- Employment-driven revocation (#834): one business, one user. The join to
-- hr_employees keeps the decision inside this module's SQL while the
-- employment fact stays authoritative in hr_core's table. The join column is
-- pinned with sqlc.arg so both business-id placeholders type as int64.
UPDATE user_app_permissions p
JOIN hr_employees e
  ON e.user_id = p.user_id
 AND e.business_id = sqlc.arg('business_id')
SET p.is_active = FALSE, p.updated_at = NOW()
WHERE p.business_id = sqlc.arg('business_id')
  AND p.user_id = sqlc.arg('user_id')
  AND p.is_active = TRUE
  AND e.status = 'terminated';

-- cross-schema read from hr_core — sweep permissions for terminated employees.
-- name: ReconcileTerminatedEmployeeAccess :execrows
-- Daily sweep: every still-active permission row whose user is a terminated
-- employee of that same business is deactivated. Catches anything the
-- synchronous hook missed (crash between status write and revoke, direct SQL,
-- or the hook being unwired).
UPDATE user_app_permissions p
JOIN hr_employees e
  ON e.user_id = p.user_id
 AND e.business_id = p.business_id
SET p.is_active = FALSE, p.updated_at = NOW()
WHERE e.status = 'terminated'
  AND p.is_active = TRUE;
