-- name: ListLocationsByOwner :many
SELECT id, owner_id, business_id, name, address, phone, is_active, created_at
FROM pos_locations WHERE business_id = ? ORDER BY name;

-- name: CreateLocation :execresult
INSERT INTO pos_locations (owner_id, business_id, name, address, phone, is_active)
VALUES (?, ?, ?, ?, ?, ?);

-- name: LockLocationPlanLimit :one
SELECT p.max_locations
FROM businesses b
JOIN subscriptions s ON s.owner_id = b.owner_id
JOIN plans p ON p.id = s.plan_id
WHERE b.id = ?
FOR UPDATE;

-- name: CountLocationsByBusiness :one
SELECT COUNT(*) FROM pos_locations WHERE business_id = ?;

-- name: UpdateLocation :execresult
UPDATE pos_locations SET name = ?, address = ?, phone = ?, is_active = COALESCE(?, is_active)
WHERE id = ? AND business_id = ?;
