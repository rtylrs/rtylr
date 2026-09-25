-- name: ListReceipts :many
-- cross-schema read from pos_orders (POS orders module) — list paid orders as receipts for the business.
SELECT id, order_number, customer_name, total, payment_method, paid_at, business_id
FROM pos_orders
WHERE business_id = ? AND payment_status = 'paid'
ORDER BY paid_at DESC, id DESC
LIMIT 500;

-- name: ListReceiptsWithDateFrom :many
-- cross-schema read from pos_orders (POS orders module) — list paid orders as receipts from the requested date.
SELECT id, order_number, customer_name, total, payment_method, paid_at, business_id
FROM pos_orders
WHERE business_id = ? AND payment_status = 'paid' AND paid_at >= ?
ORDER BY paid_at DESC, id DESC
LIMIT 500;

-- name: ListReceiptsWithDateRange :many
-- cross-schema read from pos_orders (POS orders module) — list paid orders as receipts within the requested date range.
SELECT id, order_number, customer_name, total, payment_method, paid_at, business_id
FROM pos_orders
WHERE business_id = ? AND payment_status = 'paid' AND paid_at >= ? AND paid_at < ?
ORDER BY paid_at DESC, id DESC
LIMIT 500;

-- name: GetReceiptByID :one
-- cross-schema read from pos_orders (POS orders module) — fetch one paid order for receipt display.
SELECT id, order_number, customer_name, total, payment_method, paid_at, business_id
FROM pos_orders
WHERE id = ? AND business_id = ? AND payment_status = 'paid';

-- name: ListReceiptsWithDateTo :many
-- cross-schema read from pos_orders (POS orders module) — list paid orders as receipts up to the requested date.
SELECT id, order_number, customer_name, total, payment_method, paid_at, business_id
FROM pos_orders
WHERE business_id = ? AND payment_status = 'paid' AND paid_at < ?
ORDER BY paid_at DESC, id DESC
LIMIT 500;
