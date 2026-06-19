-- name: ListVariantsByProduct :many
SELECT id, owner_id, business_id, product_id, name, sku, barcode,
       cost_price, sale_price, stock_quantity, COALESCE(attributes, '[]') AS attributes, is_active,
       created_at, updated_at
FROM erp_product_variants
WHERE business_id = ? AND product_id = ?
ORDER BY name ASC;

-- name: ListVariantsByOwner :many
SELECT id, owner_id, business_id, product_id, name, sku, barcode,
       cost_price, sale_price, stock_quantity, COALESCE(attributes, '[]') AS attributes, is_active,
       created_at, updated_at
FROM erp_product_variants
WHERE business_id = ?
ORDER BY name ASC;

-- name: CreateVariant :execresult
INSERT INTO erp_product_variants
(owner_id, business_id, product_id, name, sku, barcode, cost_price, sale_price, stock_quantity, attributes, is_active)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: UpdateVariant :exec
UPDATE erp_product_variants
SET product_id = ?, name = ?, sku = ?, barcode = ?,
    cost_price = ?, sale_price = ?, stock_quantity = ?, attributes = ?, is_active = ?
WHERE business_id = ? AND id = ?;

-- name: DeleteVariant :exec
DELETE FROM erp_product_variants WHERE business_id = ? AND id = ?;

-- name: CountVariantsByOwner :one
SELECT COUNT(*) FROM erp_product_variants WHERE business_id = ?;

-- name: CountVariantsByProduct :one
SELECT COUNT(*) FROM erp_product_variants WHERE business_id = ? AND product_id = ?;
