-- name: ListStorefronts :many
SELECT id, owner_id, business_id, name, domain, logo_url, theme_id, is_published, created_at, updated_at
FROM store_sites
WHERE business_id = ?
ORDER BY name ASC
LIMIT 200;

-- name: CreateStorefront :execresult
INSERT INTO store_sites (owner_id, business_id, name, domain, logo_url, theme_id, is_published)
VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: UpdateStorefront :exec
UPDATE store_sites
SET name = ?, domain = ?, logo_url = ?, theme_id = ?, is_published = ?
WHERE business_id = ? AND id = ?;

-- name: CountThemeForOwner :one
SELECT COUNT(*) FROM store_themes
WHERE id = ? AND business_id = ?;
