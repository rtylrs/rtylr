-- name: GetExpensesByCategory :many
SELECT COALESCE(c.id, 0) AS cat_id,
       COALESCE(c.name, 'Uncategorised') AS cat_name,
       COALESCE(SUM(e.amount), 0) AS amount
FROM finance_expenses e
LEFT JOIN finance_expense_categories c ON c.id = e.category_id AND c.business_id = e.business_id
WHERE e.business_id = ?
  AND e.expense_date BETWEEN ? AND ?
GROUP BY c.id, c.name
ORDER BY SUM(e.amount) DESC;

-- name: GetInvoiceAging :many
SELECT
  CASE
    WHEN DATEDIFF(CURDATE(), due_date) BETWEEN 0 AND 30 THEN '0-30'
    WHEN DATEDIFF(CURDATE(), due_date) BETWEEN 31 AND 60 THEN '31-60'
    WHEN DATEDIFF(CURDATE(), due_date) BETWEEN 61 AND 90 THEN '61-90'
    WHEN DATEDIFF(CURDATE(), due_date) > 90 THEN '90+'
    ELSE 'not_due'
  END AS bucket,
  COUNT(*) AS invoice_count,
  COALESCE(SUM(amount_due), 0) AS amount_due
FROM finance_invoices
WHERE business_id = ?
  AND created_at >= DATE_SUB(UTC_TIMESTAMP(), INTERVAL 12 MONTH)
  AND status NOT IN ('paid','cancelled','void')
  AND amount_due > 0
GROUP BY bucket;

-- name: GetDailyPaidInvoices :many
SELECT reporting_date AS day,
       COALESCE(SUM(total), 0) AS amount
FROM finance_invoices
WHERE business_id = ?
  AND status = 'paid'
  AND reporting_date >= ?
  AND reporting_date <= ?
GROUP BY day
LIMIT 200;

-- name: GetDailyExpenses :many
SELECT DATE(expense_date) AS day,
       COALESCE(SUM(amount), 0) AS amount
FROM finance_expenses
WHERE business_id = ?
  AND expense_date BETWEEN ? AND ?
GROUP BY day;
