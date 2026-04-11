-- name: ListPaymentMethods :many
SELECT id, owner_id, provider, config, is_active, created_at
FROM store_payment_methods WHERE owner_id = ? ORDER BY provider ASC;

-- name: CreatePaymentMethod :execresult
INSERT INTO store_payment_methods (owner_id, provider, config, is_active) VALUES (?, ?, ?, ?);

-- name: UpdatePaymentMethod :exec
UPDATE store_payment_methods SET provider = ?, config = ?, is_active = ? WHERE owner_id = ? AND id = ?;
