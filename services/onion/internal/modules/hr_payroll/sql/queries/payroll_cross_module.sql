-- ListApprovedTimesheetSummaries returns one row per employee for a single
-- business, so it stays unpaginated by design - bounded by the same generous
-- hard LIMIT used for the other per-employee reports in this HR domain (see
-- hr_reports's AttendanceSummaryReport).
-- cross-schema read from hr_time — include approved timesheets in payroll periods.
--
-- Final approval in hr_time locks the timesheet (LockTimesheet sets
-- status='locked'), so 'locked' is the status that represents approved time;
-- 'approved' is accepted too so rows written by older code (the currently
-- uncalled ApproveTimesheet statement) still count. If hr_time ever changes
-- what final approval writes, this set must change with it (B11-R05).
--
-- Minutes are summed from hr_time_entries, not read from
-- hr_timesheets.total_minutes: no write path maintains that column after
-- CreateTimesheet, so it is always 0 and hourly base pay would silently be zero
-- (B11-R05). Entries are attributed by employee and the timesheet's own period,
-- which hr_time_entries indexes on (business_id, employee_id, date), so the
-- optional hr_time_entries.timesheet_id is not required. A run that covers only
-- part of a timesheet period still prorates by the overlapping day count, as
-- this query did before.
-- name: ListApprovedTimesheetSummaries :many
SELECT t.employee_id,
       ROUND(SUM(
           (
               SELECT COALESCE(SUM(e.duration_minutes), 0)
               FROM hr_time_entries e
               WHERE e.business_id = t.business_id
                 AND e.employee_id = t.employee_id
                 AND e.date >= t.period_start
                 AND e.date <= t.period_end
           ) *
           (DATEDIFF(
                LEAST(t.period_end, sqlc.arg(period_end)),
                GREATEST(t.period_start, sqlc.arg(period_start))
            ) + 1) /
           (DATEDIFF(t.period_end, t.period_start) + 1)
       ), 0) AS total_minutes
FROM hr_timesheets t
WHERE t.business_id = ?
  AND t.status IN ('approved', 'locked')
  AND t.period_start <= sqlc.arg(period_end)
  AND t.period_end >= sqlc.arg(period_start)
GROUP BY t.employee_id
ORDER BY t.employee_id
LIMIT 5001;

-- ListApprovedOvertimeSummaries returns one row per employee for a single
-- business, so it stays unpaginated by design - bounded by the same generous
-- hard LIMIT as ListApprovedTimesheetSummaries above.
-- cross-schema read from hr_overtime — include approved overtime requests in payroll periods.
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
-- cross-schema read from hr_core — resolve active employee compensation data for payroll.
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

-- cross-schema read from hr_overtime — apply the configured payroll overtime multiplier.
-- name: GetOvertimeConfigForPayroll :one
SELECT default_multiplier, auto_detect_enabled
FROM hr_overtime_config
WHERE business_id = ?;
