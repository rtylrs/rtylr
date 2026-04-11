-- name: ListStoreCategories :many
SELECT id, owner_id, name, slug, description, image_url, parent_id, sort_order, created_at
FROM store_categories WHERE owner_id = ? ORDER BY sort_order ASC;

-- name: CreateStoreCategory :execresult
INSERT INTO store_categories (owner_id, name, slug, description, image_url, parent_id, sort_order)
VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: UpdateStoreCategory :exec
UPDATE store_categories SET name = ?, slug = ?, description = ?, image_url = ?, parent_id = ?, sort_order = ?
WHERE owner_id = ? AND id = ?;

-- name: DeleteStoreCategory :exec
DELETE FROM store_categories WHERE owner_id = ? AND id = ?;
