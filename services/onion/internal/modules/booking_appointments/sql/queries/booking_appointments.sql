-- name: GetAppointment :one
SELECT id, owner_id, business_id, service_id, resource_id, staff_id,
       customer_name, customer_email, customer_phone, start_time, end_time,
       status, notes, created_at, updated_at
FROM booking_appointments
WHERE business_id = ? AND id = ?;

-- name: ListAppointments :many
SELECT id, owner_id, business_id, service_id, resource_id, staff_id, customer_name, customer_email, customer_phone, start_time, end_time, status, notes, created_at, updated_at
FROM booking_appointments WHERE business_id = ? ORDER BY start_time DESC;

-- name: ListAppointmentsByDateRange :many
SELECT id, owner_id, business_id, service_id, resource_id, staff_id, customer_name, customer_email, customer_phone, start_time, end_time, status, notes, created_at, updated_at
FROM booking_appointments WHERE business_id = ? AND start_time >= ? AND start_time < ? ORDER BY start_time ASC;

-- name: CreateAppointment :execresult
INSERT INTO booking_appointments (owner_id, business_id, service_id, resource_id, staff_id, customer_name, customer_email, customer_phone, start_time, end_time, status, notes)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: UpdateAppointment :exec
UPDATE booking_appointments SET service_id = ?, resource_id = ?, staff_id = ?, customer_name = ?, customer_email = ?, customer_phone = ?, start_time = ?, end_time = ?, status = ?, notes = ?
WHERE business_id = ? AND id = ?;

-- name: DeleteAppointment :exec
DELETE FROM booking_appointments WHERE business_id = ? AND id = ?;
