-- name: ListGrants :many
-- cross-schema read from plans (subscriptions module) — include plan names in console access-grant results.
SELECT g.id, g.business_id, b.name AS business_name, b.owner_id, COALESCE(u.full_name, '') AS owner_name,
       u.email AS owner_email, g.plan_id, p.name AS plan_name,
       g.granted_by, adm.full_name AS granter_name, g.note, g.expires_at, g.revoked_at, g.created_at
FROM manual_access_grants g
JOIN plans p ON p.id = g.plan_id
LEFT JOIN businesses b ON b.id = g.business_id
LEFT JOIN users u ON u.id = b.owner_id
LEFT JOIN users adm ON adm.id = g.granted_by
ORDER BY g.created_at DESC
LIMIT ? OFFSET ?;

-- name: CreateGrant :execresult
INSERT INTO manual_access_grants (business_id, plan_id, granted_by, note, expires_at)
VALUES (?, ?, ?, ?, ?);

-- name: GetGrantBusinessID :one
SELECT business_id FROM manual_access_grants WHERE id = ? LIMIT 1;

-- name: RevokeGrant :exec
UPDATE manual_access_grants SET revoked_at = NOW() WHERE id = ? AND business_id = ?;

-- name: DeleteGrant :exec
DELETE FROM manual_access_grants WHERE id = ? AND business_id = ?;
