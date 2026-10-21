-- name: ListSessionsByOwner :many
SELECT id, owner_id, business_id, location_id, staff_id, status, opening_cash, closing_cash, expected_cash,
       cash_difference, total_sales, total_orders, total_cash_sales, total_card_sales,
       total_online_sales, total_refunds, opened_at, closed_at, notes
FROM pos_sessions WHERE business_id = ? ORDER BY opened_at DESC LIMIT 100;

-- name: GetSessionByID :one
SELECT id, owner_id, business_id, location_id, staff_id, status, opening_cash, closing_cash, expected_cash,
       cash_difference, total_sales, total_orders, total_cash_sales, total_card_sales,
       total_online_sales, total_refunds, opened_at, closed_at, notes
FROM pos_sessions WHERE id = ? AND business_id = ?;

-- name: OpenSession :execresult
INSERT INTO pos_sessions (owner_id, business_id, location_id, staff_id, opening_cash, status, opened_at)
VALUES (?, ?, ?, ?, ?, 'open', NOW());

-- name: CloseSession :execresult
UPDATE pos_sessions SET closing_cash = ?, status = 'closed', closed_at = NOW()
WHERE id = ? AND business_id = ? AND status = 'open';
