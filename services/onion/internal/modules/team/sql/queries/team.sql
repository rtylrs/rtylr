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
INSERT INTO users (email, username, password_hash, full_name) VALUES (?, ?, ?, ?);

-- name: ListStaffUsers :many
SELECT u.id, u.email, COALESCE(u.full_name,'') AS full_name, u.created_at,
       p.apps, p.role, p.is_active
FROM user_app_permissions p
JOIN users u ON u.id = p.user_id
WHERE p.owner_id = ?
ORDER BY u.created_at DESC, u.id DESC
LIMIT ? OFFSET ?;

-- The team list is a list of business *members*, so it is sourced from both the
-- normalized membership table and the legacy permission table. Sourcing it from
-- user_app_permissions alone silently hid every member without a legacy row,
-- including the owner (who never gets one), so the owner could not see
-- themselves in their own team.
--
-- Both sources are selected separately and reconciled in Go: the normalized
-- dash_role_key and status win, with the legacy columns as the fallback for
-- members that have not been backfilled yet.
-- name: ListStaffUsersByBusinessID :many
SELECT u.id, u.email, COALESCE(u.full_name,'') AS full_name, u.created_at,
       COALESCE(member.dash_role_key, '') AS dash_role_key,
       permissions.role AS legacy_role,
       permissions.apps AS legacy_apps,
       member.status AS member_status,
       permissions.is_active AS legacy_is_active
FROM users AS u
LEFT JOIN user_app_permissions AS permissions
       ON permissions.user_id = u.id AND permissions.business_id = ?
LEFT JOIN authorization_business_members AS member
       ON member.user_id = u.id AND member.business_id = ?
WHERE permissions.user_id IS NOT NULL OR member.user_id IS NOT NULL
ORDER BY u.created_at DESC, u.id DESC
LIMIT ? OFFSET ?;

-- name: CountStaffUsersByBusinessID :one
SELECT COUNT(*)
FROM users AS u
LEFT JOIN user_app_permissions AS permissions
       ON permissions.user_id = u.id AND permissions.business_id = ?
LEFT JOIN authorization_business_members AS member
       ON member.user_id = u.id AND member.business_id = ?
WHERE permissions.user_id IS NOT NULL OR member.user_id IS NOT NULL;

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

-- name: ReleaseInviteToken :exec
-- Reverts the single-use claim made by ConsumeInviteToken when a later step of
-- the accept flow fails, so a failed accept leaves the invitation retryable
-- instead of stranding the invitee with a consumed token and no activation.
UPDATE invite_tokens SET used_at = NULL WHERE token = ?;

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
    LEFT JOIN user_app_permissions p ON p.owner_id = b.owner_id AND p.user_id = u.id
    WHERE b.owner_id = ? AND COALESCE(b.is_active, TRUE) = TRUE AND COALESCE(u.is_banned, FALSE) = FALSE
      -- Deactivated members keep users.business_id, so the permission row decides
      -- whether they still occupy a seat. A missing row still counts (legacy data).
      AND COALESCE(p.is_active, TRUE) = TRUE
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
INSERT INTO hub_notifications (owner_id, business_id, user_id, title, body, type, link)
VALUES (?, ?, ?, ?, ?, ?, ?);

-- cross-schema write to hr_core — every invited user is also an hr_employees row
-- cross-schema read from users — attach an invited member to their business.
-- Only fills an unset business. users.business_id is the user's current business
-- and a person can belong to more than one, so inviting somebody who already has
-- one must not move them out of it. An invited member's row is NULL until this
-- runs, which is why their token carried no business_id and every Flow call 403d.
-- name: SetUserBusinessIDIfUnset :execrows
UPDATE users
SET business_id = ?
WHERE id = ?
  AND (business_id IS NULL OR business_id = 0);

-- name: EnsureEmployeeForUser :exec
-- business_id is passed in rather than read from the invited user's own row.
-- users.business_id is still NULL at invite time, so COALESCE(u.business_id, 0)
-- collapsed to 0 and every invited member's employee record was scoped to no
-- business at all. The owner's row looked correct only because
-- CreateBusinessForUser calls SetUserBusinessID before this runs.
INSERT INTO hr_employees (owner_id, business_id, user_id, first_name, last_name, email, status)
SELECT ?, ?, u.id,
       COALESCE(NULLIF(SUBSTRING_INDEX(u.full_name, ' ', 1), ''), u.email),
       CASE WHEN LOCATE(' ', u.full_name) > 0 THEN SUBSTRING(u.full_name, LOCATE(' ', u.full_name) + 1) ELSE '' END,
       u.email, 'active'
FROM users u
WHERE u.id = ?
  AND NOT EXISTS (SELECT 1 FROM hr_employees e WHERE e.user_id = u.id);

-- cross-schema write to hr_core: link an existing Directory row to its employee.
-- name: LinkStaffEmployeeForUser :exec
UPDATE hr_staff s
JOIN hr_employees e
  ON e.owner_id = s.owner_id
 AND e.business_id = s.business_id
 AND e.user_id = s.user_id
SET s.employee_id = e.id
WHERE s.owner_id = ?
  AND s.business_id = ?
  AND s.user_id = ?
  AND (s.employee_id IS NULL OR s.employee_id = 0);

-- cross-schema write to hr: mirror an invited employee in the Directory source.
-- The staff table uses enums for fields that hr_employees stores as strings, so
-- map unknown employment types to the schema default.
-- name: EnsureStaffForUser :exec
INSERT INTO hr_staff
    (owner_id, business_id, user_id, employee_id, department_id, first_name, last_name, email,
     phone, role, employment_type, status, salary, hourly_rate, overtime_rate, start_date,
     end_date, address, emergency_contact_name, emergency_contact_phone, apps, notes, photo_url, bio)
SELECT e.owner_id, e.business_id, e.user_id, e.id, e.department_id, e.first_name, e.last_name, e.email,
       e.phone, ?,
       CASE e.employment_type
           WHEN 'part_time' THEN 'part_time'
           WHEN 'contractor' THEN 'contractor'
           WHEN 'intern' THEN 'intern'
           ELSE 'full_time'
       END,
       'active', e.salary, e.hourly_rate, e.overtime_rate, COALESCE(e.start_date, CURDATE()),
       e.end_date, e.address, e.emergency_contact_name, e.emergency_contact_phone,
       JSON_ARRAY(), e.notes, NULL, NULL
FROM hr_employees e
WHERE e.owner_id = ?
  AND e.business_id = ?
  AND e.user_id = ?
  AND NOT EXISTS (
      SELECT 1
      FROM hr_staff s
      WHERE s.owner_id = e.owner_id
        AND s.business_id = e.business_id
        AND s.user_id = e.user_id
  );

-- cross-schema read from hr_core — revoke access when the employee is terminated.
-- name: RevokeBusinessUserPermissions :execrows
-- Employment-driven revocation (#834): one business, one user. The join to
-- hr_employees keeps the decision inside this module's SQL while the
-- employment fact stays authoritative in hr_core's table.
--
-- Both business-id comparisons live in the WHERE clause. sqlc rewrites
-- sqlc.arg in a WHERE expression into a placeholder but left it untouched
-- inside the JOIN ON, so the statement reached MySQL with the literal text
-- sqlc.arg('business_id') in it and every call failed with
-- "Error 1305: FUNCTION sqlc.arg does not exist". The join is an inner join,
-- so moving the condition out of ON and into WHERE selects exactly the same
-- rows.
--
-- The two comparisons are separately named because the columns have different
-- nullability: user_app_permissions.business_id is nullable and types as
-- sql.NullInt64, hr_employees.business_id is NOT NULL and types as int64.
-- Sharing one name makes sqlc refuse the query outright. Both are filled from
-- the same caller argument.
UPDATE user_app_permissions p
JOIN hr_employees e
  ON e.user_id = p.user_id
SET p.is_active = FALSE, p.updated_at = NOW()
WHERE p.business_id = sqlc.arg('business_id')
  AND e.business_id = sqlc.arg('employee_business_id')
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
