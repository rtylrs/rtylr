-- name: CountAppointmentsToday :one
-- cross-schema read from booking_appointments (appointments module) — count appointments in the requested business-day window.
SELECT COUNT(*) AS cnt
FROM booking_appointments
WHERE business_id = ? AND start_time >= ? AND start_time < ?;

-- name: CountUpcomingAppointments :one
-- cross-schema read from booking_appointments (appointments module) — count pending or confirmed future appointments.
SELECT COUNT(*) AS cnt
FROM booking_appointments
WHERE business_id = ? AND start_time >= ? AND status IN ('pending', 'confirmed');

-- name: CountAppointmentsThisMonth :one
-- cross-schema read from booking_appointments (appointments module) — count appointments from the requested month start.
SELECT COUNT(*) AS cnt
FROM booking_appointments
WHERE business_id = ? AND start_time >= ?;

-- name: CountCancelledAppointmentsThisMonth :one
-- cross-schema read from booking_appointments (appointments module) — count cancelled appointments from the requested month start.
SELECT COUNT(*) AS cnt
FROM booking_appointments
WHERE business_id = ? AND start_time >= ? AND status = 'cancelled';

-- name: SumMonthRevenue :one
-- cross-schema read from booking_appointments + booking_services (appointments and services modules) — sum service prices for completed or confirmed appointments.
SELECT COALESCE(SUM(bs.price), 0) AS revenue
FROM booking_appointments ba
JOIN booking_services bs ON ba.service_id = bs.id
WHERE ba.business_id = ? AND ba.start_time >= ?
  AND ba.status IN ('completed', 'confirmed');
