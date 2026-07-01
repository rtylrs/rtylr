-- name: FetchCalendarAppointments :many
SELECT id, service_id, customer_name, start_time, end_time, status, business_id
FROM booking_appointments
WHERE business_id = ?
  AND start_time >= ?
  AND start_time < ?
ORDER BY start_time ASC;
