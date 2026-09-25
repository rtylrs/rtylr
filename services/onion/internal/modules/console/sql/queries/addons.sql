-- name: ListAddons :many
-- cross-schema read from addons (addons module) — list the add-ons available to console administrators.
SELECT id, name, description, icon, price, is_free, sort_order
FROM addons
ORDER BY sort_order ASC, id ASC
LIMIT 100;

-- name: AddonExists :one
-- cross-schema read from addons (addons module) — validate that the requested add-on exists.
SELECT COUNT(*) FROM addons WHERE id = ?;

-- name: ListAddonGrants :many
-- cross-schema read from addons (addons module) — include add-on names and prices in grant administration results.
SELECT g.id, g.user_id, COALESCE(u.full_name, '') AS user_name, u.email AS user_email,
       g.business_id, COALESCE(b.name, '') AS business_name,
       g.addon_id, a.name AS addon_name, a.price AS addon_price,
       g.granted_by, adm.full_name AS granter_name, g.note, g.expires_at, g.revoked_at,
       g.enabled_by_grant, g.created_at
FROM manual_addon_grants g
JOIN addons a ON a.id = g.addon_id
LEFT JOIN users u ON u.id = g.user_id
LEFT JOIN businesses b ON b.id = g.business_id
LEFT JOIN users adm ON adm.id = g.granted_by
ORDER BY g.created_at DESC
LIMIT ? OFFSET ?;

-- name: CreateAddonGrant :execresult
INSERT INTO manual_addon_grants (user_id, business_id, addon_id, granted_by, note, expires_at, enabled_by_grant)
VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: GetAddonGrantScope :one
SELECT user_id, business_id, addon_id, enabled_by_grant
FROM manual_addon_grants
WHERE id = ?
ORDER BY id ASC
LIMIT 1;

-- name: RevokeAddonGrant :exec
UPDATE manual_addon_grants SET revoked_at = NOW() WHERE id = ? AND business_id = ?;

-- name: DeleteAddonGrant :exec
DELETE FROM manual_addon_grants WHERE id = ? AND business_id = ?;

-- name: IsBusinessAddonEnabled :one
-- cross-schema read from business_addons (addons module) — check whether an add-on is enabled for a business.
SELECT COUNT(*) FROM business_addons
WHERE business_id = ? AND addon_id = ? AND is_enabled = TRUE;

-- name: EnableBusinessAddon :exec
-- cross-schema write to business_addons (addons module) — enable an add-on for a business.
INSERT INTO business_addons (business_id, addon_id, is_enabled, enabled_at)
VALUES (?, ?, TRUE, NOW())
ON DUPLICATE KEY UPDATE is_enabled = TRUE, enabled_at = NOW(), disabled_at = NULL;

-- name: DisableBusinessAddon :exec
-- cross-schema write to business_addons (addons module) — disable an add-on for a business.
UPDATE business_addons
SET is_enabled = FALSE, disabled_at = NOW()
WHERE business_id = ? AND addon_id = ?;

-- name: HasActiveAddonGrant :one
SELECT COUNT(*) FROM manual_addon_grants
WHERE business_id = ? AND addon_id = ? AND revoked_at IS NULL
  AND (expires_at IS NULL OR expires_at > CURRENT_TIMESTAMP);
