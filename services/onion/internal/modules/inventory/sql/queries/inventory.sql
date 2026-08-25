-- name: InsertInventoryTransaction :execresult
INSERT INTO erp_inventory_transactions
(owner_id, business_id, product_id, type, quantity_change, quantity_before, quantity_after, notes)
VALUES (?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetInventoryReport :many
SELECT id, business_id, name, sku, unit, stock_quantity, low_stock_threshold,
       reorder_quantity, cost_price, sale_price,
       CAST(stock_quantity * cost_price AS DECIMAL(18,4)) AS stock_value,
       CASE WHEN low_stock_threshold IS NOT NULL AND stock_quantity <= low_stock_threshold THEN 1 ELSE 0 END AS is_low_stock
FROM erp_products
WHERE business_id = ? AND track_inventory = TRUE AND is_active = TRUE
ORDER BY name ASC
LIMIT 100;

-- name: GetLowStockProducts :many
SELECT id, business_id, name, sku, unit, stock_quantity, low_stock_threshold,
       reorder_quantity, cost_price, sale_price,
       CAST(stock_quantity * cost_price AS DECIMAL(18,4)) AS stock_value,
       1 AS is_low_stock
FROM erp_products
WHERE business_id = ? AND track_inventory = TRUE AND is_active = TRUE
  AND low_stock_threshold IS NOT NULL
  AND stock_quantity <= low_stock_threshold
ORDER BY stock_quantity ASC
LIMIT 100;

-- name: GetValuation :many
SELECT
    COALESCE(c.name, 'Uncategorized') AS category_name,
    l.name AS warehouse_name,
    COUNT(DISTINCT p.id) AS total_items,
    SUM(COALESCE(cl.total_cost, 0)) AS total_value,
    SUM(COALESCE(cl.total_cost, 0)) AS total_cost
FROM stock_levels sl
JOIN stock_locations l
  ON l.id = sl.location_id AND l.owner_id = sl.owner_id AND l.is_active = TRUE
JOIN erp_products p
  ON p.id = sl.product_id AND p.business_id = sl.business_id
LEFT JOIN erp_categories c ON c.id = p.category_id AND c.business_id = p.business_id
LEFT JOIN (
    SELECT owner_id, business_id, product_id, location_id,
           SUM(quantity_remaining * unit_cost) AS total_cost
    FROM inventory_cost_layers
    WHERE inventory_cost_layers.business_id = ?
      AND quantity_remaining > 0
    GROUP BY owner_id, business_id, product_id, location_id
) cl ON cl.owner_id = sl.owner_id
    AND cl.business_id = sl.business_id
    AND cl.product_id = sl.product_id
    AND cl.location_id = sl.location_id
WHERE sl.business_id = ? AND p.is_active = TRUE AND p.track_inventory = TRUE
  AND sl.on_hand > 0
GROUP BY c.name, l.name
ORDER BY total_value DESC;
