-- name: ListBookingServices :many
SELECT id, owner_id, business_id, name, description, duration_minutes, price, color, is_active, created_at, updated_at
FROM booking_services WHERE business_id = ? ORDER BY name ASC LIMIT 100;

-- name: CreateBookingService :execresult
INSERT INTO booking_services (owner_id, business_id, name, description, duration_minutes, price, color, is_active) VALUES (?, ?, ?, ?, ?, ?, ?, ?);

-- name: UpdateBookingService :exec
UPDATE booking_services SET name = ?, description = ?, duration_minutes = ?, price = ?, color = ?, is_active = ? WHERE business_id = ? AND id = ?;

-- name: DeleteBookingService :exec
DELETE FROM booking_services WHERE business_id = ? AND id = ?;
