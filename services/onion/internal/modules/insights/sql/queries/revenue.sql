-- name: GetRevenueSeries :many
SELECT reporting_date AS day,
       COALESCE(SUM(total), 0)               AS amount
FROM finance_invoices
WHERE business_id = ?
  AND status = 'paid'
  AND reporting_date >= ?
  AND reporting_date <= ?
GROUP BY day
ORDER BY day ASC
LIMIT 200;

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
  AND reporting_date >= ?
  AND reporting_date <= ?;
