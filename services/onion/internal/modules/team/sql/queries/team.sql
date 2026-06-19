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
ORDER BY created_at DESC;

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

-- name: CountActiveMembers :one
SELECT COUNT(*) FROM user_app_permissions WHERE owner_id = ? AND is_active = 1;

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

-- name: CreateHubNotification :exec
INSERT INTO hub_notifications (owner_id, user_id, title, body, type, link)
VALUES (?, ?, ?, ?, ?, ?);
