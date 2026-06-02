-- name: ListWaitlist :many
SELECT id, owner_id, service_id, customer_name, customer_email, customer_phone, preferred_date, preferred_time, status, created_at
FROM booking_waitlist WHERE owner_id = ? ORDER BY created_at DESC;

-- name: GetBookingServiceForOwner :one
SELECT id, owner_id FROM booking_services WHERE id = ? AND owner_id = ?;

-- name: CreateWaitlistEntry :execresult
INSERT INTO booking_waitlist (owner_id, service_id, customer_name, customer_email, customer_phone, preferred_date, preferred_time, status)
VALUES (?, ?, ?, ?, ?, ?, ?, ?);

-- name: UpdateWaitlistEntry :exec
UPDATE booking_waitlist SET service_id = ?, customer_name = ?, customer_email = ?, customer_phone = ?, preferred_date = ?, preferred_time = ?, status = ?
WHERE owner_id = ? AND id = ?;
