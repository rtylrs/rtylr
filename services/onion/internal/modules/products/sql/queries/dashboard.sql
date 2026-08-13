-- name: SumInventoryValue :one
SELECT COALESCE(SUM(stock_quantity * cost_price), 0) AS inventory_value
FROM erp_products
WHERE business_id = ? AND is_active = TRUE AND track_inventory = TRUE;

-- name: ListTopLowStockItems :many
SELECT p.id, p.name, COALESCE(p.sku, '') AS sku,
       p.stock_quantity, COALESCE(p.low_stock_threshold, 0) AS low_stock_threshold
FROM erp_products p
WHERE p.business_id = ? AND p.is_active = TRUE AND p.track_inventory = TRUE
  AND p.low_stock_threshold IS NOT NULL
  AND p.stock_quantity <= p.low_stock_threshold
ORDER BY (p.stock_quantity - p.low_stock_threshold) ASC, p.stock_quantity ASC
LIMIT ?;

-- cross-schema read from erp_suppliers — join for display name
-- name: ListRecentPurchaseOrders :many
SELECT po.id, po.po_number, COALESCE(s.name, 'Unknown') AS supplier_name,
       po.status, po.total, po.created_at
FROM erp_purchase_orders po
LEFT JOIN erp_suppliers s ON s.id = po.supplier_id
WHERE po.business_id = ?
ORDER BY po.created_at DESC
LIMIT ?;

-- cross-schema read from erp_categories — join for display name
-- name: ListTopCategories :many
SELECT COALESCE(c.name, 'Uncategorized') AS name, COUNT(p.id) AS product_count
FROM erp_products p
LEFT JOIN erp_categories c ON c.id = p.category_id
WHERE p.business_id = ? AND p.is_active = TRUE
GROUP BY c.id, c.name
ORDER BY product_count DESC
LIMIT ?;

-- name: ListReorderSuggestions :many
SELECT id, name, COALESCE(sku, '') AS sku,
       stock_quantity, COALESCE(low_stock_threshold, 0) AS low_stock_threshold,
       COALESCE(reorder_quantity, 0) AS reorder_quantity
FROM erp_products
WHERE business_id = ? AND is_active = TRUE AND track_inventory = TRUE
  AND low_stock_threshold IS NOT NULL
  AND stock_quantity <= low_stock_threshold
ORDER BY (stock_quantity - low_stock_threshold) ASC, id ASC
LIMIT 101;

-- name: CountProductsByCategory :one
SELECT COUNT(*) FROM erp_products WHERE business_id = ? AND category_id = ?;

-- name: CountAllProducts :one
SELECT COUNT(*) FROM erp_products WHERE business_id = ?;

-- name: DeletePrimaryProductImage :exec
DELETE FROM erp_product_images WHERE business_id = ? AND product_id = ? AND is_primary = TRUE;

-- name: InsertPrimaryProductImage :exec
INSERT INTO erp_product_images (owner_id, business_id, product_id, url, alt, position, is_primary)
VALUES (?, ?, ?, ?, ?, 0, TRUE);

-- name: LockProductStockForUpdate :one
SELECT stock_quantity FROM erp_products WHERE business_id = ? AND id = ? FOR UPDATE;
