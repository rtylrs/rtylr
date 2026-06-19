-- name: ListApprovedTimesheetSummaries :many
SELECT employee_id, SUM(total_minutes) AS total_minutes
FROM hr_timesheets
WHERE business_id = ? AND status = 'approved' AND period_start >= ? AND period_end <= ?
GROUP BY employee_id;

-- name: ListApprovedOvertimeSummaries :many
SELECT employee_id, SUM(planned_extra_minutes) AS total_minutes, AVG(multiplier_applied) AS avg_multiplier
FROM hr_overtime_requests
WHERE business_id = ? AND status = 'approved' AND date >= ? AND date <= ?
GROUP BY employee_id;

-- name: ListEmployeePayInfo :many
SELECT id, owner_id, business_id, salary, hourly_rate
FROM hr_employees
WHERE business_id = ? AND status = 'active';

-- name: GetOvertimeConfigForPayroll :one
SELECT default_multiplier, auto_detect_enabled
FROM hr_overtime_config
WHERE business_id = ?;
