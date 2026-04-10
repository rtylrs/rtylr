-- name: ListBankAccounts :many
SELECT id, owner_id, name, institution, account_type, balance, currency, is_connected, last_synced_at, created_at
FROM finance_bank_accounts
WHERE owner_id = ?
ORDER BY name ASC;

-- name: CreateBankAccount :execresult
INSERT INTO finance_bank_accounts (owner_id, name, institution, account_type, balance, currency, is_connected)
VALUES (?, ?, ?, ?, ?, ?, ?);
