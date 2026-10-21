-- cross-schema read from journal — an account's current balance is its opening
-- balance plus the posted lines on it.
-- name: ListAccounts :many
SELECT a.id, a.owner_id, a.business_id, a.name, a.code, a.type, a.currency, a.balance, a.is_active, a.created_at,
       CAST(COALESCE(p.debit, 0) - COALESCE(p.credit, 0) AS DECIMAL(15,2)) AS posted_net_debit
FROM finance_accounts a
LEFT JOIN (
    SELECT l.account_id,
           SUM(CAST(l.debit AS DECIMAL(15,2))) AS debit,
           SUM(CAST(l.credit AS DECIMAL(15,2))) AS credit
    FROM finance_journal_lines l
    JOIN finance_journal_entries e ON e.id = l.entry_id
    WHERE e.business_id = sqlc.arg(business_id)
      AND e.status = 'posted'
    GROUP BY l.account_id
) p ON p.account_id = a.id
WHERE a.business_id = sqlc.arg(business_id)
ORDER BY a.type ASC, a.code ASC, a.name ASC
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

-- name: FindLedgerAccount :one
SELECT id
FROM finance_accounts
WHERE business_id = sqlc.arg(business_id)
  AND code = sqlc.arg(code)
  AND type = sqlc.arg(account_type)
  AND COALESCE(NULLIF(currency, ''), 'USD') = sqlc.arg(currency)
ORDER BY id ASC
LIMIT 1;
