-- name: ListLocationsByOwner :many
SELECT id, owner_id, name, address, phone, is_active, created_at
FROM pos_locations WHERE owner_id = ? ORDER BY name;

-- name: CreateLocation :execresult
INSERT INTO pos_locations (owner_id, name, address, phone, is_active)
VALUES (?, ?, ?, ?, ?);

-- name: UpdateLocation :execresult
UPDATE pos_locations SET name = ?, address = ?, phone = ?, is_active = COALESCE(?, is_active)
WHERE id = ? AND owner_id = ?;
