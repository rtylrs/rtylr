-- name: ListShippingMethods :many
SELECT id, owner_id, business_id, name, price, estimated_days, is_active, created_at
FROM store_shipping_methods WHERE business_id = ? ORDER BY name ASC
LIMIT 200;

-- name: CreateShippingMethod :execresult
INSERT INTO store_shipping_methods (owner_id, business_id, name, price, estimated_days, is_active) VALUES (?, ?, ?, ?, ?, ?);

-- name: UpdateShippingMethod :exec
UPDATE store_shipping_methods SET name = ?, price = ?, estimated_days = ?, is_active = ? WHERE business_id = ? AND id = ?;

-- name: DeleteShippingMethod :exec
DELETE FROM store_shipping_methods WHERE business_id = ? AND id = ?;
