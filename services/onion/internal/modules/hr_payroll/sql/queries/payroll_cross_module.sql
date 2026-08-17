-- ListApprovedTimesheetSummaries returns one row per employee for a single
-- business, so it stays unpaginated by design - bounded by the same generous
-- hard LIMIT used for the other per-employee reports in this HR domain (see
-- hr_reports's AttendanceSummaryReport).
-- name: ListApprovedTimesheetSummaries :many
SELECT employee_id, SUM(total_minutes) AS total_minutes
FROM hr_timesheets
WHERE business_id = ? AND status = 'approved' AND period_start >= ? AND period_end <= ?
GROUP BY employee_id
ORDER BY employee_id
LIMIT 5001;

-- ListApprovedOvertimeSummaries returns one row per employee for a single
-- business, so it stays unpaginated by design - bounded by the same generous
-- hard LIMIT as ListApprovedTimesheetSummaries above.
-- name: ListApprovedOvertimeSummaries :many
SELECT employee_id, SUM(planned_extra_minutes) AS total_minutes, AVG(multiplier_applied) AS avg_multiplier
FROM hr_overtime_requests
WHERE business_id = ? AND status = 'approved' AND date >= ? AND date <= ?
GROUP BY employee_id
ORDER BY employee_id
LIMIT 5001;

-- ListEmployeePayInfo returns one row per active employee for a single
-- business, so it stays unpaginated by design - bounded by the same generous
-- hard LIMIT as the summaries above.
-- name: ListEmployeePayInfo :many
SELECT id, owner_id, business_id, salary, hourly_rate
FROM hr_employees
WHERE business_id = ? AND status = 'active'
ORDER BY id
LIMIT 5001;

-- name: GetOvertimeConfigForPayroll :one
SELECT default_multiplier, auto_detect_enabled
FROM hr_overtime_config
WHERE business_id = ?;
