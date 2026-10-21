-- Reads served to Insights reports, so Insights never queries ERP tables.

-- name: InsightsCountProducts :one
SELECT COUNT(*) AS count FROM erp_products WHERE business_id = ?;

-- name: InsightsCountActiveProducts :one
SELECT COUNT(*) AS count FROM erp_products WHERE business_id = ? AND is_active = TRUE;

-- name: InsightsCountOutOfStock :one
SELECT COUNT(*) AS count
FROM erp_products
WHERE business_id = ?
  AND track_inventory = TRUE
  AND stock_quantity <= 0;

-- name: InsightsCountLowStock :one
SELECT COUNT(*) AS count
FROM erp_products
WHERE business_id = ? AND is_active = TRUE AND track_inventory = TRUE
  AND low_stock_threshold IS NOT NULL
  AND stock_quantity <= low_stock_threshold;

-- name: InsightsAvgMarginPct :one
SELECT COALESCE(AVG((sale_price - cost_price) / sale_price * 100), 0) AS margin_pct
FROM erp_products
WHERE business_id = ?
  AND is_active = TRUE
  AND sale_price > 0;

-- name: InsightsMarginDistribution :one
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

-- name: InsightsProductFacts :many
SELECT p.id,
       p.name,
       COALESCE(c.id, 0) AS category_id,
       COALESCE(c.name, '') AS category_name,
       CAST(p.cost_price * 100 AS SIGNED) AS cost_minor_units
FROM erp_products p
LEFT JOIN erp_categories c ON c.id = p.category_id AND c.business_id = p.business_id
WHERE p.business_id = sqlc.arg(business_id)
  AND p.id IN (sqlc.slice(ids));

-- name: InsightsActiveProducts :many
SELECT id, name, stock_quantity
FROM erp_products
WHERE business_id = ?
  AND is_active = TRUE
ORDER BY name ASC, id ASC;

-- name: InsightsERPLastChange :one
SELECT CAST(COALESCE(MAX(updated_at), '1000-01-01') AS DATETIME) AS last_change
FROM erp_products
WHERE business_id = ?;
