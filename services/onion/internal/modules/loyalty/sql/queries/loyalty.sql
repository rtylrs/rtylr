-- name: CountTransactions :one
SELECT COUNT(*) FROM crm_loyalty_transactions WHERE business_id = ?;

-- name: CountTransactionsByCustomer :one
SELECT COUNT(*) FROM crm_loyalty_transactions WHERE business_id = ? AND customer_id = ?;

-- name: ListTransactions :many
SELECT id, owner_id, business_id, customer_id, points, type, description, reference_id, reference_type, created_at
FROM crm_loyalty_transactions WHERE business_id = ?
ORDER BY created_at DESC LIMIT ? OFFSET ?;

-- name: ListTransactionsByCustomer :many
SELECT id, owner_id, business_id, customer_id, points, type, description, reference_id, reference_type, created_at
FROM crm_loyalty_transactions WHERE business_id = ? AND customer_id = ?
ORDER BY created_at DESC LIMIT ? OFFSET ?;

-- name: CreateTransaction :exec
INSERT INTO crm_loyalty_transactions (owner_id, business_id, customer_id, points, type, description, reference_id, reference_type)
VALUES (?, ?, ?, ?, ?, ?, ?, ?);

-- name: CreateRedeemTransaction :exec
INSERT INTO crm_loyalty_transactions (owner_id, business_id, customer_id, points, type, description, reference_id, reference_type)
VALUES (?, ?, ?, ?, 'redeem', ?, ?, ?);

-- name: GetConfig :one
SELECT id, owner_id, business_id, points_per_currency_unit, currency_per_point, min_redemption_points, is_enabled, updated_at
FROM crm_loyalty_config WHERE business_id = ?;

-- name: UpsertConfig :exec
INSERT INTO crm_loyalty_config (owner_id, business_id, points_per_currency_unit, currency_per_point, min_redemption_points, is_enabled)
VALUES (?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
  points_per_currency_unit = VALUES(points_per_currency_unit),
  currency_per_point = VALUES(currency_per_point),
  min_redemption_points = VALUES(min_redemption_points),
  is_enabled = VALUES(is_enabled);

-- name: CreateReward :execresult
INSERT INTO crm_loyalty_rewards (owner_id, business_id, name, description, category, points_cost, stock, icon, active)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: UpdateReward :exec
UPDATE crm_loyalty_rewards
SET name = ?, description = ?, category = ?, points_cost = ?, stock = ?, icon = ?, active = ?
WHERE id = ? AND business_id = ?;
