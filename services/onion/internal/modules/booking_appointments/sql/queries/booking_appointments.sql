-- name: GetAppointment :one
SELECT id, owner_id, business_id, service_id, resource_id, staff_id,
       customer_name, customer_email, customer_phone, start_time, end_time,
       status, notes, created_at, updated_at
FROM booking_appointments
WHERE business_id = ? AND id = ?;

-- name: ListAppointments :many
SELECT id, owner_id, business_id, service_id, resource_id, staff_id, customer_name, customer_email, customer_phone, start_time, end_time, status, notes, created_at, updated_at
FROM booking_appointments WHERE business_id = ? ORDER BY start_time DESC, id DESC LIMIT 500;

-- name: ListAppointmentsByDateRange :many
SELECT id, owner_id, business_id, service_id, resource_id, staff_id, customer_name, customer_email, customer_phone, start_time, end_time, status, notes, created_at, updated_at
FROM booking_appointments WHERE business_id = ? AND start_time >= ? AND start_time < ? ORDER BY start_time ASC, id ASC LIMIT 500;

-- Lock conflicting appointments in the same transaction as create/reschedule.
-- The resource and staff predicates protect either scheduling dimension.
-- name: LockAppointmentOverlaps :many
SELECT id
FROM booking_appointments
WHERE business_id = ?
  AND status NOT IN ('cancelled', 'no_show')
  AND start_time < ?
  AND end_time > ?
  AND (? = 0 OR id <> ?)
  AND ((? > 0 AND resource_id = ?) OR (? > 0 AND staff_id = ?))
LIMIT 1
FOR UPDATE;

-- name: GetAppointmentStatusForUpdate :one
SELECT status FROM booking_appointments WHERE business_id = ? AND id = ? FOR UPDATE;

-- Serialize booking mutations per tenant before the overlap read. This is the
-- race-safe backstop when no conflicting appointment row exists yet.
-- name: LockBusinessForBooking :one
SELECT id FROM businesses WHERE id = ? FOR UPDATE;

-- name: GetBookingServiceForAppointment :one
SELECT id FROM booking_services WHERE id = ? AND business_id = ? AND is_active = TRUE;

-- name: GetBookingResourceForAppointment :one
SELECT id FROM booking_resources WHERE id = ? AND business_id = ? AND is_active = TRUE;

-- name: CreateAppointment :execresult
INSERT INTO booking_appointments (owner_id, business_id, service_id, resource_id, staff_id, customer_name, customer_email, customer_phone, start_time, end_time, status, notes)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: UpdateAppointment :execresult
UPDATE booking_appointments SET service_id = ?, resource_id = ?, staff_id = ?, customer_name = ?, customer_email = ?, customer_phone = ?, start_time = ?, end_time = ?, status = ?, notes = ?
WHERE business_id = ? AND id = ?;

-- name: DeleteAppointment :execresult
DELETE FROM booking_appointments WHERE business_id = ? AND id = ?;
