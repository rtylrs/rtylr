-- name: GetDaySummaryToday :one
SELECT
    COALESCE(SUM(total), 0) AS total_sales,
    COALESCE(SUM(tax_amount), 0) AS total_tax,
    COUNT(*) AS total_orders
FROM pos_orders
WHERE business_id = ? AND payment_status = 'paid' AND DATE(paid_at) = CURDATE();

-- name: GetDaySummaryByDate :one
SELECT
    COALESCE(SUM(total), 0) AS total_sales,
    COALESCE(SUM(tax_amount), 0) AS total_tax,
    COUNT(*) AS total_orders
FROM pos_orders
WHERE business_id = ? AND payment_status = 'paid' AND DATE(paid_at) = ?;

-- name: GetRefundTotalToday :one
SELECT COALESCE(SUM(amount), 0) AS total_refunds FROM pos_refunds
WHERE business_id = ? AND DATE(refunded_at) = CURDATE();

-- name: GetRefundTotalByDate :one
SELECT COALESCE(SUM(amount), 0) AS total_refunds FROM pos_refunds
WHERE business_id = ? AND DATE(refunded_at) = ?;

-- name: GetPaymentBreakdownToday :many
SELECT payment_method, COALESCE(SUM(total), 0) AS method_total
FROM pos_orders
WHERE business_id = ? AND payment_status = 'paid' AND DATE(paid_at) = CURDATE()
GROUP BY payment_method;

-- name: GetPaymentBreakdownByDate :many
SELECT payment_method, COALESCE(SUM(total), 0) AS method_total
FROM pos_orders
WHERE business_id = ? AND payment_status = 'paid' AND DATE(paid_at) = ?
GROUP BY payment_method;
