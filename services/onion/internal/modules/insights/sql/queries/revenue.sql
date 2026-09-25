-- cross-schema read from finance — build the paid-invoice revenue series.
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

-- cross-schema read from pos — calculate net POS revenue after refunds.
-- name: SumPosRevenueMTD :one
SELECT COALESCE(SUM(total), 0) AS total
FROM (
    SELECT GREATEST(o.total - COALESCE(SUM(r.amount), 0), 0) AS total
    FROM pos_orders o
    LEFT JOIN pos_refunds r ON r.order_id = o.id AND r.business_id = o.business_id
    WHERE o.business_id = ?
      AND o.payment_status IN ('paid', 'refunded')
      AND o.paid_at >= ?
      AND o.paid_at <= ?
    GROUP BY o.id, o.total
) AS net_pos_orders;

-- cross-schema read from finance — calculate month-to-date invoice revenue.
-- name: SumFinanceRevenueMTD :one
SELECT COALESCE(SUM(total), 0) AS total
FROM finance_invoices
WHERE business_id = ?
  AND status = 'paid'
  AND reporting_date >= ?
  AND reporting_date <= ?;
