-- Reads served to Insights reports, so Insights never queries POS tables.
-- Customer figures count the customer recorded on each paid order.

-- Net POS revenue after refunds.
-- name: InsightsNetRevenue :one
SELECT COALESCE(SUM(total), 0) AS total
FROM (
    SELECT GREATEST(o.total - COALESCE(SUM(r.amount), 0), 0) AS total
    FROM pos_orders o
    LEFT JOIN pos_refunds r ON r.order_id = o.id AND r.business_id = o.business_id
    WHERE o.business_id = sqlc.arg(business_id)
      AND o.payment_status IN ('paid', 'refunded')
      AND o.paid_at >= sqlc.arg(date_from)
      AND o.paid_at <= sqlc.arg(date_to)
    GROUP BY o.id, o.total
) AS net_pos_orders;

-- name: InsightsCountOrdersToday :one
SELECT COUNT(*) FROM pos_orders
WHERE business_id = ?
  AND created_at >= CURDATE()
  AND created_at < CURDATE() + INTERVAL 1 DAY;

-- name: InsightsSumRevenueToday :one
SELECT COALESCE(SUM(total), 0) FROM pos_orders
WHERE business_id = ?
  AND created_at >= CURDATE()
  AND created_at < CURDATE() + INTERVAL 1 DAY;

-- Customers with a paid order during the calendar days date_from to date_to.
-- name: InsightsCountBuyers :one
SELECT COUNT(DISTINCT customer_id) AS count
FROM pos_orders
WHERE business_id = sqlc.arg(business_id)
  AND customer_id IS NOT NULL
  AND payment_status = 'paid'
  AND created_at >= sqlc.arg(date_from)
  AND created_at < DATE_ADD(sqlc.arg(date_to), INTERVAL 1 DAY);

-- Customers who paid for an order before date_from but not during the
-- calendar days date_from to date_to.
-- name: InsightsCountLapsedBuyers :one
SELECT COUNT(*) AS count
FROM (
  SELECT customer_id
  FROM pos_orders
  WHERE business_id = sqlc.arg(business_id)
    AND customer_id IS NOT NULL
    AND payment_status = 'paid'
    AND created_at < DATE_ADD(sqlc.arg(date_to), INTERVAL 1 DAY)
  GROUP BY customer_id
  HAVING MAX(created_at) < sqlc.arg(date_from)
) AS lapsed;

-- The allow-listed ordering and numeric direction are bound parameters rather
-- than interpolated SQL identifiers.
-- name: InsightsTopBuyers :many
SELECT
  customer_id,
  COALESCE(MAX(customer_name), '') AS customer_name,
  SUM(total) AS spend,
  COUNT(id) AS orders,
  MAX(created_at) AS last_order_at
FROM pos_orders
WHERE business_id = sqlc.arg(business_id)
  AND customer_id IS NOT NULL
  AND payment_status = 'paid'
  AND created_at >= sqlc.arg(date_from)
  AND created_at < DATE_ADD(sqlc.arg(date_to), INTERVAL 1 DAY)
GROUP BY customer_id
ORDER BY
  (CASE WHEN sqlc.arg(by_order_count) THEN COUNT(id) ELSE SUM(total) END) * sqlc.arg(direction) DESC,
  customer_id ASC
LIMIT ? OFFSET ?;

-- Lifetime paid spend per customer, bucketed for the LTV histogram. Only
-- customers with a paid order are counted.
-- name: InsightsBuyerSpendBuckets :many
WITH buckets AS (
  SELECT 0 AS bucket_order, '0-100' AS label, 0 AS lower_minor_units, 10000 AS upper_minor_units
  UNION ALL SELECT 1, '100-500', 10000, 50000
  UNION ALL SELECT 2, '500-1k', 50000, 100000
  UNION ALL SELECT 3, '1k-5k', 100000, 500000
  UNION ALL SELECT 4, '5k-10k', 500000, 1000000
  UNION ALL SELECT 5, '10k+', 1000000, 0
),
customer_spend AS (
  SELECT customer_id, COALESCE(SUM(total), 0) * 100 AS spend_minor_units
  FROM pos_orders
  WHERE business_id = ?
    AND customer_id IS NOT NULL
    AND payment_status = 'paid'
  GROUP BY customer_id
)
SELECT b.label, b.lower_minor_units, b.upper_minor_units, COUNT(cs.customer_id) AS customer_count
FROM buckets b
LEFT JOIN customer_spend cs
  ON cs.spend_minor_units >= b.lower_minor_units
 AND (b.upper_minor_units = 0 OR cs.spend_minor_units < b.upper_minor_units)
GROUP BY b.bucket_order, b.label, b.lower_minor_units, b.upper_minor_units
ORDER BY b.bucket_order;

-- Paid, non-cancelled line items per product since date_from, and up to
-- date_to when it is set.
-- name: InsightsPaidProductSales :many
SELECT oi.menu_item_id AS product_id,
       COALESCE(SUM(oi.quantity), 0) AS units,
       COALESCE(SUM(oi.total), 0)    AS revenue
FROM pos_order_items oi
JOIN pos_orders o ON o.id = oi.order_id
WHERE o.business_id = sqlc.arg(business_id)
  AND o.payment_status = 'paid'
  AND oi.status <> 'cancelled'
  AND oi.menu_item_id IS NOT NULL
  AND o.created_at >= sqlc.arg(date_from)
  AND (sqlc.narg(date_to) IS NULL OR o.created_at <= sqlc.narg(date_to))
GROUP BY oi.menu_item_id;

-- All-time line items per product, whatever the order or item status.
-- name: InsightsLifetimeProductSales :many
SELECT oi.menu_item_id AS product_id,
       COALESCE(SUM(oi.quantity), 0) AS units,
       CAST(COALESCE(SUM(oi.total), 0) * 100 AS SIGNED) AS revenue_minor_units
FROM pos_order_items oi
JOIN pos_orders o ON o.id = oi.order_id
WHERE o.business_id = ?
  AND oi.menu_item_id IS NOT NULL
GROUP BY oi.menu_item_id
HAVING SUM(oi.quantity) > 0;

-- name: InsightsProductsSoldSince :many
SELECT DISTINCT oi.menu_item_id
FROM pos_order_items oi
JOIN pos_orders o ON o.id = oi.order_id
WHERE o.business_id = ?
  AND o.created_at >= ?
  AND oi.menu_item_id IS NOT NULL;

-- name: InsightsPOSLastChange :one
SELECT CAST(GREATEST(
  COALESCE((SELECT MAX(o.updated_at) FROM pos_orders o WHERE o.business_id = sqlc.arg(business_id)), '1000-01-01'),
  COALESCE((SELECT MAX(r.refunded_at) FROM pos_refunds r WHERE r.business_id = sqlc.arg(business_id)), '1000-01-01')
) AS DATETIME) AS last_change;
