-- name: GetProduct :one
SELECT id, owner_id, category_id, supplier_id, name, sku, barcode, description,
       unit, cost_price, sale_price, stock_quantity, low_stock_threshold,
       reorder_quantity, is_active, track_inventory, created_at, updated_at
FROM erp_products
WHERE owner_id = ? AND id = ?;

-- name: CreateProduct :execresult
INSERT INTO erp_products
(owner_id, category_id, supplier_id, name, sku, barcode, description,
 unit, cost_price, sale_price, low_stock_threshold, reorder_quantity,
 is_active, track_inventory)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: UpdateProduct :exec
UPDATE erp_products
SET category_id = ?, supplier_id = ?, name = ?, sku = ?, barcode = ?,
    description = ?, unit = ?, cost_price = ?, sale_price = ?,
    low_stock_threshold = ?, reorder_quantity = ?, is_active = ?, track_inventory = ?
WHERE owner_id = ? AND id = ?;

-- name: DeleteProduct :exec
DELETE FROM erp_products WHERE owner_id = ? AND id = ?;
