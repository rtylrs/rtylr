-- cross-schema read from hr_core — count active employees for operations insights.
-- name: CountActiveStaffOps :one
SELECT COUNT(*) AS count FROM hr_employees WHERE business_id = ? AND status = 'active';

-- cross-schema read from hr_shifts — count scheduled assignments for the current week.
-- name: CountShiftsThisWeek :one
SELECT COUNT(*) AS count
 FROM hr_shift_assignments
 WHERE business_id = ?
   AND date >= DATE_SUB(CURDATE(), INTERVAL WEEKDAY(CURDATE()) DAY)
   AND date <  DATE_ADD(DATE_SUB(CURDATE(), INTERVAL WEEKDAY(CURDATE()) DAY), INTERVAL 7 DAY);

-- cross-schema read from hr_leave — count pending leave requests for operations insights.
-- name: CountPendingLeaveOps :one
SELECT COUNT(*) AS count FROM hr_leave_requests WHERE business_id = ? AND status = 'pending';

-- cross-schema read from hr_leave — count non-terminal leave requests.
-- name: CountOpenLeaveRequests :one
SELECT COUNT(*) AS count
 FROM hr_leave_requests
 WHERE business_id = ?
   AND status NOT IN ('rejected','cancelled','completed');

-- cross-schema read from hr_core — calculate headcount as of a date. Historical
-- rows must not depend on the employee's current status.
-- name: GetHeadcountAsOf :one
SELECT COUNT(*) AS count
FROM hr_employees
WHERE business_id = ?
  AND (start_date IS NULL OR start_date <= ?)
  AND (end_date IS NULL OR end_date > ?);

-- cross-schema read from hr_core — count gross hires during a period.
-- name: CountHeadcountHires :one
SELECT COUNT(*) AS count
FROM hr_employees
WHERE business_id = ?
  AND start_date >= ?
  AND start_date <= ?;

-- cross-schema read from hr_core — count gross departures during a period.
-- name: CountHeadcountDepartures :one
SELECT COUNT(*) AS count
FROM hr_employees
WHERE business_id = ?
  AND end_date >= ?
  AND end_date <= ?;

-- cross-schema read from hr_shifts — summarize assignment statuses for a date range.
-- name: GetShiftMetrics :one
SELECT
  SUM(CASE WHEN status <> 'cancelled' THEN 1 ELSE 0 END) AS scheduled,
  SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END)  AS completed,
  SUM(CASE WHEN status = 'no_show'   THEN 1 ELSE 0 END)  AS no_shows
FROM hr_shift_assignments
WHERE business_id = ?
  AND date BETWEEN ? AND ?;

-- cross-schema read from hr_leave — count requests starting during a period.
-- name: CountPendingTimeOff :one
SELECT COUNT(*) AS count
FROM hr_leave_requests
WHERE business_id = ?
  AND status = 'pending'
  AND start_date BETWEEN ? AND ?;

-- cross-schema read from hr_leave — count approved requests starting during a period.
-- name: CountApprovedTimeOff :one
SELECT COUNT(*) AS count
FROM hr_leave_requests
WHERE business_id = ?
  AND status = 'approved'
  AND start_date BETWEEN ? AND ?;

-- cross-schema read from hr_core — count employees active at the end of a period.
-- name: CountEmployeesAsOf :one
SELECT COUNT(*) AS count
FROM hr_employees
WHERE business_id = ?
  AND (start_date IS NULL OR start_date <= ?)
  AND (end_date IS NULL OR end_date > ?);

-- cross-schema read from hr_shifts — sum scheduled hours during a period.
-- name: SumScheduledHoursInRange :one
SELECT COALESCE(SUM(
  TIME_TO_SEC(STR_TO_DATE(s.end_time, '%H:%i'))
  - TIME_TO_SEC(STR_TO_DATE(s.start_time, '%H:%i'))
) / 3600, 0) AS scheduled_hours
FROM hr_shift_assignments a
JOIN hr_shifts s ON s.id = a.shift_id AND s.business_id = a.business_id
WHERE a.business_id = ?
  AND a.status <> 'cancelled'
  AND a.date >= ?
  AND a.date <= ?;
