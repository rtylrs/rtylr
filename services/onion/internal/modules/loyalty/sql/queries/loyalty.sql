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

-- Aggregated member balance over the transaction ledger; scoped to one
-- customer per call so the result set is bounded.
-- name: GetCustomerPointsBalance :one
SELECT CAST(COALESCE(SUM(points), 0) AS SIGNED) AS balance
FROM crm_loyalty_transactions WHERE business_id = ? AND customer_id = ?;

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

-- name: UpdateReward :execresult
UPDATE crm_loyalty_rewards
SET name = ?, description = ?, category = ?, points_cost = ?, stock = ?, icon = ?, active = ?
WHERE id = ? AND business_id = ?;

-- cross-schema read from crm_customers — loyalty members projection
-- name: ListLoyaltyMembers :many
SELECT
    id,
    owner_id,
    CONCAT_WS(' ', first_name, COALESCE(last_name, '')) AS name,
    COALESCE(loyalty_points, 0) AS loyalty_points,
    COALESCE(loyalty_tier, 'bronze') AS loyalty_tier,
    COALESCE(visit_count, 0) AS visit_count,
    created_at,
    COALESCE(last_visit_at, created_at) AS last_active_at
FROM crm_customers
WHERE business_id = ? AND is_active = TRUE
  AND (loyalty_points > 0 OR visit_count > 0)
ORDER BY loyalty_points DESC, visit_count DESC
LIMIT ? OFFSET ?;

-- name: ListLoyaltyRewards :many
SELECT id, owner_id, name, COALESCE(description, '') AS description,
       COALESCE(category, 'general') AS category, points_cost, stock,
       COALESCE(icon, '') AS icon, COALESCE(active, TRUE) AS active,
       created_at, updated_at
FROM crm_loyalty_rewards
WHERE business_id = ?
ORDER BY points_cost ASC, id ASC
LIMIT 200;

-- name: ListActiveLoyaltyRewards :many
SELECT id, owner_id, name, COALESCE(description, '') AS description,
       COALESCE(category, 'general') AS category, points_cost, stock,
       COALESCE(icon, '') AS icon, COALESCE(active, TRUE) AS active,
       created_at, updated_at
FROM crm_loyalty_rewards
WHERE business_id = ? AND active = TRUE
ORDER BY points_cost ASC, id ASC
LIMIT 200;

