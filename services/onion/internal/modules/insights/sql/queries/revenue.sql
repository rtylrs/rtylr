-- name: GetRevenueSeries :many
SELECT DATE(COALESCE(paid_date, issue_date)) AS day,
       COALESCE(SUM(total), 0)               AS amount
FROM finance_invoices
WHERE business_id = ?
  AND status = 'paid'
  AND COALESCE(paid_date, issue_date) >= ?
  AND COALESCE(paid_date, issue_date) <= ?
GROUP BY day
ORDER BY day ASC;

-- name: SumPosRevenueMTD :one
SELECT COALESCE(SUM(total), 0) AS total
FROM pos_orders
WHERE business_id = ?
  AND payment_status = 'paid'
  AND created_at >= ?
  AND created_at <= ?;

-- name: SumFinanceRevenueMTD :one
SELECT COALESCE(SUM(total), 0) AS total
FROM finance_invoices
WHERE business_id = ?
  AND status = 'paid'
  AND COALESCE(paid_date, issue_date) >= ?
  AND COALESCE(paid_date, issue_date) <= ?;
