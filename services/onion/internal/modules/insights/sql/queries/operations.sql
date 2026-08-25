-- name: CountActiveStaffOps :one
SELECT COUNT(*) AS count FROM hr_employees WHERE business_id = ? AND status = 'active';

-- name: CountShiftsThisWeek :one
SELECT COUNT(*) AS count
 FROM hr_shift_assignments
 WHERE business_id = ?
   AND date >= DATE_SUB(CURDATE(), INTERVAL WEEKDAY(CURDATE()) DAY)
   AND date <  DATE_ADD(DATE_SUB(CURDATE(), INTERVAL WEEKDAY(CURDATE()) DAY), INTERVAL 7 DAY);

-- name: CountPendingLeaveOps :one
SELECT COUNT(*) AS count FROM hr_leave_requests WHERE business_id = ? AND status = 'pending';

-- name: CountOpenLeaveRequests :one
SELECT COUNT(*) AS count
 FROM hr_leave_requests
 WHERE business_id = ?
   AND status NOT IN ('rejected','cancelled','completed');

-- name: GetHeadcountNow :one
SELECT COUNT(*) AS count FROM hr_employees
 WHERE business_id = ?
   AND (start_date IS NULL OR start_date <= ?)
   AND (end_date IS NULL OR end_date > ?)
   AND status = 'active';

-- name: GetHeadcountPrior :one
SELECT COUNT(*) AS count FROM hr_employees
WHERE business_id = ?
   AND (start_date IS NULL OR start_date <= ?)
   AND (end_date IS NULL OR end_date > ?)
   AND status = 'active';

-- name: GetShiftMetrics :one
SELECT
  SUM(CASE WHEN status <> 'cancelled' THEN 1 ELSE 0 END) AS scheduled,
  SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END)  AS completed,
  SUM(CASE WHEN status = 'no_show'   THEN 1 ELSE 0 END)  AS no_shows
FROM hr_shift_assignments
WHERE business_id = ?
  AND date BETWEEN ? AND ?;

-- name: CountPendingTimeOff :one
SELECT COUNT(*) AS count FROM hr_leave_requests WHERE business_id = ? AND status = 'pending';

-- name: CountApprovedTimeOff :one
SELECT COUNT(*) AS count FROM hr_leave_requests WHERE business_id = ? AND status = 'approved';

-- name: CountActiveEmployeesUtilization :one
SELECT COUNT(*) AS count FROM hr_employees WHERE business_id = ? AND status = 'active';

-- name: SumScheduledHoursNext7Days :one
SELECT COALESCE(SUM(
  TIME_TO_SEC(STR_TO_DATE(s.end_time, '%H:%i'))
  - TIME_TO_SEC(STR_TO_DATE(s.start_time, '%H:%i'))
) / 3600, 0) AS scheduled_hours
FROM hr_shift_assignments a
JOIN hr_shifts s ON s.id = a.shift_id AND s.business_id = a.business_id
WHERE a.business_id = ?
  AND a.status <> 'cancelled'
  AND a.date BETWEEN CURDATE() AND DATE_ADD(CURDATE(), INTERVAL 7 DAY);
