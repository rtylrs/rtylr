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
