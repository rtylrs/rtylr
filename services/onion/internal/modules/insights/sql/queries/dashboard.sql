-- name: SumPaidInvoices :one
SELECT COALESCE(SUM(total), 0) AS total
FROM finance_invoices
WHERE business_id = ?
  AND status = 'paid'
  AND COALESCE(paid_date, issue_date) >= ?
  AND COALESCE(paid_date, issue_date) <= ?;

-- name: CountPosOrders :one
SELECT COUNT(*) AS count
FROM pos_orders
WHERE business_id = ?
  AND created_at >= ?
  AND created_at <= ?;

-- name: CountCrmCustomersCreated :one
SELECT COUNT(*) AS count
FROM crm_customers
WHERE business_id = ?
  AND created_at >= ?
  AND created_at <= ?;

-- name: SumExpenses :one
SELECT COALESCE(SUM(amount), 0) AS total
FROM finance_expenses
WHERE business_id = ?
  AND expense_date >= ?
  AND expense_date <= ?;
