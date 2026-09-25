-- name: ListBankFeedTransactions :many
SELECT id, owner_id, business_id, bank_account_id, amount_minor_units, currency, occurred_at,
       description, external_id, matched_expense_id, matched_invoice_id, created_at
FROM finance_bank_feeds
WHERE business_id = ?
ORDER BY occurred_at DESC, id DESC
LIMIT ? OFFSET ?;

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

-- cross-schema read from finance_bank_feeds — apply all imported rows on the
-- first sync, before a cursor exists.
-- name: ListBankFeedAmounts :many
SELECT amount_minor_units
FROM finance_bank_feeds
WHERE business_id = ? AND bank_account_id = ?
ORDER BY occurred_at ASC, id ASC
LIMIT 100;

-- cross-schema read from finance_bank_feeds — apply only rows imported after
-- the account's last sync cursor. created_at tracks import time; occurred_at
-- can legitimately be an older bank transaction date.
-- name: ListBankFeedAmountsSince :many
SELECT amount_minor_units
FROM finance_bank_feeds
WHERE business_id = ? AND bank_account_id = ? AND created_at > ?
ORDER BY created_at ASC, id ASC;

-- cross-schema write to finance_bank_accounts — persist the applied feed
-- balance and advance the idempotency cursor together.
-- name: UpdateSyncedBankAccount :execresult
UPDATE finance_bank_accounts
SET balance = ?, last_synced_at = CURRENT_TIMESTAMP
WHERE id = ? AND business_id = ?;
