-- cross-schema read from erp — count all product SKUs for insight totals.
-- name: CountProductSKUsTotal :one
SELECT COUNT(*) AS count FROM erp_products WHERE business_id = ?;

-- cross-schema read from erp — count active product SKUs for insight totals.
-- name: CountProductSKUsActive :one
SELECT COUNT(*) AS count FROM erp_products WHERE business_id = ? AND is_active = TRUE;

-- cross-schema read from erp — count tracked products with no stock.
-- name: CountProductSKUsOutOfStock :one
SELECT COUNT(*) AS count
 FROM erp_products
 WHERE business_id = ?
   AND track_inventory = TRUE
   AND stock_quantity <= 0;

-- cross-schema read from erp — calculate the active product margin average.
-- name: AvgProductMarginPct :one
SELECT COALESCE(AVG((sale_price - cost_price) / sale_price * 100), 0) AS margin_pct
 FROM erp_products
 WHERE business_id = ?
   AND is_active = TRUE
   AND sale_price > 0;

-- cross-schema read from erp and pos — combine products with paid order volume.
-- name: GetBestSellers :many
SELECT p.id,
       p.name,
       COALESCE(SUM(oi.quantity), 0) AS units,
       COALESCE(SUM(oi.total), 0)    AS revenue
FROM erp_products p
JOIN pos_order_items oi ON oi.menu_item_id = p.id
JOIN pos_orders o       ON o.id = oi.order_id AND o.business_id = p.business_id
WHERE p.business_id = ?
  AND o.payment_status = 'paid'
  AND oi.status <> 'cancelled'
  AND o.created_at >= ?
GROUP BY p.id, p.name
ORDER BY units DESC
LIMIT ?;

-- cross-schema read from erp and pos — aggregate paid order volume by product category.
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
  AND o.payment_status = 'paid'
  AND oi.status <> 'cancelled'
  AND o.created_at >= ?
  AND o.created_at <= ?
GROUP BY c.id, c.name
ORDER BY revenue DESC;

-- cross-schema read from erp — calculate margin distribution buckets for active products.
-- name: GetProductMarginDistribution :one
SELECT
  CAST(COALESCE(SUM(margin_pct >= -1000 AND margin_pct < 0), 0) AS SIGNED) AS below_zero_count,
  CAST(COALESCE(SUM(margin_pct >= 0 AND margin_pct < 10), 0) AS SIGNED) AS zero_to_ten_count,
  CAST(COALESCE(SUM(margin_pct >= 10 AND margin_pct < 25), 0) AS SIGNED) AS ten_to_twenty_five_count,
  CAST(COALESCE(SUM(margin_pct >= 25 AND margin_pct < 50), 0) AS SIGNED) AS twenty_five_to_fifty_count,
  CAST(COALESCE(SUM(margin_pct >= 50 AND margin_pct < 75), 0) AS SIGNED) AS fifty_to_seventy_five_count,
  CAST(COALESCE(SUM(margin_pct >= 75 AND margin_pct < 1000), 0) AS SIGNED) AS seventy_five_plus_count
FROM (
  SELECT (sale_price - cost_price) / sale_price * 100 AS margin_pct
  FROM erp_products
  WHERE business_id = ?
    AND is_active = TRUE
    AND sale_price > 0
) AS active_product_margins;

-- cross-schema read from erp and pos — find active products without recent sales.
-- name: GetSlowMovers :many
SELECT p.id, p.name, p.stock_quantity,
       FALSE AS recent_sale
FROM erp_products p
LEFT JOIN (
    SELECT DISTINCT oi.menu_item_id
    FROM pos_order_items oi
    JOIN pos_orders o ON o.id = oi.order_id
    WHERE o.business_id = ? AND o.created_at >= ?
) AS recent_sales ON recent_sales.menu_item_id = p.id
WHERE p.business_id = ?
  AND p.is_active = TRUE
  AND recent_sales.menu_item_id IS NULL
ORDER BY p.name ASC
LIMIT 500;

-- cross-schema read from erp and pos — combine product costs with all-time sales volume.
-- name: GetProductMarginRows :many
-- Per-product margin rows (#853): joins sales volume to cost price so the
-- response carries per-item economics alongside the distribution buckets.
-- All-time window; the request message carries no date params.
SELECT p.id,
       p.name,
       COALESCE(SUM(oi.quantity), 0)     AS units_sold,
       COALESCE(SUM(oi.total), 0)        AS revenue_minor_units_x100,
       CAST(p.cost_price * 100 AS SIGNED) AS cost_minor_units_x100
FROM erp_products p
JOIN pos_order_items oi ON oi.menu_item_id = p.id
JOIN pos_orders o       ON o.id = oi.order_id AND o.business_id = p.business_id
WHERE p.business_id = ?
GROUP BY p.id, p.name, p.cost_price
HAVING SUM(oi.quantity) > 0
ORDER BY revenue DESC
LIMIT 200;
