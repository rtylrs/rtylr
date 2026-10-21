-- name: GetHubNavigationMenu :one
SELECT items
FROM hub_navigation_menus
WHERE business_id = ?
ORDER BY business_id ASC
LIMIT 1;

-- name: UpsertHubNavigationMenu :exec
INSERT INTO hub_navigation_menus (
    owner_id,
    business_id,
    items
) VALUES (
    ?, ?, ?
)
ON DUPLICATE KEY UPDATE
    owner_id = VALUES(owner_id),
    items = VALUES(items),
    updated_at = CURRENT_TIMESTAMP;
