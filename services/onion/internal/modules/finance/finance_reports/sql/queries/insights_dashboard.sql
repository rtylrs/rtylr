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

-- name: InsightsPaidInvoicesByDay :many
SELECT reporting_date AS day,
       COALESCE(SUM(total), 0) AS amount
FROM finance_invoices
WHERE business_id = sqlc.arg(business_id)
  AND status = 'paid'
  AND reporting_date >= sqlc.arg(date_from)
  AND reporting_date <= sqlc.arg(date_to)
GROUP BY day
ORDER BY day ASC
LIMIT 200;

-- name: InsightsExpensesByDay :many
SELECT DATE(expense_date) AS day,
       COALESCE(SUM(amount), 0) AS amount
FROM finance_expenses
WHERE business_id = sqlc.arg(business_id)
  AND status = 'approved'
  AND expense_date BETWEEN sqlc.arg(date_from) AND sqlc.arg(date_to)
GROUP BY day;

-- name: InsightsExpensesByCategory :many
SELECT COALESCE(c.id, 0) AS cat_id,
       COALESCE(c.name, 'Uncategorised') AS cat_name,
       COALESCE(SUM(e.amount), 0) AS amount
FROM finance_expenses e
LEFT JOIN finance_expense_categories c ON c.id = e.category_id AND c.business_id = e.business_id
WHERE e.business_id = sqlc.arg(business_id)
  AND e.status = 'approved'
  AND e.expense_date BETWEEN sqlc.arg(date_from) AND sqlc.arg(date_to)
GROUP BY c.id, c.name
ORDER BY SUM(e.amount) DESC;

-- Bucket labels match the insights.financial.v1 InvoiceAgingBucket contract:
-- current (not past due, including due today), 1-30, 31-60, 61-90, 90+.
-- name: InsightsInvoiceAging :many
SELECT
  CASE
    WHEN DATEDIFF(CURDATE(), due_date) <= 0 THEN 'current'
    WHEN DATEDIFF(CURDATE(), due_date) BETWEEN 1 AND 30 THEN '1-30'
    WHEN DATEDIFF(CURDATE(), due_date) BETWEEN 31 AND 60 THEN '31-60'
    WHEN DATEDIFF(CURDATE(), due_date) BETWEEN 61 AND 90 THEN '61-90'
    WHEN DATEDIFF(CURDATE(), due_date) > 90 THEN '90+'
    ELSE 'current'
  END AS bucket,
  COUNT(*) AS invoice_count,
  COALESCE(SUM(amount_due), 0) AS amount_due
FROM finance_invoices
WHERE business_id = sqlc.arg(business_id)
  AND created_at >= DATE_SUB(UTC_TIMESTAMP(), INTERVAL 12 MONTH)
  AND status NOT IN ('paid','cancelled','void')
  AND amount_due > 0
GROUP BY bucket;

-- name: SumInsightsInvoiceRevenue :one
SELECT COALESCE(SUM(total), 0) AS total
FROM finance_invoices
WHERE business_id = sqlc.arg(business_id)
  AND status = 'paid'
  AND reporting_date >= sqlc.arg(date_from)
  AND reporting_date <= sqlc.arg(date_to);

-- The latest change to the finance rows Insights reports on. Insights uses it
-- to version its cached reports.
-- name: InsightsFinanceLastChange :one
SELECT CAST(GREATEST(
  COALESCE((SELECT MAX(i.updated_at) FROM finance_invoices i WHERE i.business_id = sqlc.arg(business_id)), '1000-01-01'),
  COALESCE((SELECT MAX(e.updated_at) FROM finance_expenses e WHERE e.business_id = sqlc.arg(business_id)), '1000-01-01')
) AS DATETIME) AS last_change;
