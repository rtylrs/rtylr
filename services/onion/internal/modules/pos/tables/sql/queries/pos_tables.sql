-- name: ListTablesByOwner :many
SELECT id, owner_id, business_id, location_id, name, capacity, status, current_session_order_id, qr_code_url, created_at, updated_at
FROM pos_tables WHERE business_id = ? ORDER BY name;

-- name: CreateTable :execresult
INSERT INTO pos_tables (owner_id, business_id, location_id, name, capacity, status)
VALUES (?, ?, ?, ?, ?, 'available');

-- name: UpdateTable :exec
UPDATE pos_tables SET location_id = ?, name = ?, capacity = ?, status = ?
WHERE id = ? AND business_id = ?;

-- name: UpdateTableStatus :execresult
UPDATE pos_tables SET status = ? WHERE id = ? AND business_id = ?;

-- name: DeleteTable :execresult
DELETE FROM pos_tables WHERE id = ? AND business_id = ?;
