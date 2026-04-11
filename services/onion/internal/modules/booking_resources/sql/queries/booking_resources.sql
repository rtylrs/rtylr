-- name: ListResources :many
SELECT id, owner_id, name, type, capacity, is_active, created_at
FROM booking_resources WHERE owner_id = ? ORDER BY name ASC;

-- name: CreateResource :execresult
INSERT INTO booking_resources (owner_id, name, type, capacity, is_active) VALUES (?, ?, ?, ?, ?);

-- name: UpdateResource :exec
UPDATE booking_resources SET name = ?, type = ?, capacity = ?, is_active = ? WHERE owner_id = ? AND id = ?;

-- name: DeleteResource :exec
DELETE FROM booking_resources WHERE owner_id = ? AND id = ?;
