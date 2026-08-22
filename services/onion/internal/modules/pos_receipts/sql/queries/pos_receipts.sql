-- name: ListReceipts :many
SELECT id, order_number, customer_name, total, payment_method, paid_at, business_id
FROM pos_orders
WHERE business_id = ? AND payment_status = 'paid'
ORDER BY paid_at DESC, id DESC
LIMIT 500;

-- name: ListReceiptsWithDateFrom :many
SELECT id, order_number, customer_name, total, payment_method, paid_at, business_id
FROM pos_orders
WHERE business_id = ? AND payment_status = 'paid' AND paid_at >= ?
ORDER BY paid_at DESC, id DESC
LIMIT 500;

-- name: ListReceiptsWithDateRange :many
SELECT id, order_number, customer_name, total, payment_method, paid_at, business_id
FROM pos_orders
WHERE business_id = ? AND payment_status = 'paid' AND paid_at >= ? AND paid_at < ?
ORDER BY paid_at DESC, id DESC
LIMIT 500;

-- name: GetReceiptByID :one
SELECT id, order_number, customer_name, total, payment_method, paid_at, business_id
FROM pos_orders
WHERE id = ? AND business_id = ? AND payment_status = 'paid';
