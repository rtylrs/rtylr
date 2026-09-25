-- name: ListAccounts :many
SELECT id, owner_id, business_id, name, code, type, currency, balance, is_active, created_at
FROM finance_accounts
WHERE business_id = ?
ORDER BY type ASC, name ASC
LIMIT ? OFFSET ?;

-- ListAccountsForTree must return the full account set to build the chart of
-- accounts tree, so it stays unpaginated by design — bounded by the same
-- generous hard LIMIT as above.
-- name: ListAccountsForTree :many
SELECT id, owner_id, business_id, name, code, parent_id, type, currency, balance, is_active, created_at
FROM finance_accounts
WHERE business_id = ?
ORDER BY type ASC, code ASC, name ASC
LIMIT 1000;

-- name: CreateAccount :execresult
INSERT INTO finance_accounts (owner_id, business_id, name, code, type, currency, balance, is_active)
VALUES (?, ?, ?, ?, ?, ?, ?, ?);

-- name: UpdateAccount :execresult
UPDATE finance_accounts
SET name = ?, code = ?, type = ?, currency = ?, balance = ?, is_active = ?
WHERE id = ? AND business_id = ?;

-- name: DeleteAccount :execresult
DELETE FROM finance_accounts
WHERE id = ? AND business_id = ?;

-- name: GetPayrollAccountMapping :one
SELECT business_id, expense_account_id, payable_account_id, deductions_payable_account_id
FROM finance_payroll_account_mappings
WHERE business_id = ?;
