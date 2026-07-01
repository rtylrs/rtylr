-- name: CountProductSKUsTotal :one
SELECT COUNT(*) AS count FROM erp_products WHERE business_id = ?;

-- name: CountProductSKUsActive :one
SELECT COUNT(*) AS count FROM erp_products WHERE business_id = ? AND is_active = TRUE;

-- name: CountProductSKUsOutOfStock :one
SELECT COUNT(*) AS count
 FROM erp_products
 WHERE business_id = ?
   AND track_inventory = TRUE
   AND stock_quantity <= 0;

-- name: AvgProductMarginPct :one
SELECT COALESCE(AVG((sale_price - cost_price) / sale_price * 100), 0) AS margin_pct
 FROM erp_products
 WHERE business_id = ?
   AND is_active = TRUE
   AND sale_price > 0;

-- name: GetBestSellers :many
SELECT p.id,
       p.name,
       COALESCE(SUM(oi.quantity), 0) AS units,
       COALESCE(SUM(oi.total), 0)    AS revenue
FROM erp_products p
JOIN pos_order_items oi ON oi.menu_item_id = p.id
JOIN pos_orders o       ON o.id = oi.order_id AND o.business_id = p.business_id
WHERE p.business_id = ?
  AND o.created_at >= ?
GROUP BY p.id, p.name
ORDER BY units DESC
LIMIT ?;

-- name: GetSalesByCategory :many
SELECT COALESCE(c.id, 0) AS category_id,
       COALESCE(c.name, 'Uncategorised') AS category_name,
       COALESCE(SUM(oi.quantity), 0) AS units,
       COALESCE(SUM(oi.total), 0)    AS revenue
FROM pos_order_items oi
JOIN pos_orders o        ON o.id = oi.order_id
JOIN erp_products p      ON p.id = oi.menu_item_id AND p.business_id = o.business_id
LEFT JOIN erp_categories c ON c.id = p.category_id AND c.business_id = p.business_id
WHERE o.business_id = ?
  AND o.created_at >= ?
  AND o.created_at <= ?
GROUP BY c.id, c.name
ORDER BY revenue DESC;

-- name: GetProductMargins :many
SELECT (sale_price - cost_price) / sale_price * 100 AS margin_pct
FROM erp_products
WHERE business_id = ?
  AND is_active = TRUE
  AND sale_price > 0;

-- name: GetSlowMovers :many
SELECT p.id, p.name, p.stock_quantity,
       EXISTS (
         SELECT 1 FROM pos_order_items oi
         JOIN pos_orders o ON o.id = oi.order_id AND o.business_id = p.business_id
         WHERE oi.menu_item_id = p.id AND o.created_at >= ?
       ) AS recent_sale
FROM erp_products p
WHERE p.business_id = ?
  AND p.is_active = TRUE
HAVING recent_sale = 0
ORDER BY p.name ASC
LIMIT 500;
