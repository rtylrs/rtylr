-- cross-schema read from crm — count customers that existed by the end of a period.
-- name: CountCustomersAsOf :one
SELECT COUNT(*) AS count
FROM crm_customers
WHERE business_id = ?
  AND created_at < DATE_ADD(?, INTERVAL 1 DAY);

-- cross-schema read from crm — count customers created during a period.
-- name: CountCustomersNewInPeriod :one
SELECT COUNT(*) AS count
FROM crm_customers
WHERE business_id = ?
  AND created_at >= ?
  AND created_at < DATE_ADD(?, INTERVAL 1 DAY);

-- cross-schema read from crm and pos — count customers with paid orders during a period.
-- name: CountCustomersActive :one
SELECT COUNT(DISTINCT c.id) AS count
FROM crm_customers c
JOIN pos_orders o
  ON o.customer_id = c.id
 AND o.business_id = c.business_id
 AND o.payment_status = 'paid'
WHERE c.business_id = ?
  AND o.created_at >= ?
  AND o.created_at < DATE_ADD(?, INTERVAL 1 DAY);

-- cross-schema read from crm and pos — count previously purchasing customers with no paid order during a period.
-- name: CountCustomersChurned :one
SELECT COUNT(*) AS count
FROM crm_customers c
WHERE c.business_id = ?
  AND EXISTS (
    SELECT 1
    FROM pos_orders o
    WHERE o.business_id = c.business_id
      AND o.customer_id = c.id
      AND o.payment_status = 'paid'
      AND o.created_at < ?
  )
  AND NOT EXISTS (
    SELECT 1
    FROM pos_orders o2
    WHERE o2.business_id = c.business_id
      AND o2.customer_id = c.id
      AND o2.payment_status = 'paid'
      AND o2.created_at >= ?
      AND o2.created_at < DATE_ADD(?, INTERVAL 1 DAY)
  );

-- cross-schema read from crm and pos — bucket customer lifetime spend.
-- name: GetLTVDistribution :many
WITH buckets AS (
  SELECT 0 AS bucket_order, '0-100' AS label, 0 AS lower_minor_units, 10000 AS upper_minor_units
  UNION ALL SELECT 1, '100-500', 10000, 50000
  UNION ALL SELECT 2, '500-1k', 50000, 100000
  UNION ALL SELECT 3, '1k-5k', 100000, 500000
  UNION ALL SELECT 4, '5k-10k', 500000, 1000000
  UNION ALL SELECT 5, '10k+', 1000000, 0
),
customer_spend AS (
  SELECT c.id, COALESCE(SUM(o.total), 0) * 100 AS spend_minor_units
  FROM crm_customers c
  LEFT JOIN pos_orders o
    ON o.business_id = c.business_id
   AND o.customer_id = c.id
   AND o.payment_status = 'paid'
  WHERE c.business_id = ?
  GROUP BY c.id
)
SELECT b.label, b.lower_minor_units, b.upper_minor_units, COUNT(cs.id) AS customer_count
FROM buckets b
LEFT JOIN customer_spend cs
  ON cs.spend_minor_units >= b.lower_minor_units
 AND (b.upper_minor_units = 0 OR cs.spend_minor_units < b.upper_minor_units)
GROUP BY b.bucket_order, b.label, b.lower_minor_units, b.upper_minor_units
ORDER BY b.bucket_order;

-- cross-schema read from crm and pos — count customers active during a period.
-- name: CountCustomersWithPaidOrders :one
SELECT COUNT(DISTINCT c.id) AS count
FROM crm_customers c
JOIN pos_orders o
  ON o.customer_id = c.id
 AND o.business_id = c.business_id
 AND o.payment_status = 'paid'
 AND o.created_at >= ?
 AND o.created_at < DATE_ADD(?, INTERVAL 1 DAY)
WHERE c.business_id = ?;

-- cross-schema read from crm and pos — rank customers in a period.
-- The allow-listed column and numeric direction are bound parameters rather than
-- interpolated SQL identifiers.
-- name: GetTopCustomers :many
SELECT
  c.id,
  COALESCE(c.first_name, '') AS first_name,
  COALESCE(c.last_name, '') AS last_name,
  COALESCE(c.email, '') AS email,
  SUM(o.total) AS spend,
  COUNT(o.id) AS orders,
  MAX(o.created_at) AS last_order_at
FROM crm_customers c
JOIN pos_orders o
  ON o.customer_id = c.id
 AND o.business_id = c.business_id
 AND o.payment_status = 'paid'
 AND o.created_at >= ?
 AND o.created_at < DATE_ADD(?, INTERVAL 1 DAY)
WHERE c.business_id = ?
GROUP BY c.id, c.first_name, c.last_name, c.email
ORDER BY
  (CASE WHEN sqlc.arg(by_order_count) THEN COUNT(o.id) ELSE SUM(o.total) END) * sqlc.arg(direction) DESC,
  c.id ASC
LIMIT ? OFFSET ?;
