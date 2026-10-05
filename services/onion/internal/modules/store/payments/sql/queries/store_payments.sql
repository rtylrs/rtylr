-- name: ListPaymentMethods :many
SELECT id, owner_id, business_id, provider, COALESCE(config, '{}') AS config, is_active, created_at
FROM store_payment_methods WHERE business_id = ? ORDER BY provider ASC
LIMIT 200;

-- name: CreatePaymentMethod :execresult
INSERT INTO store_payment_methods (owner_id, business_id, provider, config, is_active) VALUES (?, ?, ?, ?, ?);

-- name: UpdatePaymentMethod :exec
UPDATE store_payment_methods SET provider = ?, config = ?, is_active = ? WHERE business_id = ? AND id = ?;
