-- name: GetActiveBusinessOwnerID :one
SELECT owner_id FROM businesses WHERE id = ? AND is_active = TRUE LIMIT 1;

-- name: GetUserDefaultBusinessID :one
SELECT business_id FROM users WHERE id = ? LIMIT 1;

-- name: GetAuthorizedBusinessID :one
SELECT b.id
FROM businesses b
WHERE b.id = ? AND COALESCE(b.is_active, TRUE) = TRUE
  AND (b.owner_id = sqlc.arg(user_id) OR EXISTS (
    SELECT 1 FROM authorization_business_members abm
    WHERE abm.user_id = sqlc.arg(user_id) AND abm.business_id = b.id AND abm.status = 'active'
  ))
LIMIT 1;

-- name: SubscriptionAccess :one
SELECT s.status,
       EXISTS (SELECT 1 FROM manual_access_grants g
           WHERE g.business_id = s.business_id AND g.plan_id = s.plan_id
             AND g.revoked_at IS NULL
             AND (g.expires_at IS NULL OR g.expires_at > CURRENT_TIMESTAMP)) AS granted,
       s.trial_ends_at
FROM subscriptions s WHERE s.business_id = ?;

-- name: UploadKeyScope :one
SELECT owner_id, business_id FROM hub_api_keys
WHERE key_hash = ? AND is_active = TRUE
  AND (expires_at IS NULL OR expires_at > CURRENT_TIMESTAMP)
  AND JSON_CONTAINS(scopes, JSON_QUOTE('upload'))
LIMIT 1;

-- name: AccountPasswordHash :one
SELECT password_hash FROM users WHERE id = ?;

-- name: AccountEmail :one
SELECT email FROM users WHERE id = ?;

-- name: BusinessUserName :one
SELECT u.full_name FROM users u
WHERE u.id = sqlc.arg(user_id) AND EXISTS (
 SELECT 1 FROM businesses b WHERE b.id = sqlc.arg(business_id)
 AND (b.owner_id = u.id OR EXISTS (SELECT 1 FROM authorization_business_members m WHERE m.business_id = b.id AND m.user_id = u.id))
);

-- name: AuthorizationVersion :one
SELECT COALESCE(m.authorization_version, 1) AS version
FROM businesses b
LEFT JOIN authorization_business_members m ON m.business_id = b.id AND m.user_id = sqlc.arg(user_id) AND m.status = 'active'
WHERE b.id = sqlc.arg(business_id) AND b.is_active = TRUE
 AND (b.owner_id = sqlc.arg(user_id) OR m.user_id IS NOT NULL);

-- name: LocationLimit :one
SELECT p.max_locations FROM businesses b
JOIN subscriptions s ON s.business_id = b.id
JOIN plans p ON p.id = s.plan_id
WHERE b.id = ? AND b.is_active = TRUE;

-- name: BusinessUsers :many
SELECT u.id, COALESCE(u.full_name, '') AS name, u.email
FROM users u
WHERE u.id IN (sqlc.slice('user_ids')) AND EXISTS (
 SELECT 1 FROM businesses b WHERE b.id = sqlc.arg(business_id)
 AND (b.owner_id = u.id OR u.business_id = b.id OR EXISTS (
 SELECT 1 FROM authorization_business_members m WHERE m.business_id = b.id AND m.user_id = u.id))
);

-- name: ActiveBusinessUsers :many
SELECT u.id, COALESCE(u.full_name, '') AS name, u.email
FROM users u
JOIN businesses b ON b.id = sqlc.arg(business_id) AND COALESCE(b.is_active, TRUE) = TRUE
WHERE b.owner_id = u.id OR EXISTS (
  SELECT 1 FROM authorization_business_members m
  WHERE m.business_id = b.id AND m.user_id = u.id AND m.status = 'active')
ORDER BY name, u.id;

-- name: AccountByEmail :one
SELECT id FROM users WHERE email = ? LIMIT 1;

-- name: BusinessUserByUsername :one
SELECT u.id FROM users u WHERE u.username = sqlc.arg(username) AND EXISTS (
 SELECT 1 FROM businesses b WHERE b.id = sqlc.arg(business_id) AND b.is_active = TRUE
 AND (b.owner_id = u.id OR EXISTS (
 SELECT 1 FROM authorization_business_members m WHERE m.business_id = b.id AND m.user_id = u.id AND m.status = 'active'))
) LIMIT 1;

-- name: AssistantIdentity :one
SELECT COALESCE(u.full_name, '') AS name, b.name AS business, COALESCE(b.timezone,'UTC') AS timezone, COALESCE(b.currency,'USD') AS currency
FROM users u JOIN businesses b ON b.id = sqlc.arg(business_id)
WHERE u.id = sqlc.arg(user_id) AND b.is_active = TRUE
 AND (b.owner_id = u.id OR EXISTS (SELECT 1 FROM authorization_business_members m
 WHERE m.business_id=b.id AND m.user_id=u.id AND m.status='active'));
