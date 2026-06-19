-- name: CreateCheckout :execresult
INSERT INTO store_checkouts (owner_id, business_id, session_id, customer_email, items, subtotal, tax, total)
VALUES (?, ?, ?, ?, ?, ?, ?, ?);

-- name: CompleteCheckout :exec
UPDATE store_checkouts SET status = 'completed' WHERE business_id = ? AND id = ?;
