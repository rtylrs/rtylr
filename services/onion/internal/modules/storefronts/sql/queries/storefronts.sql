-- name: ListStorefronts :many
SELECT id, owner_id, name, domain, logo_url, theme_id, is_published, created_at, updated_at
FROM store_sites
WHERE owner_id = ?
ORDER BY name ASC;

-- name: CreateStorefront :execresult
INSERT INTO store_sites (owner_id, name, domain, logo_url, theme_id, is_published)
VALUES (?, ?, ?, ?, ?, ?);

-- name: UpdateStorefront :exec
UPDATE store_sites
SET name = ?, domain = ?, logo_url = ?, theme_id = ?, is_published = ?
WHERE owner_id = ? AND id = ?;
