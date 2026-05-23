-- name: ListQuotes :many
SELECT id, owner_id, customer_name, customer_email, status,
       items, subtotal, tax, discount, total, valid_until, notes, created_at, updated_at
FROM finance_quotes
WHERE owner_id = ?
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: CreateQuote :execresult
INSERT INTO finance_quotes
(owner_id, customer_name, customer_email, status, items, subtotal, tax, discount, total, valid_until, notes)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: UpdateQuote :exec
UPDATE finance_quotes
SET customer_name = ?, customer_email = ?, status = ?, items = ?,
    subtotal = ?, tax = ?, discount = ?, total = ?, valid_until = ?, notes = ?
WHERE id = ? AND owner_id = ?;

-- name: DeleteQuote :exec
DELETE FROM finance_quotes WHERE id = ? AND owner_id = ?;

-- name: GetQuote :one
SELECT id, owner_id, customer_name, customer_email, status,
       items, subtotal, tax, discount, total, valid_until, notes, created_at, updated_at
FROM finance_quotes
WHERE id = ? AND owner_id = ?;

-- name: UpdateQuoteStatus :execresult
UPDATE finance_quotes
SET status = ?
WHERE id = ? AND owner_id = ?;
