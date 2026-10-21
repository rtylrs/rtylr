-- name: GetLoyaltyConfig :one
SELECT id, business_id, is_enabled, points_per_dollar, redemption_rate, min_redeem_points,
       created_at, updated_at
FROM pos_loyalty_config WHERE business_id = ?;

-- name: UpsertLoyaltyConfig :execresult
INSERT INTO pos_loyalty_config (business_id, is_enabled, points_per_dollar, redemption_rate, min_redeem_points)
VALUES (?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
    is_enabled = VALUES(is_enabled),
    points_per_dollar = VALUES(points_per_dollar),
    redemption_rate = VALUES(redemption_rate),
    min_redeem_points = VALUES(min_redeem_points);

-- name: ListLoyaltyTransactions :many
SELECT id, business_id, customer_id, order_id, type, points, balance_after, note, created_at
FROM pos_loyalty_transactions
WHERE business_id = ? AND customer_id = ?
ORDER BY created_at DESC LIMIT 100;

-- name: CreateLoyaltyTransaction :execresult
INSERT INTO pos_loyalty_transactions (business_id, customer_id, order_id, type, points, balance_after, note)
VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: GetLoyaltyBalance :one
SELECT COALESCE(
  (SELECT balance_after FROM pos_loyalty_transactions
   WHERE business_id = ? AND customer_id = ?
   ORDER BY id DESC LIMIT 1),
  0
) AS balance;
