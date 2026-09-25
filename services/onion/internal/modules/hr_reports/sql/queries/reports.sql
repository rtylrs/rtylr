-- cross-schema read from hr_core — headcount uses employee, department, and location records.
-- name: HeadcountReport :many
SELECT COALESCE(d.name, 'Unassigned') AS department,
       COALESCE(l.name, 'Unassigned') AS location,
       e.employment_type,
       e.status,
       COUNT(*) AS employee_count
FROM hr_employees e
LEFT JOIN hr_departments d ON d.id = e.department_id AND d.business_id = e.business_id
LEFT JOIN hr_locations l ON l.id = e.location_id AND l.business_id = e.business_id
WHERE e.business_id = ?
  AND e.status = 'active'
  AND e.created_at <= ?
  AND (sqlc.narg('location_id') IS NULL OR e.location_id = sqlc.narg('location_id'))
  AND (sqlc.narg('department_id') IS NULL OR e.department_id = sqlc.narg('department_id'))
  AND (sqlc.narg('employee_id') IS NULL OR e.id = sqlc.narg('employee_id'))
GROUP BY d.name, l.name, e.employment_type, e.status
ORDER BY department, location
LIMIT 5001;

-- cross-schema read from hr_core and hr_attendance — summarize employee attendance by period.
-- name: AttendanceSummaryReport :many
SELECT e.id AS employee_id,
       CONCAT(e.first_name, ' ', e.last_name) AS employee_name,
       COALESCE(d.name, 'Unassigned') AS department,
       COUNT(a.id) AS total_records,
       SUM(CASE WHEN a.status = 'present' THEN 1 ELSE 0 END) AS days_present,
       SUM(CASE WHEN a.is_late = TRUE THEN 1 ELSE 0 END) AS days_late,
       COALESCE(SUM(a.late_minutes), 0) AS total_late_minutes,
       COALESCE(SUM(a.duration_minutes), 0) AS total_duration_minutes
FROM hr_employees e
LEFT JOIN hr_departments d ON d.id = e.department_id AND d.business_id = e.business_id
LEFT JOIN hr_attendance_records a ON a.employee_id = e.id AND a.business_id = e.business_id AND a.date BETWEEN ? AND ?
WHERE e.business_id = ?
  AND e.status = 'active'
  AND (sqlc.narg('location_id') IS NULL OR e.location_id = sqlc.narg('location_id'))
  AND (sqlc.narg('department_id') IS NULL OR e.department_id = sqlc.narg('department_id'))
  AND (sqlc.narg('employee_id') IS NULL OR e.id = sqlc.narg('employee_id'))
GROUP BY e.id, e.first_name, e.last_name, d.name
ORDER BY employee_name
LIMIT 5001;

-- cross-schema read from hr_core and hr_leave — summarize leave balances and requests.
-- name: LeaveSummaryReport :many
SELECT e.id AS employee_id,
       CONCAT(e.first_name, ' ', e.last_name) AS employee_name,
       COALESCE(d.name, 'Unassigned') AS department,
       lt.name AS leave_type,
       COALESCE(lb.entitled_days, 0) AS entitled_days,
       COALESCE(lb.used_days, 0) AS used_days,
       COALESCE(lb.pending_days, 0) AS pending_days,
       COALESCE(lb.entitled_days - lb.used_days - lb.pending_days, 0) AS remaining_days,
       COALESCE(lr.requests_in_period, 0) AS requests_in_period,
       COALESCE(lr.approved_days_in_period, 0) AS approved_days_in_period
FROM hr_employees e
LEFT JOIN hr_departments d ON d.id = e.department_id AND d.business_id = e.business_id
CROSS JOIN hr_leave_types lt ON lt.business_id = e.business_id AND lt.is_active = TRUE
LEFT JOIN hr_leave_balances lb ON lb.employee_id = e.id AND lb.business_id = e.business_id AND lb.leave_type_id = lt.id AND lb.year = YEAR(?)
LEFT JOIN (
    SELECT req.employee_id, req.business_id, req.leave_type_id,
           COUNT(*) AS requests_in_period,
           COALESCE(SUM(CASE WHEN req.status = 'approved' THEN req.total_days ELSE 0 END), 0) AS approved_days_in_period
    FROM hr_leave_requests req
    WHERE req.business_id = ? AND req.start_date <= ? AND req.end_date >= ?
    GROUP BY req.employee_id, req.business_id, req.leave_type_id
) lr ON lr.employee_id = e.id AND lr.business_id = e.business_id AND lr.leave_type_id = lt.id
WHERE e.business_id = ?
  AND e.status = 'active'
  AND (sqlc.narg('location_id') IS NULL OR e.location_id = sqlc.narg('location_id'))
  AND (sqlc.narg('department_id') IS NULL OR e.department_id = sqlc.narg('department_id'))
  AND (sqlc.narg('employee_id') IS NULL OR e.id = sqlc.narg('employee_id'))
GROUP BY e.id, e.first_name, e.last_name, d.name, lt.id, lt.name, lb.entitled_days, lb.used_days, lb.pending_days, lr.requests_in_period, lr.approved_days_in_period
ORDER BY employee_name, leave_type
LIMIT 5001;

-- cross-schema read from hr_core and hr_payroll — summarize payroll entries by employee.
-- name: PayrollSummaryReport :many
SELECT e.id AS employee_id,
       CONCAT(e.first_name, ' ', e.last_name) AS employee_name,
       COALESCE(d.name, 'Unassigned') AS department,
       COALESCE(SUM(pe.base_pay), 0) AS total_base_pay,
       COALESCE(SUM(pe.overtime_pay), 0) AS total_overtime_pay,
       COALESCE(SUM(pe.bonuses), 0) AS total_bonuses,
       COALESCE(SUM(pe.deductions), 0) AS total_deductions,
       COALESCE(SUM(pe.base_pay + pe.overtime_pay + pe.bonuses), 0) AS total_gross_pay,
       COALESCE(SUM(pe.net_pay), 0) AS total_net_pay,
       COALESCE(SUM(pe.hours_worked), 0) AS total_hours_worked,
       COALESCE(SUM(pe.overtime_hours), 0) AS total_overtime_hours
FROM hr_employees e
LEFT JOIN hr_departments d ON d.id = e.department_id AND d.business_id = e.business_id
LEFT JOIN hr_payroll_entries pe ON pe.employee_id = e.id AND pe.business_id = e.business_id
LEFT JOIN hr_payroll_runs pr ON pr.id = pe.payroll_run_id AND pr.business_id = e.business_id AND pr.period_start >= ? AND pr.period_end <= ?
WHERE e.business_id = ?
  AND e.status = 'active'
  AND (sqlc.narg('location_id') IS NULL OR e.location_id = sqlc.narg('location_id'))
  AND (sqlc.narg('department_id') IS NULL OR e.department_id = sqlc.narg('department_id'))
  AND (sqlc.narg('employee_id') IS NULL OR e.id = sqlc.narg('employee_id'))
GROUP BY e.id, e.first_name, e.last_name, d.name
ORDER BY employee_name
LIMIT 5001;

-- cross-schema read from hr_core and hr_overtime — summarize overtime requests by period.
-- name: OvertimeSummaryReport :many
SELECT e.id AS employee_id,
       CONCAT(e.first_name, ' ', e.last_name) AS employee_name,
       COALESCE(d.name, 'Unassigned') AS department,
       COUNT(o.id) AS total_requests,
       SUM(CASE WHEN o.status = 'approved' THEN 1 ELSE 0 END) AS approved_requests,
       SUM(CASE WHEN o.status = 'pending' THEN 1 ELSE 0 END) AS pending_requests,
       SUM(CASE WHEN o.status = 'rejected' THEN 1 ELSE 0 END) AS rejected_requests,
       COALESCE(SUM(CASE WHEN o.status = 'approved' THEN o.planned_extra_minutes ELSE 0 END), 0) AS approved_overtime_minutes,
       COALESCE(AVG(CASE WHEN o.status = 'approved' THEN o.multiplier_applied ELSE NULL END), 0) AS avg_multiplier
FROM hr_employees e
LEFT JOIN hr_departments d ON d.id = e.department_id AND d.business_id = e.business_id
LEFT JOIN hr_overtime_requests o ON o.employee_id = e.id AND o.business_id = e.business_id AND o.date BETWEEN ? AND ?
WHERE e.business_id = ?
  AND e.status = 'active'
  AND (sqlc.narg('location_id') IS NULL OR e.location_id = sqlc.narg('location_id'))
  AND (sqlc.narg('department_id') IS NULL OR e.department_id = sqlc.narg('department_id'))
  AND (sqlc.narg('employee_id') IS NULL OR e.id = sqlc.narg('employee_id'))
GROUP BY e.id, e.first_name, e.last_name, d.name
ORDER BY employee_name
LIMIT 5001;

-- cross-schema read from hr_core and hr_time — summarize timesheets by period.
-- name: TimesheetSummaryReport :many
SELECT e.id AS employee_id,
       CONCAT(e.first_name, ' ', e.last_name) AS employee_name,
       COALESCE(d.name, 'Unassigned') AS department,
       COUNT(DISTINCT ts.id) AS total_timesheets,
       COALESCE(SUM(ts.total_minutes), 0) AS total_logged_minutes,
       COALESCE(SUM(CASE WHEN ts.status = 'approved' THEN ts.total_minutes ELSE 0 END), 0) AS approved_minutes,
       COALESCE(SUM(CASE WHEN ts.status = 'draft' THEN ts.total_minutes ELSE 0 END), 0) AS draft_minutes,
       COALESCE(SUM(CASE WHEN ts.status = 'submitted' THEN ts.total_minutes ELSE 0 END), 0) AS submitted_minutes
FROM hr_employees e
LEFT JOIN hr_departments d ON d.id = e.department_id AND d.business_id = e.business_id
LEFT JOIN hr_timesheets ts ON ts.employee_id = e.id AND ts.business_id = e.business_id AND ts.period_start >= ? AND ts.period_end <= ?
WHERE e.business_id = ?
  AND e.status = 'active'
  AND (sqlc.narg('location_id') IS NULL OR e.location_id = sqlc.narg('location_id'))
  AND (sqlc.narg('department_id') IS NULL OR e.department_id = sqlc.narg('department_id'))
  AND (sqlc.narg('employee_id') IS NULL OR e.id = sqlc.narg('employee_id'))
GROUP BY e.id, e.first_name, e.last_name, d.name
ORDER BY employee_name
LIMIT 5001;
