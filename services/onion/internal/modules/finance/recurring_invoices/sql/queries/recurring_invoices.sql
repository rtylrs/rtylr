-- name: ListRecurringInvoicesPage :many
SELECT id, owner_id, business_id, template_invoice_id, frequency, next_invoice_date, end_date, is_active, created_at
FROM finance_recurring_invoices
WHERE business_id = ?
ORDER BY next_invoice_date ASC, id ASC
LIMIT ? OFFSET ?;

-- name: GetRecurringInvoice :one
SELECT id, owner_id, business_id, template_invoice_id, frequency, next_invoice_date, end_date, is_active, created_at
FROM finance_recurring_invoices
WHERE id = ? AND business_id = ?;

-- name: CreateRecurringInvoice :execresult
INSERT INTO finance_recurring_invoices
(owner_id, business_id, template_invoice_id, frequency, next_invoice_date, end_date, is_active)
VALUES (?, ?, ?, ?, ?, ?, ?);

-- Partial update: an absent optional field keeps its stored value. set_end_date
-- is 1 when the caller supplied end_date (an empty value clears the bound) and
-- 0 when the field was omitted.
-- name: UpdateRecurringInvoice :execresult
UPDATE finance_recurring_invoices
SET template_invoice_id = COALESCE(sqlc.narg(template_invoice_id), template_invoice_id),
    frequency = ?,
    next_invoice_date = ?,
    end_date = CASE WHEN CAST(sqlc.arg(set_end_date) AS UNSIGNED) = 1 THEN sqlc.narg(end_date) ELSE end_date END,
    is_active = COALESCE(sqlc.narg(is_active), is_active)
WHERE id = ? AND business_id = ?;

-- name: RecurringInvoiceExists :one
SELECT EXISTS (
    SELECT 1 FROM finance_recurring_invoices WHERE id = ? AND business_id = ?
);

-- name: SetRecurringInvoiceActive :execresult
UPDATE finance_recurring_invoices
SET is_active = ?
WHERE id = ? AND business_id = ?;
