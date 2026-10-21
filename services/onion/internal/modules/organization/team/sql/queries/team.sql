-- name: GetUserByEmail :one
SELECT id, email, COALESCE(full_name,'') AS full_name FROM users WHERE email = ? LIMIT 1;

-- name: GetUserByID :one
SELECT email, COALESCE(full_name,'') AS full_name, is_console_admin FROM users WHERE id = ?;

-- name: GetUserBusinessID :one
SELECT business_id FROM users WHERE id = ?;

-- name: CreateUser :execresult
INSERT INTO users (email, username, password_hash, full_name) VALUES (?, ?, ?, ?);

-- Team access is sourced only from active business membership.
-- name: ListStaffUsersByBusinessID :many
SELECT u.id, u.email, COALESCE(u.full_name,'') AS full_name, u.created_at,
       member.member_type, member.status AS member_status
FROM authorization_business_members AS member
JOIN users AS u ON u.id = member.user_id
WHERE member.business_id = ?
ORDER BY u.created_at DESC, u.id DESC
LIMIT ? OFFSET ?;

-- Business owners are visible to Flow even in older tenants where an owner
-- membership row was never backfilled.
-- name: ListActiveMemberUserIDsByBusinessID :many
SELECT business.owner_id AS user_id
FROM businesses AS business
WHERE business.id = ?
UNION
SELECT user_id
FROM authorization_business_members
WHERE business_id = ? AND status = 'active';

-- name: CountStaffUsersByBusinessID :one
SELECT COUNT(*)
FROM authorization_business_members
WHERE business_id = ?;

-- name: GetStaffUserByID :one
SELECT u.id, u.email, COALESCE(u.full_name,'') AS full_name, u.created_at,
       member.member_type, member.status AS member_status
FROM authorization_business_members AS member
JOIN businesses AS business ON business.id = member.business_id
JOIN users AS u ON u.id = member.user_id
WHERE business.owner_id = ? AND member.business_id = ? AND member.user_id = ?
LIMIT 1;

-- A consumer that owns a durable notification can resolve only the creator's
-- active contact details, without reading the organization tables itself.
-- name: GetActiveMemberContact :one
SELECT u.email, COALESCE(u.full_name, '') AS full_name
FROM authorization_business_members AS member
JOIN users AS u ON u.id = member.user_id
WHERE member.business_id = ? AND member.user_id = ? AND member.status = 'active'
LIMIT 1;

-- name: CreateInviteToken :exec
INSERT INTO invite_tokens (user_id, owner_id, business_id, token, expires_at) VALUES (?, ?, ?, ?, ?);

-- name: GetInviteToken :one
SELECT user_id, owner_id, business_id, expires_at, used_at FROM invite_tokens WHERE token = ?;

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
SELECT COUNT(*)
FROM (
    SELECT b.owner_id AS user_id
    FROM businesses b
    JOIN users u ON u.id = b.owner_id
    WHERE b.owner_id = ? AND COALESCE(b.is_active, TRUE) = TRUE AND COALESCE(u.is_banned, FALSE) = FALSE
    UNION
    SELECT member.user_id
    FROM authorization_business_members member
    JOIN businesses b ON b.id = member.business_id
    JOIN users u ON u.id = member.user_id
    WHERE b.owner_id = ? AND COALESCE(b.is_active, TRUE) = TRUE
      AND member.status = 'active' AND COALESCE(u.is_banned, FALSE) = FALSE
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

-- cross-schema read from hub — business administrator notification fan-out
-- name: ListAdminsAndOwnerUserIDs :many
SELECT b.owner_id AS user_id
FROM businesses b
WHERE b.id = ?
UNION
SELECT member.user_id
FROM authorization_business_members member
WHERE member.business_id = ? AND member.member_type = 'admin' AND member.status = 'active';

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

-- Only fills an unset business. users.business_id is the user's current business
-- and a person can belong to more than one, so inviting somebody who already has
-- one must not move them out of it.
-- name: SetUserBusinessIDIfUnset :execrows
UPDATE users
SET business_id = ?
WHERE id = ?
  AND (business_id IS NULL OR business_id = 0);
