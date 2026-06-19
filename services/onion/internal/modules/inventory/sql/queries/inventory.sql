-- name: InsertInventoryTransaction :execresult
INSERT INTO erp_inventory_transactions
(owner_id, product_id, type, quantity_change, quantity_before, quantity_after, notes)
VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: GetInventoryReport :many
SELECT id, name, sku, unit, stock_quantity, low_stock_threshold,
       reorder_quantity, cost_price, sale_price,
       CAST(stock_quantity * cost_price AS DECIMAL(18,4)) AS stock_value,
       CASE WHEN low_stock_threshold IS NOT NULL AND stock_quantity <= low_stock_threshold THEN 1 ELSE 0 END AS is_low_stock
FROM erp_products
WHERE owner_id = ? AND track_inventory = TRUE AND is_active = TRUE
ORDER BY name ASC;

-- name: GetLowStockProducts :many
SELECT id, name, sku, unit, stock_quantity, low_stock_threshold,
       reorder_quantity, cost_price, sale_price,
       CAST(stock_quantity * cost_price AS DECIMAL(18,4)) AS stock_value,
       1 AS is_low_stock
FROM erp_products
WHERE owner_id = ? AND track_inventory = TRUE AND is_active = TRUE
  AND low_stock_threshold IS NOT NULL
  AND stock_quantity <= low_stock_threshold
ORDER BY stock_quantity ASC;

-- name: GetValuation :many
SELECT
    COALESCE(c.name, 'Uncategorized') AS category_name,
    '' AS warehouse_name,
    COUNT(*) AS total_items,
    SUM(p.stock_quantity * p.sale_price) AS total_value,
    SUM(p.stock_quantity * p.cost_price) AS total_cost
FROM erp_products p
LEFT JOIN erp_categories c ON c.id = p.category_id AND c.owner_id = p.owner_id
WHERE p.owner_id = ? AND p.is_active = TRUE AND p.track_inventory = TRUE
GROUP BY c.name
ORDER BY total_value DESC;
