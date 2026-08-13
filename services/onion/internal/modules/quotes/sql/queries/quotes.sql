-- name: ListQuotes :many
SELECT id, owner_id, business_id, customer_name, customer_email, status,
       items, subtotal, tax, discount, total, valid_until, notes, created_at, updated_at
FROM finance_quotes
WHERE business_id = ?
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: CreateQuote :execresult
INSERT INTO finance_quotes
(owner_id, business_id, customer_name, customer_email, status, items, subtotal, tax, discount, total, valid_until, notes)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: UpdateQuote :exec
UPDATE finance_quotes
SET customer_name = ?, customer_email = ?, status = ?, items = ?,
    subtotal = ?, tax = ?, discount = ?, total = ?, valid_until = ?, notes = ?
WHERE id = ? AND business_id = ?;

-- name: DeleteQuote :exec
DELETE FROM finance_quotes WHERE id = ? AND business_id = ?;

-- name: GetQuote :one
SELECT id, owner_id, business_id, customer_name, customer_email, status,
       items, subtotal, tax, discount, total, valid_until, notes, created_at, updated_at
FROM finance_quotes
WHERE id = ? AND business_id = ?;

-- name: UpdateQuoteStatus :execresult
UPDATE finance_quotes
SET status = ?
WHERE id = ? AND business_id = ?;

-- name: CountQuotesByStatus :many
SELECT status, COUNT(*) AS cnt
FROM finance_quotes
WHERE business_id = ?
GROUP BY status;

-- name: InsertInvoiceFromQuote :execresult
INSERT INTO finance_invoices
(owner_id, business_id, invoice_number, client_name, client_email, status,
 issue_date, due_date, subtotal, tax_amount, discount_amount, total,
 amount_due, notes)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: InsertInvoiceItemFromQuote :exec
INSERT INTO finance_invoice_items
(invoice_id, description, quantity, unit, unit_price, tax_rate, discount_pct, total, sort_order)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: CountInvoiceSequenceForQuote :one
SELECT COUNT(*)
FROM finance_invoices
WHERE business_id = ? AND YEAR(created_at) = ?;
