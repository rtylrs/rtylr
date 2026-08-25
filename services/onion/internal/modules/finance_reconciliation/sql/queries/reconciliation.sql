-- name: ListUnmatchedBankFeeds :many
SELECT bf.id, bf.owner_id, bf.business_id, bf.bank_account_id, bf.amount_minor_units, bf.currency,
       bf.occurred_at, bf.description, bf.external_id
FROM finance_bank_feeds bf
WHERE bf.business_id = ?
  AND bf.matched_expense_id IS NULL
  AND bf.matched_invoice_id IS NULL
  AND NOT EXISTS (
      SELECT 1 FROM finance_reconciliation_matches m FORCE INDEX (idx_finance_reconciliation_business_bank)
      WHERE m.business_id = bf.business_id
        AND m.bank_feed_id = bf.id
  )
ORDER BY bf.occurred_at DESC, bf.id DESC
LIMIT 201;

-- name: CreateReconciliationMatch :execresult
INSERT INTO finance_reconciliation_matches (
    owner_id, business_id, bank_feed_id, target_type, target_id, matched_by
) VALUES (?, ?, ?, ?, ?, ?);

-- name: GetBankFeedForOwner :one
SELECT id FROM finance_bank_feeds
WHERE id = ? AND business_id = ?;

-- name: LockBankFeedForBusiness :one
SELECT id FROM finance_bank_feeds
WHERE id = ? AND business_id = ?
FOR UPDATE;

-- name: GetMatchForBankFeed :one
SELECT id FROM finance_reconciliation_matches
WHERE business_id = ? AND bank_feed_id = ?
ORDER BY id ASC
LIMIT 1;

-- name: GetInvoiceForBusiness :one
SELECT id FROM finance_invoices
WHERE id = ? AND business_id = ?;

-- name: GetExpenseForBusiness :one
SELECT id FROM finance_expenses
WHERE id = ? AND business_id = ?;

-- name: ListUnmatchedInvoices :many
SELECT i.id, i.invoice_number, i.client_name, i.total, i.currency, i.issue_date, i.due_date
FROM finance_invoices i
WHERE i.business_id = ?
  AND i.status IN ('draft','sent','viewed','partial','overdue')
  AND NOT EXISTS (
      SELECT 1 FROM finance_reconciliation_matches m FORCE INDEX (idx_finance_reconciliation_business_target)
      WHERE m.business_id = i.business_id
        AND m.target_type = 'invoice'
        AND m.target_id = i.id
  )
ORDER BY i.due_date ASC, i.id ASC
LIMIT 201;

-- name: ListUnmatchedExpenses :many
SELECT e.id, e.description, COALESCE(e.vendor, '') AS vendor, e.amount, e.currency, e.expense_date
FROM finance_expenses e
WHERE e.business_id = ?
  AND NOT EXISTS (
      SELECT 1 FROM finance_reconciliation_matches m FORCE INDEX (idx_finance_reconciliation_business_target)
      WHERE m.business_id = e.business_id
        AND m.target_type = 'expense'
        AND m.target_id = e.id
  )
ORDER BY e.expense_date ASC, e.id ASC
LIMIT 201;

-- name: CountMatches :one
SELECT COUNT(*)
FROM finance_reconciliation_matches
WHERE business_id = ?;

-- name: CountMatchesByTargetType :many
SELECT target_type, COUNT(*) AS cnt
FROM finance_reconciliation_matches
WHERE business_id = ?
  AND created_at >= DATE_SUB(UTC_TIMESTAMP(), INTERVAL 12 MONTH)
GROUP BY target_type;
