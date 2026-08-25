-- ListApprovedTimesheetSummaries returns one row per employee for a single
-- business, so it stays unpaginated by design - bounded by the same generous
-- hard LIMIT used for the other per-employee reports in this HR domain (see
-- hr_reports's AttendanceSummaryReport).
-- name: ListApprovedTimesheetSummaries :many
SELECT employee_id,
       ROUND(SUM(
           total_minutes *
           (DATEDIFF(
                LEAST(period_end, sqlc.arg(period_end)),
                GREATEST(period_start, sqlc.arg(period_start))
            ) + 1) /
           (DATEDIFF(period_end, period_start) + 1)
       ), 0) AS total_minutes
FROM hr_timesheets
WHERE business_id = ?
  AND status = 'approved'
  AND period_start <= sqlc.arg(period_end)
  AND period_end >= sqlc.arg(period_start)
GROUP BY employee_id
ORDER BY employee_id
LIMIT 5001;

-- ListApprovedOvertimeSummaries returns one row per employee for a single
-- business, so it stays unpaginated by design - bounded by the same generous
-- hard LIMIT as ListApprovedTimesheetSummaries above.
-- name: ListApprovedOvertimeSummaries :many
SELECT employee_id,
       SUM(planned_extra_minutes) AS total_minutes,
       SUM(planned_extra_minutes * COALESCE(multiplier_applied, 0)) AS weighted_minutes,
       SUM(CASE WHEN multiplier_applied IS NULL THEN planned_extra_minutes ELSE 0 END) AS unweighted_minutes
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

-- A run must be revalidated at approval/export time because an employee can be
-- offboarded after the run was generated.
-- name: CountInactivePayrollEmployees :one
SELECT COUNT(*) AS inactive_count
FROM hr_payroll_entries e
LEFT JOIN hr_employees emp
  ON emp.business_id = e.business_id AND emp.id = e.employee_id
WHERE e.business_id = ?
  AND e.payroll_run_id = ?
  AND (emp.id IS NULL OR emp.status <> 'active');

-- name: GetOvertimeConfigForPayroll :one
SELECT default_multiplier, auto_detect_enabled
FROM hr_overtime_config
WHERE business_id = ?;
