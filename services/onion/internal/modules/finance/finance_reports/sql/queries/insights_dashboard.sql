-- name: SumInsightsPaidInvoices :one
SELECT COALESCE(SUM(total), 0)
FROM finance_invoices
WHERE business_id = sqlc.arg(business_id)
  AND status = 'paid'
  AND COALESCE(paid_date, issue_date) >= sqlc.arg(date_from)
  AND COALESCE(paid_date, issue_date) <= sqlc.arg(date_to);

-- name: SumInsightsExpenses :one
SELECT COALESCE(SUM(amount), 0)
FROM finance_expenses
WHERE business_id = sqlc.arg(business_id)
  AND status = 'approved'
  AND expense_date >= sqlc.arg(date_from)
  AND expense_date <= sqlc.arg(date_to);
