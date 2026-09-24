-- name: ListBankAccounts :many
SELECT id, owner_id, business_id, name, institution, account_type, last_four, balance, currency, is_primary, is_connected, last_synced_at, created_at
FROM finance_bank_accounts
WHERE business_id = ?
ORDER BY name ASC
LIMIT 500;

-- name: CreateBankAccount :execresult
INSERT INTO finance_bank_accounts (owner_id, business_id, name, institution, account_type, last_four, balance, currency, is_primary, is_connected)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: SumBankAccountBalances :one
SELECT COALESCE(SUM(CAST(COALESCE(a.balance, 0) AS DECIMAL(30,2))), 0)
FROM finance_bank_accounts a
JOIN businesses b ON b.id = a.business_id
WHERE a.business_id = ?
  AND COALESCE(NULLIF(UPPER(TRIM(a.currency)), ''), 'USD') =
      COALESCE(NULLIF(UPPER(TRIM(b.currency)), ''), 'USD');

-- name: CountConnectedBankAccounts :one
SELECT COUNT(*)
FROM finance_bank_accounts
WHERE business_id = ? AND is_connected = TRUE;
