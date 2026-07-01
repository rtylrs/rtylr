-- name: CountAppointmentsToday :one
SELECT COUNT(*) AS cnt
FROM booking_appointments
WHERE business_id = ? AND start_time >= ? AND start_time < ?;

-- name: CountUpcomingAppointments :one
SELECT COUNT(*) AS cnt
FROM booking_appointments
WHERE business_id = ? AND start_time >= ? AND status IN ('pending', 'confirmed');

-- name: CountAppointmentsThisMonth :one
SELECT COUNT(*) AS cnt
FROM booking_appointments
WHERE business_id = ? AND start_time >= ?;

-- name: CountCancelledAppointmentsThisMonth :one
SELECT COUNT(*) AS cnt
FROM booking_appointments
WHERE business_id = ? AND start_time >= ? AND status = 'cancelled';

-- name: SumMonthRevenue :one
SELECT COALESCE(SUM(bs.price), 0) AS revenue
FROM booking_appointments ba
JOIN booking_services bs ON ba.service_id = bs.id
WHERE ba.business_id = ? AND ba.start_time >= ?
  AND ba.status IN ('completed', 'confirmed');
