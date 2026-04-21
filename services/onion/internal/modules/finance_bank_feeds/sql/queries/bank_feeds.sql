-- name: ListBankFeedTransactions :many
SELECT id, owner_id, bank_account_id, amount_minor_units, currency, occurred_at,
       description, external_id, matched_expense_id, matched_invoice_id, created_at
FROM finance_bank_feeds
WHERE owner_id = ?
ORDER BY occurred_at DESC, id DESC
LIMIT ? OFFSET ?;

-- name: CountBankFeedTransactions :one
SELECT COUNT(*) FROM finance_bank_feeds WHERE owner_id = ?;
