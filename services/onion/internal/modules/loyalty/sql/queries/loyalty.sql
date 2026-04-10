-- name: CountTransactions :one
SELECT COUNT(*) FROM crm_loyalty_transactions WHERE owner_id = ?;

-- name: CountTransactionsByCustomer :one
SELECT COUNT(*) FROM crm_loyalty_transactions WHERE owner_id = ? AND customer_id = ?;

-- name: ListTransactions :many
SELECT id, owner_id, customer_id, points, type, description, reference_id, reference_type, created_at
FROM crm_loyalty_transactions WHERE owner_id = ?
ORDER BY created_at DESC LIMIT ? OFFSET ?;

-- name: ListTransactionsByCustomer :many
SELECT id, owner_id, customer_id, points, type, description, reference_id, reference_type, created_at
FROM crm_loyalty_transactions WHERE owner_id = ? AND customer_id = ?
ORDER BY created_at DESC LIMIT ? OFFSET ?;

-- name: CreateTransaction :exec
INSERT INTO crm_loyalty_transactions (owner_id, customer_id, points, type, description, reference_id, reference_type)
VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: CreateRedeemTransaction :exec
INSERT INTO crm_loyalty_transactions (owner_id, customer_id, points, type, description, reference_id, reference_type)
VALUES (?, ?, ?, 'redeem', ?, ?, ?);

-- name: GetConfig :one
SELECT id, owner_id, points_per_currency_unit, currency_per_point, is_enabled, updated_at
FROM crm_loyalty_config WHERE owner_id = ?;

-- name: UpsertConfig :exec
INSERT INTO crm_loyalty_config (owner_id, points_per_currency_unit, currency_per_point, is_enabled)
VALUES (?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
  points_per_currency_unit = VALUES(points_per_currency_unit),
  currency_per_point = VALUES(currency_per_point),
  is_enabled = VALUES(is_enabled);
