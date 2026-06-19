-- name: GetTaxConfig :one
SELECT id, owner_id, business_id, tax_name, tax_rate, is_inclusive, is_active, created_at, updated_at
FROM finance_tax_config
WHERE business_id = ?;

-- name: UpsertTaxConfig :exec
INSERT INTO finance_tax_config (owner_id, business_id, tax_name, tax_rate, is_inclusive, is_active)
VALUES (?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
  business_id = VALUES(business_id),
  tax_name = VALUES(tax_name),
  tax_rate = VALUES(tax_rate),
  is_inclusive = VALUES(is_inclusive),
  is_active = VALUES(is_active);

-- name: SumPaidInvoicesTotal :one
SELECT COALESCE(SUM(total), 0) FROM finance_invoices WHERE business_id = ? AND status = 'paid';

-- name: SumAllExpensesAmount :one
SELECT COALESCE(SUM(amount), 0) FROM finance_expenses WHERE business_id = ?;
