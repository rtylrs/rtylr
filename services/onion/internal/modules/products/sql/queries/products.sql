-- name: GetProduct :one
SELECT id, owner_id, category_id, supplier_id, brand_id, name, sku, barcode, description,
       unit, cost_price, sale_price, compare_at_price, tax_rate, currency,
       stock_quantity, low_stock_threshold, reorder_quantity, allow_backorder, lead_time_days,
       weight, weight_unit, length, width, height, dimension_unit,
       country_of_origin, hs_code, is_active, track_inventory, created_at, updated_at
FROM erp_products
WHERE owner_id = ? AND id = ?;

-- name: CreateProduct :execresult
INSERT INTO erp_products
(owner_id, category_id, supplier_id, brand_id, name, sku, barcode, description,
 unit, cost_price, sale_price, compare_at_price, tax_rate, currency,
 low_stock_threshold, reorder_quantity, allow_backorder, lead_time_days,
 weight, weight_unit, length, width, height, dimension_unit,
 country_of_origin, hs_code, is_active, track_inventory)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: UpdateProduct :exec
UPDATE erp_products
SET category_id = ?, supplier_id = ?, brand_id = ?, name = ?, sku = ?, barcode = ?,
    description = ?, unit = ?, cost_price = ?, sale_price = ?, compare_at_price = ?,
    tax_rate = ?, currency = ?, low_stock_threshold = ?,
    reorder_quantity = ?, allow_backorder = ?, lead_time_days = ?, weight = ?,
    weight_unit = ?, length = ?, width = ?, height = ?, dimension_unit = ?,
    country_of_origin = ?, hs_code = ?, is_active = ?, track_inventory = ?
WHERE owner_id = ? AND id = ?;

-- name: DeleteProduct :exec
DELETE FROM erp_products WHERE owner_id = ? AND id = ?;

-- name: GetProductByBarcode :one
SELECT id, owner_id, category_id, supplier_id, brand_id, name, sku, barcode, description,
       unit, cost_price, sale_price, compare_at_price, tax_rate, currency,
       stock_quantity, low_stock_threshold, reorder_quantity, allow_backorder, lead_time_days,
       weight, weight_unit, length, width, height, dimension_unit,
       country_of_origin, hs_code, is_active, track_inventory, created_at, updated_at
FROM erp_products
WHERE owner_id = ? AND barcode = ?
LIMIT 1;

-- name: CountActiveProducts :one
SELECT COUNT(*) FROM erp_products WHERE owner_id = ? AND is_active = TRUE;

-- name: CountLowStockProducts :one
SELECT COUNT(*) FROM erp_products
WHERE owner_id = ? AND is_active = TRUE
  AND track_inventory = TRUE
  AND low_stock_threshold IS NOT NULL
  AND stock_quantity <= low_stock_threshold;

-- name: CountPendingPurchaseOrders :one
SELECT COUNT(*) FROM erp_purchase_orders
WHERE owner_id = ? AND status IN ('draft','sent','partial');

-- name: CountActiveSuppliers :one
SELECT COUNT(*) FROM erp_suppliers WHERE owner_id = ? AND is_active = TRUE;

-- name: ListProductAttributes :many
SELECT id, name, value, position FROM erp_product_attributes
WHERE owner_id = ? AND product_id = ?
ORDER BY position ASC, id ASC;

-- name: DeleteProductAttributes :exec
DELETE FROM erp_product_attributes WHERE owner_id = ? AND product_id = ?;

-- name: InsertProductAttribute :exec
INSERT INTO erp_product_attributes (owner_id, product_id, name, value, position)
VALUES (?, ?, ?, ?, ?);
