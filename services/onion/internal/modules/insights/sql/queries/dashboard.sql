-- cross-schema read from finance — total paid invoices for the dashboard period.
-- name: SumPaidInvoices :one
SELECT COALESCE(SUM(total), 0) AS total
FROM finance_invoices
WHERE business_id = ?
  AND status = 'paid'
  AND COALESCE(paid_date, issue_date) >= ?
  AND COALESCE(paid_date, issue_date) <= ?;

-- cross-schema read from pos — count orders for the dashboard period.
-- name: CountPosOrders :one
SELECT COUNT(*) AS count
FROM pos_orders
WHERE business_id = ?
  AND created_at >= ?
  AND created_at <= ?;

-- cross-schema read from crm — count customers created in the dashboard period.
-- name: CountCrmCustomersCreated :one
SELECT COUNT(*) AS count
FROM crm_customers
WHERE business_id = ?
  AND created_at >= ?
  AND created_at <= ?;

-- cross-schema read from finance — total expenses for the dashboard period.
-- name: SumExpenses :one
SELECT COALESCE(SUM(amount), 0) AS total
FROM finance_expenses
WHERE business_id = ?
  AND expense_date >= ?
  AND expense_date <= ?;
