-- name: GetProduct :one
SELECT id, owner_id, business_id, category_id, supplier_id, brand_id, name, sku, barcode, description,
       unit, cost_price, sale_price, compare_at_price, tax_rate, currency,
       stock_quantity, low_stock_threshold, reorder_quantity, allow_backorder, lead_time_days,
       weight, weight_unit, length, width, height, dimension_unit,
       country_of_origin, hs_code, is_active, track_inventory, created_at, updated_at
FROM erp_products
WHERE business_id = ? AND id = ?;

-- name: CreateProduct :execresult
INSERT INTO erp_products
(owner_id, business_id, category_id, supplier_id, brand_id, name, sku, barcode, description,
 unit, cost_price, sale_price, compare_at_price, tax_rate, currency,
 low_stock_threshold, reorder_quantity, allow_backorder, lead_time_days,
 weight, weight_unit, length, width, height, dimension_unit,
 country_of_origin, hs_code, is_active, track_inventory)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetCategoryReference :one
-- cross-schema read from erp_categories (categories module) — validate the product’s category reference for the tenant.
SELECT id FROM erp_categories WHERE business_id = ? AND id = ?;

-- name: GetSupplierReference :one
-- cross-schema read from erp_suppliers (suppliers module) — validate the product’s supplier reference for the tenant.
SELECT id FROM erp_suppliers WHERE business_id = ? AND id = ?;

-- name: GetBrandReference :one
-- cross-schema read from erp_brands (brands module) — validate the product’s brand reference for the tenant.
SELECT id FROM erp_brands WHERE business_id = ? AND id = ?;

-- name: UpdateProduct :exec
UPDATE erp_products
SET category_id = ?, supplier_id = ?, brand_id = ?, name = ?, sku = ?, barcode = ?,
    description = ?, unit = ?, cost_price = ?, sale_price = ?, compare_at_price = ?,
    tax_rate = ?, currency = ?, low_stock_threshold = ?,
    reorder_quantity = ?, allow_backorder = ?, lead_time_days = ?, weight = ?,
    weight_unit = ?, length = ?, width = ?, height = ?, dimension_unit = ?,
    country_of_origin = ?, hs_code = ?, is_active = ?, track_inventory = ?
WHERE business_id = ? AND id = ?;

-- name: DeleteProduct :exec
DELETE FROM erp_products WHERE business_id = ? AND id = ?;

-- name: GetProductByBarcode :one
SELECT id, owner_id, business_id, category_id, supplier_id, brand_id, name, sku, barcode, description,
       unit, cost_price, sale_price, compare_at_price, tax_rate, currency,
       stock_quantity, low_stock_threshold, reorder_quantity, allow_backorder, lead_time_days,
       weight, weight_unit, length, width, height, dimension_unit,
       country_of_origin, hs_code, is_active, track_inventory, created_at, updated_at
FROM erp_products
WHERE business_id = ? AND barcode = ?
LIMIT 1;

-- name: CountActiveProducts :one
SELECT COUNT(*) FROM erp_products WHERE business_id = ? AND is_active = TRUE;

-- name: CountLowStockProducts :one
SELECT COUNT(*) FROM erp_products
WHERE business_id = ? AND is_active = TRUE
  AND track_inventory = TRUE
  AND low_stock_threshold IS NOT NULL
  AND stock_quantity <= low_stock_threshold;

-- name: CountPendingPurchaseOrders :one
-- cross-schema read from erp_purchase_orders (purchase-orders module) — count open purchase orders for the product dashboard.
SELECT COUNT(*) FROM erp_purchase_orders
WHERE business_id = ? AND status IN ('draft','sent','partial');

-- name: CountActiveSuppliers :one
-- cross-schema read from erp_suppliers (suppliers module) — count active suppliers for the product dashboard.
SELECT COUNT(*) FROM erp_suppliers WHERE business_id = ? AND is_active = TRUE;

-- name: ListProductAttributes :many
SELECT id, name, value, position FROM erp_product_attributes
WHERE business_id = ? AND product_id = ?
ORDER BY position ASC, id ASC;

-- name: DeleteProductAttributes :exec
DELETE FROM erp_product_attributes WHERE business_id = ? AND product_id = ?;

-- name: InsertProductAttribute :exec
INSERT INTO erp_product_attributes (owner_id, business_id, product_id, name, value, position)
VALUES (?, ?, ?, ?, ?, ?);

-- name: GetSalesProduct :one
SELECT name, COALESCE(sku, '') AS sku, sale_price, COALESCE(tax_rate, 0) AS tax_rate
FROM erp_products WHERE business_id = ? AND id = ? AND is_active = TRUE;

-- name: GetProductForUpdate :one
SELECT id, owner_id, business_id, category_id, supplier_id, brand_id, name, sku, barcode, description,
       unit, cost_price, sale_price, compare_at_price, tax_rate, currency,
       stock_quantity, low_stock_threshold, reorder_quantity, allow_backorder, lead_time_days,
       weight, weight_unit, length, width, height, dimension_unit,
       country_of_origin, hs_code, is_active, track_inventory, created_at, updated_at
FROM erp_products
WHERE business_id = ? AND id = ? FOR UPDATE;
