-- name: ListGrants :many
-- cross-schema read from plans (subscriptions module) — include plan names in console access-grant results.
SELECT g.id, g.user_id, COALESCE(u.full_name, '') AS user_name, g.plan_id, p.name AS plan_name,
       g.granted_by, adm.full_name AS granter_name, g.note, g.expires_at, g.revoked_at, g.created_at,
       u.email AS user_email, b.id AS business_id, b.name AS business_name
FROM manual_access_grants g
JOIN plans p ON p.id = g.plan_id
LEFT JOIN users u ON u.id = g.user_id
LEFT JOIN businesses b ON b.owner_id = g.user_id AND COALESCE(b.is_active, TRUE) = TRUE
LEFT JOIN users adm ON adm.id = g.granted_by
ORDER BY g.created_at DESC
LIMIT ? OFFSET ?;

-- name: CreateGrant :execresult
INSERT INTO manual_access_grants (user_id, plan_id, granted_by, note, expires_at)
VALUES (?, ?, ?, ?, ?);

-- name: GetGrantUserID :one
SELECT user_id FROM manual_access_grants WHERE id = ? ORDER BY id ASC LIMIT 1;

-- name: RevokeGrant :exec
UPDATE manual_access_grants SET revoked_at = NOW() WHERE id = ? AND user_id = ?;

-- name: DeleteGrant :exec
DELETE FROM manual_access_grants WHERE id = ? AND user_id = ?;
