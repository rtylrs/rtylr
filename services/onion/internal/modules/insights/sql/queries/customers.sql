-- name: CountCustomersTotal :one
SELECT COUNT(*) AS count FROM crm_customers WHERE business_id = ?;

-- name: CountCustomersNewThisMonth :one
SELECT COUNT(*) AS count FROM crm_customers WHERE business_id = ? AND created_at >= ?;

-- name: CountCustomersActive :one
SELECT COUNT(DISTINCT c.id) AS count
FROM crm_customers c
JOIN pos_orders o ON o.customer_id = c.id AND o.business_id = c.business_id
WHERE c.business_id = ?
  AND o.created_at >= ?;

-- name: CountCustomersChurned :one
SELECT COUNT(*) AS count
FROM crm_customers c
WHERE c.business_id = ?
  AND EXISTS (SELECT 1 FROM pos_orders o WHERE o.business_id = c.business_id AND o.customer_id = c.id)
  AND NOT EXISTS (SELECT 1 FROM pos_orders o2 WHERE o2.business_id = c.business_id AND o2.customer_id = c.id AND o2.created_at >= ?);

-- name: GetLTVSpendByCustomer :many
SELECT c.id, COALESCE(SUM(o.total), 0) AS spend
FROM crm_customers c
LEFT JOIN pos_orders o ON o.customer_id = c.id AND o.business_id = c.business_id AND o.payment_status = 'paid'
WHERE c.business_id = ?
GROUP BY c.id;

-- name: GetTopCustomers :many
SELECT
  c.id,
  COALESCE(c.first_name, '') AS first_name,
  COALESCE(c.last_name, '') AS last_name,
  COALESCE(c.email, '') AS email,
  COALESCE(SUM(o.total), 0) AS spend,
  COUNT(o.id) AS orders
FROM crm_customers c
LEFT JOIN pos_orders o ON o.customer_id = c.id AND o.business_id = c.business_id AND o.payment_status = 'paid'
WHERE c.business_id = ?
GROUP BY c.id, c.first_name, c.last_name, c.email
ORDER BY spend DESC
LIMIT ?;
