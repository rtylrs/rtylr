-- name: GetDaySummaryToday :one
SELECT
    COALESCE(SUM(total), 0) AS total_sales,
    COALESCE(SUM(tax_amount), 0) AS total_tax,
    COALESCE(SUM(tip_amount), 0) AS total_tips,
    COUNT(*) AS total_orders
FROM pos_orders
WHERE business_id = ? AND payment_status = 'paid'
  AND paid_at >= CURDATE()
  AND paid_at < CURDATE() + INTERVAL 1 DAY;

-- name: GetDaySummaryByDate :one
SELECT
    COALESCE(SUM(total), 0) AS total_sales,
    COALESCE(SUM(tax_amount), 0) AS total_tax,
    COALESCE(SUM(tip_amount), 0) AS total_tips,
    COUNT(*) AS total_orders
FROM pos_orders
WHERE business_id = ? AND payment_status = 'paid'
  AND paid_at >= sqlc.arg(paid_at)
  AND paid_at < sqlc.arg(paid_at_end);

-- name: GetRefundTotalToday :one
SELECT COALESCE(SUM(r.amount), 0) AS total_refunds FROM pos_refunds r
JOIN pos_orders o ON o.id = r.order_id AND o.business_id = r.business_id
WHERE r.business_id = ?
  AND refunded_at >= CURDATE()
  AND refunded_at < CURDATE() + INTERVAL 1 DAY;

-- name: GetRefundTotalByDate :one
SELECT COALESCE(SUM(r.amount), 0) AS total_refunds FROM pos_refunds r
JOIN pos_orders o ON o.id = r.order_id AND o.business_id = r.business_id
WHERE r.business_id = ?
  AND refunded_at >= sqlc.arg(refunded_at)
  AND refunded_at < sqlc.arg(refunded_at_end);

-- name: GetPaymentBreakdownToday :many
SELECT payment_method, COALESCE(SUM(total), 0) AS method_total
FROM pos_orders
WHERE business_id = ? AND payment_status = 'paid'
  AND paid_at >= CURDATE()
  AND paid_at < CURDATE() + INTERVAL 1 DAY
GROUP BY payment_method
LIMIT 200;

-- name: GetPaymentBreakdownByDate :many
SELECT payment_method, COALESCE(SUM(total), 0) AS method_total
FROM pos_orders
WHERE business_id = ? AND payment_status = 'paid'
  AND paid_at >= sqlc.arg(paid_at)
  AND paid_at < sqlc.arg(paid_at_end)
GROUP BY payment_method
LIMIT 200;
