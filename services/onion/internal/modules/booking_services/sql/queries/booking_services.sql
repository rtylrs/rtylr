-- name: ListBookingServices :many
SELECT id, owner_id, name, description, duration_minutes, price, color, is_active, created_at, updated_at
FROM booking_services WHERE owner_id = ? ORDER BY name ASC;

-- name: CreateBookingService :execresult
INSERT INTO booking_services (owner_id, name, description, duration_minutes, price, color, is_active) VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: UpdateBookingService :exec
UPDATE booking_services SET name = ?, description = ?, duration_minutes = ?, price = ?, color = ?, is_active = ? WHERE owner_id = ? AND id = ?;

-- name: DeleteBookingService :exec
DELETE FROM booking_services WHERE owner_id = ? AND id = ?;
