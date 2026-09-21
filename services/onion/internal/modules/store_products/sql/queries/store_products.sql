-- name: CreateStoreProduct :execresult
INSERT INTO store_products (owner_id, business_id, product_id, title, slug, description, images, price, compare_at_price, is_published, sort_order)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: UpdateStoreProduct :execrows
UPDATE store_products SET product_id = ?, title = ?, slug = ?, description = ?, images = ?, price = ?, compare_at_price = ?, is_published = ?, sort_order = ?
WHERE business_id = ? AND id = ?;

-- name: DeleteStoreProduct :exec
DELETE FROM store_products WHERE business_id = ? AND id = ?;
