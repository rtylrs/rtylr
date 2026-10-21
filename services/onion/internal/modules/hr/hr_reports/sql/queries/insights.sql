-- Reads served to Insights reports, so Insights never queries HR tables.

-- name: InsightsCountActiveEmployees :one
SELECT COUNT(*) AS count FROM hr_employees WHERE business_id = ? AND status = 'active';

-- name: InsightsCountActiveStaff :one
SELECT COUNT(*) AS count FROM hr_staff WHERE business_id = ? AND status = 'active';

-- name: InsightsCountShiftsThisWeek :one
SELECT COUNT(*) AS count
FROM hr_shift_assignments
WHERE business_id = ?
  AND date >= DATE_SUB(CURDATE(), INTERVAL WEEKDAY(CURDATE()) DAY)
  AND date <  DATE_ADD(DATE_SUB(CURDATE(), INTERVAL WEEKDAY(CURDATE()) DAY), INTERVAL 7 DAY);

-- name: InsightsCountPendingLeave :one
SELECT COUNT(*) AS count FROM hr_leave_requests WHERE business_id = ? AND status = 'pending';

-- name: InsightsCountOpenLeave :one
SELECT COUNT(*) AS count
FROM hr_leave_requests
WHERE business_id = ?
  AND status NOT IN ('rejected','cancelled','completed');

-- Headcount on a date. Historical rows must not depend on the employee's
-- current status.
-- name: InsightsHeadcountAsOf :one
SELECT COUNT(*) AS count
FROM hr_employees
WHERE business_id = sqlc.arg(business_id)
  AND (start_date IS NULL OR start_date <= sqlc.arg(as_of))
  AND (end_date IS NULL OR end_date > sqlc.arg(as_of));

-- name: InsightsCountHires :one
SELECT COUNT(*) AS count
FROM hr_employees
WHERE business_id = sqlc.arg(business_id)
  AND start_date >= sqlc.arg(date_from)
  AND start_date <= sqlc.arg(date_to);

-- name: InsightsCountDepartures :one
SELECT COUNT(*) AS count
FROM hr_employees
WHERE business_id = sqlc.arg(business_id)
  AND end_date >= sqlc.arg(date_from)
  AND end_date <= sqlc.arg(date_to);

-- name: InsightsShiftMetrics :one
SELECT
  SUM(CASE WHEN status <> 'cancelled' THEN 1 ELSE 0 END) AS scheduled,
  SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END)  AS completed,
  SUM(CASE WHEN status = 'no_show'   THEN 1 ELSE 0 END)  AS no_shows
FROM hr_shift_assignments
WHERE business_id = sqlc.arg(business_id)
  AND date BETWEEN sqlc.arg(date_from) AND sqlc.arg(date_to);

-- name: InsightsCountLeaveStarting :one
SELECT COUNT(*) AS count
FROM hr_leave_requests
WHERE business_id = sqlc.arg(business_id)
  AND status = sqlc.arg(status)
  AND start_date BETWEEN sqlc.arg(date_from) AND sqlc.arg(date_to);

-- name: InsightsScheduledHours :one
SELECT COALESCE(SUM(
  TIME_TO_SEC(STR_TO_DATE(s.end_time, '%H:%i'))
  - TIME_TO_SEC(STR_TO_DATE(s.start_time, '%H:%i'))
) / 3600, 0) AS scheduled_hours
FROM hr_shift_assignments a
JOIN hr_shifts s ON s.id = a.shift_id AND s.business_id = a.business_id
WHERE a.business_id = sqlc.arg(business_id)
  AND a.status <> 'cancelled'
  AND a.date >= sqlc.arg(date_from)
  AND a.date <= sqlc.arg(date_to);

-- The latest change to the HR rows Insights reports on, used to version its
-- cached reports.
-- name: InsightsHRLastChange :one
SELECT CAST(GREATEST(
  COALESCE((SELECT MAX(e.updated_at) FROM hr_employees e WHERE e.business_id = sqlc.arg(business_id)), '1000-01-01'),
  COALESCE((SELECT MAX(l.updated_at) FROM hr_leave_requests l WHERE l.business_id = sqlc.arg(business_id)), '1000-01-01'),
  COALESCE((SELECT MAX(a.updated_at) FROM hr_shift_assignments a WHERE a.business_id = sqlc.arg(business_id)), '1000-01-01')
) AS DATETIME) AS last_change;
