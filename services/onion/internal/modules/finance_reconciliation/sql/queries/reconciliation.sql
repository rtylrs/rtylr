-- name: ListUnmatchedBankFeeds :many
SELECT bf.id, bf.owner_id, bf.bank_account_id, bf.amount_minor_units, bf.currency,
       bf.occurred_at, bf.description, bf.external_id
FROM finance_bank_feeds bf
WHERE bf.owner_id = ?
  AND bf.matched_expense_id IS NULL
  AND bf.matched_invoice_id IS NULL
  AND NOT EXISTS (
      SELECT 1 FROM finance_reconciliation_matches m
      WHERE m.owner_id = bf.owner_id
        AND m.bank_feed_id = bf.id
  )
ORDER BY bf.occurred_at DESC, bf.id DESC
LIMIT 200;

-- name: CreateReconciliationMatch :execresult
INSERT INTO finance_reconciliation_matches (
    owner_id, bank_feed_id, target_type, target_id, matched_by
) VALUES (?, ?, ?, ?, ?);

-- name: GetBankFeedForOwner :one
SELECT id FROM finance_bank_feeds
WHERE id = ? AND owner_id = ?;
