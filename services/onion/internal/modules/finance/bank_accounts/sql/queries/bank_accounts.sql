-- name: CountConnectedBankAccounts :one
SELECT COUNT(*)
FROM finance_bank_accounts
WHERE business_id = ? AND is_connected = TRUE;
