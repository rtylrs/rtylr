-- name: GetTaxConfig :one
SELECT id, owner_id, business_id, name, rate, type, is_default, is_active, is_inclusive, created_at
FROM finance_tax_config
WHERE business_id = ? AND is_default = TRUE
ORDER BY id ASC
LIMIT 1;

-- name: ClearDefaultTaxConfig :exec
UPDATE finance_tax_config SET is_default = FALSE WHERE business_id = ?;

-- name: UpsertTaxConfig :exec
INSERT INTO finance_tax_config (owner_id, business_id, name, rate, type, is_default, is_active, is_inclusive)
VALUES (?, ?, ?, ?, ?, TRUE, ?, ?)
ON DUPLICATE KEY UPDATE
  name = VALUES(name),
  rate = VALUES(rate),
  is_default = TRUE,
  is_active = VALUES(is_active),
  is_inclusive = VALUES(is_inclusive);

-- cross-schema read from invoices — calculate paid invoice totals for tax reporting.
-- name: SumPaidInvoicesTotal :one
SELECT COALESCE(SUM(total), 0) FROM finance_invoices WHERE business_id = ? AND status = 'paid';

-- cross-schema read from expenses — calculate the expense base for tax reporting.
-- name: SumAllExpensesAmount :one
SELECT COALESCE(SUM(amount), 0) FROM finance_expenses WHERE business_id = ? AND status = 'approved';
