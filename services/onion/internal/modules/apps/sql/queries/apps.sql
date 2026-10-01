-- cross-schema read from business_addons — addon enablement owned by addons module
-- name: ListEnabledAddonIDs :many
SELECT addon_id
FROM business_addons
WHERE business_id = ? AND is_enabled = 1
LIMIT 100;
