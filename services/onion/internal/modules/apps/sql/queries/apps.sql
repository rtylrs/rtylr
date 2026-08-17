-- cross-schema read from business_addons — addon enablement owned by addons module
-- name: ListEnabledAddonIDs :many
SELECT addon_id
FROM business_addons
WHERE business_id = ? AND is_enabled = 1
LIMIT 100;

-- cross-schema read from user_app_permissions — app permissions owned by hub module
-- name: GetUserAllowedAppsJSON :one
SELECT apps
FROM user_app_permissions
WHERE user_id = ?
  AND (
    business_id = ?
    OR (
      business_id IS NULL
      AND owner_id = (SELECT b.owner_id FROM businesses b WHERE b.id = ? ORDER BY b.owner_id ASC LIMIT 1)
    )
  )
LIMIT 1;
