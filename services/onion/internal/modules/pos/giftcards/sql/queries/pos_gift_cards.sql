-- name: ListGiftCardsByOwner :many
SELECT id, owner_id, business_id, customer_id, code, initial_amount, current_balance,
       customer_name, customer_email, is_active, expires_at, created_at, updated_at
FROM pos_gift_cards WHERE business_id = ? ORDER BY created_at DESC
LIMIT 200;

-- name: CreateGiftCard :execresult
INSERT INTO pos_gift_cards (owner_id, business_id, customer_id, code, initial_amount, current_balance, customer_name, customer_email, expires_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: LinkCustomerToGiftCard :exec
UPDATE pos_gift_cards
SET customer_id = ?
WHERE business_id = ? AND id = ?;

-- name: GetGiftCardByCode :one
SELECT id, owner_id, business_id, customer_id, code, initial_amount, current_balance,
       customer_name, customer_email, is_active, expires_at, created_at, updated_at
FROM pos_gift_cards WHERE business_id = ? AND code = ?;

-- name: GetGiftCardByCodeForUpdate :one
SELECT id, owner_id, business_id, customer_id, code, initial_amount, current_balance,
       customer_name, customer_email, is_active, expires_at, created_at, updated_at
FROM pos_gift_cards WHERE business_id = ? AND code = ? FOR UPDATE;

-- name: UpdateGiftCardBalance :exec
UPDATE pos_gift_cards SET current_balance = ? WHERE id = ? AND business_id = ?;
