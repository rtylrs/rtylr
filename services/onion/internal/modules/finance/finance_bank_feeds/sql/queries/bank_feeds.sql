-- name: CountBankFeedTransactions :one
SELECT COUNT(*) FROM finance_bank_feeds WHERE business_id = ?;

-- name: SumBankFeedAmountByAccount :one
SELECT COALESCE(SUM(amount_minor_units), 0)
FROM finance_bank_feeds
WHERE business_id = ? AND bank_account_id = ?;

-- name: CountUnmatchedBankFeedTransactions :one
SELECT COUNT(*)
FROM finance_bank_feeds
WHERE business_id = ?
  AND matched_expense_id IS NULL
  AND matched_invoice_id IS NULL;

-- cross-schema read from finance_bank_accounts — synchronize imported feed rows
-- into connected account balances.
-- name: ListSyncBankAccounts :many
SELECT id, balance, last_synced_at
FROM finance_bank_accounts
WHERE business_id = ? AND is_connected = TRUE
FOR UPDATE;

-- cross-schema write to finance_bank_accounts — persist the applied feed
-- balance and advance the idempotency cursor together.
-- name: UpdateSyncedBankAccount :execresult
UPDATE finance_bank_accounts
SET balance = ?, last_synced_at = CURRENT_TIMESTAMP
WHERE id = ? AND business_id = ?;
