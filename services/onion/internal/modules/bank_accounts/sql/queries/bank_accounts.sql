-- name: ListBankAccounts :many
SELECT id, owner_id, business_id, name, institution, account_type, balance, currency, is_connected, last_synced_at, created_at
FROM finance_bank_accounts
WHERE business_id = ?
ORDER BY name ASC;

-- name: CreateBankAccount :execresult
INSERT INTO finance_bank_accounts (owner_id, business_id, name, institution, account_type, balance, currency, is_connected)
VALUES (?, ?, ?, ?, ?, ?, ?, ?);

-- name: SumBankAccountBalances :one
SELECT COALESCE(SUM(balance), 0)
FROM finance_bank_accounts
WHERE business_id = ?;

-- name: CountConnectedBankAccounts :one
SELECT COUNT(*)
FROM finance_bank_accounts
WHERE business_id = ? AND is_connected = TRUE;
