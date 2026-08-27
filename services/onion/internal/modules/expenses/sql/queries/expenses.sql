-- name: ListExpenses :many
SELECT id, owner_id, business_id, category_id, description, amount, currency, expense_date, vendor, payment_method, receipt_url, is_tax_deductible, tax_amount, notes, status, created_at, updated_at
FROM finance_expenses
WHERE business_id = ?
ORDER BY expense_date DESC
LIMIT 500;

-- name: GetExpenseByID :one
SELECT id, owner_id, business_id, category_id, description, amount, currency, expense_date, vendor, payment_method, receipt_url, is_tax_deductible, tax_amount, notes, status, created_at, updated_at
FROM finance_expenses
WHERE id = ? AND business_id = ?;

-- name: CreateExpense :execresult
INSERT INTO finance_expenses (owner_id, business_id, category_id, description, amount, currency, expense_date, vendor, payment_method, receipt_url, is_tax_deductible, tax_amount, notes, status)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'pending');

-- name: UpdateExpense :execresult
UPDATE finance_expenses SET category_id = ?, description = ?, amount = ?, currency = ?, expense_date = ?, vendor = ?, payment_method = ?, receipt_url = ?, is_tax_deductible = ?, tax_amount = ?, notes = ?
WHERE id = ? AND business_id = ? AND status <> 'approved';

-- name: DeleteExpense :execresult
DELETE FROM finance_expenses WHERE id = ? AND business_id = ? AND status <> 'approved';

-- name: UpdateExpenseStatus :execresult
UPDATE finance_expenses
SET status = sqlc.arg(status)
WHERE id = sqlc.arg(id)
  AND business_id = sqlc.arg(business_id)
  AND status = sqlc.arg(current_status);

-- name: ListExpenseCategories :many
SELECT id, owner_id, business_id, name, color, created_at
FROM finance_expense_categories
WHERE business_id = ?
ORDER BY name;

-- name: SumExpensesByBusiness :one
SELECT COALESCE(SUM(CAST(amount AS DECIMAL(14,2))), 0) AS total
FROM finance_expenses
WHERE business_id = ?;

-- name: CountExpensesByCategory :many
SELECT COALESCE(CAST(category_id AS CHAR), 'uncategorised') AS category_key, COUNT(*) AS cnt
FROM finance_expenses
WHERE business_id = ?
  AND created_at >= DATE_SUB(UTC_TIMESTAMP(), INTERVAL 12 MONTH)
GROUP BY category_id;

-- name: ListExpensesPaginated :many
SELECT id, owner_id, business_id, category_id, description, amount, currency, expense_date, vendor, payment_method, receipt_url, is_tax_deductible, tax_amount, notes, status, created_at, updated_at
FROM finance_expenses
WHERE business_id = ?
ORDER BY expense_date DESC
LIMIT ? OFFSET ?;

-- name: CountExpenses :one
SELECT COUNT(*) FROM finance_expenses WHERE business_id = ?;
