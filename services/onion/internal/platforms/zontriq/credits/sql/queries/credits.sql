-- name: GetCreditsByUserID :one
SELECT
    c.id,
    c.zontriq_user_id,
    c.balance,
    c.total_purchased,
    c.total_spent,
    c.status,
    c.created_at,
    c.updated_at
FROM zontriq_credits c
WHERE c.zontriq_user_id = $1
  AND c.status = 1;

-- name: CreateCredits :one
INSERT INTO zontriq_credits (zontriq_user_id)
VALUES ($1)
RETURNING *;

-- name: AddPurchaseCredits :one
UPDATE zontriq_credits
SET balance = balance + $2,
    total_purchased = total_purchased + $2,
    updated_at = NOW()
WHERE zontriq_user_id = $1
  AND status = 1
RETURNING *;

-- name: SpendCredits :one
UPDATE zontriq_credits
SET balance = balance - $2,
    total_spent = total_spent + $2,
    updated_at = NOW()
WHERE zontriq_user_id = $1
  AND status = 1
  AND balance >= $2
RETURNING *;

-- name: CreateCreditTransaction :one
INSERT INTO zontriq_credit_transactions (
    zontriq_user_id,
    amount,
    type,
    description,
    reference_type,
    reference_id
) VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: ListCreditTransactionsByUserID :many
SELECT
    t.id,
    t.zontriq_user_id,
    t.amount,
    t.type,
    t.description,
    t.reference_type,
    t.reference_id,
    t.status,
    t.created_at
FROM zontriq_credit_transactions t
WHERE t.zontriq_user_id = $1
  AND t.status = 1
ORDER BY t.created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountCreditTransactionsByUserID :one
SELECT COUNT(*) AS total
FROM zontriq_credit_transactions t
WHERE t.zontriq_user_id = $1
  AND t.status = 1;

-- name: GetZontriqUserIDByUserID :one
SELECT id
FROM zontriq_users
WHERE user_id = $1
  AND status = 1;
