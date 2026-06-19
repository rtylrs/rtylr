-- name: ListExpenses :many
SELECT id, owner_id, business_id, category_id, amount, currency, description, vendor, receipt_url, expense_date, status, created_at, updated_at
FROM finance_expenses
WHERE business_id = ?
ORDER BY expense_date DESC;

-- name: GetExpenseByID :one
SELECT id, owner_id, business_id, category_id, amount, currency, description, vendor, receipt_url, expense_date, status, created_at, updated_at
FROM finance_expenses
WHERE id = ? AND business_id = ?;

-- name: CreateExpense :execresult
INSERT INTO finance_expenses (owner_id, business_id, category_id, amount, currency, description, vendor, receipt_url, expense_date, status)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 'pending');

-- name: UpdateExpense :execresult
UPDATE finance_expenses SET category_id = ?, amount = ?, currency = ?, description = ?, vendor = ?, receipt_url = ?, expense_date = ?
WHERE id = ? AND business_id = ?;

-- name: DeleteExpense :execresult
DELETE FROM finance_expenses WHERE id = ? AND business_id = ?;

-- name: UpdateExpenseStatus :execresult
UPDATE finance_expenses SET status = ? WHERE id = ? AND business_id = ?;

-- name: ListExpenseCategories :many
SELECT id, owner_id, business_id, name, color, created_at
FROM finance_expense_categories
WHERE business_id = ?
ORDER BY name;
