-- The generated list contract has no pagination fields; keep the endpoint
-- bounded until the corresponding voxie contract can expose a page cursor.
-- name: ListRecurringInvoices :many
SELECT id, owner_id, business_id, template_invoice_id, frequency, next_invoice_date, end_date, is_active, created_at
FROM finance_recurring_invoices
WHERE business_id = ?
ORDER BY next_invoice_date ASC
LIMIT 500;

-- name: CreateRecurringInvoice :execresult
INSERT INTO finance_recurring_invoices
(owner_id, business_id, template_invoice_id, frequency, next_invoice_date, end_date, is_active)
VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: UpdateRecurringInvoice :exec
UPDATE finance_recurring_invoices
SET template_invoice_id = ?, frequency = ?, next_invoice_date = ?, end_date = ?, is_active = ?
WHERE id = ? AND business_id = ?;

-- name: SetRecurringInvoiceActive :execresult
UPDATE finance_recurring_invoices
SET is_active = ?
WHERE id = ? AND business_id = ?;
