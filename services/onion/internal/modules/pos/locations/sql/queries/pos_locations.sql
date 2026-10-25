-- name: ListLocationsByOwner :many
SELECT id, owner_id, business_id, name, address, phone, is_active, created_at
FROM pos_locations WHERE business_id = ? ORDER BY name;

-- name: CreateLocation :execresult
INSERT INTO pos_locations (owner_id, business_id, name, address, phone, is_active)
VALUES (?, ?, ?, ?, ?, ?);

-- name: EnsureLocationLock :exec
INSERT INTO pos_location_creation_locks (business_id) VALUES (?) ON DUPLICATE KEY UPDATE business_id = business_id;

-- name: LockLocationCreation :one
SELECT business_id FROM pos_location_creation_locks WHERE business_id = ? FOR UPDATE;

-- name: CountLocationsByBusiness :one
SELECT COUNT(*) FROM pos_locations WHERE business_id = ?;

-- name: UpdateLocation :execresult
UPDATE pos_locations SET name = ?, address = ?, phone = ?, is_active = COALESCE(?, is_active)
WHERE id = ? AND business_id = ?;

-- name: StockBranchExists :one
SELECT COUNT(*) > 0 FROM pos_locations WHERE business_id = ? AND id = ?;
