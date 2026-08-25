-- name: ListAllAddons :many
SELECT id, name, description, icon, price, is_free, is_default_enabled, sort_order
FROM addons
ORDER BY sort_order ASC
LIMIT 100;

-- name: ListBusinessAddons :many
SELECT ba.addon_id, ba.is_enabled, ba.enabled_at, ba.disabled_at,
       a.name, a.description, a.icon, a.price, a.is_free
FROM business_addons ba
JOIN addons a ON a.id = ba.addon_id
WHERE ba.business_id = ?
LIMIT 100;

-- name: HasActiveAddonGrant :one
SELECT COUNT(*) FROM manual_addon_grants
WHERE business_id = ? AND addon_id = ? AND revoked_at IS NULL
  AND (expires_at IS NULL OR expires_at > CURRENT_TIMESTAMP);

-- name: EnableAddon :exec
INSERT INTO business_addons (business_id, addon_id, is_enabled, enabled_at)
VALUES (?, ?, true, NOW())
ON DUPLICATE KEY UPDATE is_enabled = true, enabled_at = NOW(), disabled_at = NULL;

-- name: DisableAddon :exec
UPDATE business_addons SET is_enabled = false, disabled_at = NOW()
WHERE business_id = ? AND addon_id = ?;
