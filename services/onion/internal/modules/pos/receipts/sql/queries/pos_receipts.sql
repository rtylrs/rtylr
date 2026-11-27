-- name: ListReceipts :many
-- cross-schema read from pos_orders (POS orders module) — list paid orders and online menu orders as receipts for the business.
SELECT id, order_number, customer_name, total, payment_method, paid_at, created_at, business_id
FROM pos_orders
WHERE business_id = ?
  AND (
    payment_status = 'paid'
    OR ((session_id IS NULL OR type = 'online') AND status NOT IN ('cancelled', 'voided'))
  )
ORDER BY COALESCE(paid_at, created_at) DESC, id DESC
LIMIT 500;

-- name: ListReceiptsWithDateFrom :many
-- cross-schema read from pos_orders (POS orders module) — list paid and online orders as receipts from the requested date.
SELECT id, order_number, customer_name, total, payment_method, paid_at, created_at, business_id
FROM pos_orders
WHERE business_id = ?
  AND (
    payment_status = 'paid'
    OR ((session_id IS NULL OR type = 'online') AND status NOT IN ('cancelled', 'voided'))
  )
  AND COALESCE(paid_at, created_at) >= ?
ORDER BY COALESCE(paid_at, created_at) DESC, id DESC
LIMIT 500;

-- name: ListReceiptsWithDateRange :many
-- cross-schema read from pos_orders (POS orders module) — list paid and online orders as receipts within the requested date range.
SELECT id, order_number, customer_name, total, payment_method, paid_at, created_at, business_id
FROM pos_orders
WHERE business_id = ?
  AND (
    payment_status = 'paid'
    OR ((session_id IS NULL OR type = 'online') AND status NOT IN ('cancelled', 'voided'))
  )
  AND COALESCE(paid_at, created_at) >= ?
  AND COALESCE(paid_at, created_at) < ?
ORDER BY COALESCE(paid_at, created_at) DESC, id DESC
LIMIT 500;

-- name: GetReceiptByID :one
-- cross-schema read from pos_orders (POS orders module) — fetch one paid or online order for receipt display.
SELECT id, order_number, customer_name, total, payment_method, paid_at, created_at, business_id
FROM pos_orders
WHERE id = ?
  AND business_id = ?
  AND (
    payment_status = 'paid'
    OR ((session_id IS NULL OR type = 'online') AND status NOT IN ('cancelled', 'voided'))
  );

-- name: ListReceiptsWithDateTo :many
-- cross-schema read from pos_orders (POS orders module) — list paid and online orders as receipts up to the requested date.
SELECT id, order_number, customer_name, total, payment_method, paid_at, created_at, business_id
FROM pos_orders
WHERE business_id = ?
  AND (
    payment_status = 'paid'
    OR ((session_id IS NULL OR type = 'online') AND status NOT IN ('cancelled', 'voided'))
  )
  AND COALESCE(paid_at, created_at) < ?
ORDER BY COALESCE(paid_at, created_at) DESC, id DESC
LIMIT 500;
