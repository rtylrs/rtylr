-- name: ListWaitlist :many
SELECT id, owner_id, business_id, service_id, customer_name, customer_email, customer_phone, preferred_date, preferred_time, status, created_at
FROM booking_waitlist WHERE business_id = ? ORDER BY created_at DESC LIMIT 100;

-- name: GetBookingServiceForOwner :one
SELECT id, owner_id, business_id FROM booking_services WHERE id = ? AND business_id = ?;

-- name: GetWaitlistEntry :one
SELECT id FROM booking_waitlist WHERE id = ? AND business_id = ?;

-- name: CreateWaitlistEntry :execresult
INSERT INTO booking_waitlist (owner_id, business_id, service_id, customer_name, customer_email, customer_phone, preferred_date, preferred_time, status)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: UpdateWaitlistEntry :execresult
UPDATE booking_waitlist SET service_id = ?, customer_name = ?, customer_email = ?, customer_phone = ?, preferred_date = ?, preferred_time = ?, status = ?
WHERE business_id = ? AND id = ?;
