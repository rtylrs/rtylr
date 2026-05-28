-- name: ListStoreProducts :many
SELECT id, owner_id, product_id, title, slug, description, COALESCE(images, '[]') AS images, price, compare_at_price, is_published, sort_order, created_at, updated_at
FROM store_products WHERE owner_id = ? ORDER BY sort_order ASC;

-- name: CreateStoreProduct :execresult
INSERT INTO store_products (owner_id, product_id, title, slug, description, images, price, compare_at_price, is_published, sort_order)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: UpdateStoreProduct :exec
UPDATE store_products SET product_id = ?, title = ?, slug = ?, description = ?, images = ?, price = ?, compare_at_price = ?, is_published = ?, sort_order = ?
WHERE owner_id = ? AND id = ?;

-- name: DeleteStoreProduct :exec
DELETE FROM store_products WHERE owner_id = ? AND id = ?;
