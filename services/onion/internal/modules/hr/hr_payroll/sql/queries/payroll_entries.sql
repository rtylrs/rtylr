-- name: CreatePayrollEntry :execresult
INSERT INTO hr_payroll_entries (owner_id, business_id, payroll_run_id, employee_id, hours_worked, overtime_hours, base_pay, overtime_pay, bonuses, deductions, net_pay)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- ListPayrollEntries returns every entry for one payroll run, one row per
-- employee paid in that run, so it stays unpaginated by design - bounded by
-- the same generous hard LIMIT used for other per-employee listings in this
-- module (see ListEmployeePayInfo).
-- name: ListPayrollEntries :many
SELECT id, owner_id, business_id, payroll_run_id, employee_id, hours_worked, overtime_hours, base_pay, overtime_pay, bonuses, deductions, net_pay, created_at, updated_at
FROM hr_payroll_entries
WHERE business_id = ? AND payroll_run_id = ?
ORDER BY employee_id
LIMIT 5001;

-- name: GetPayrollEntryForEmployee :one
SELECT id, owner_id, business_id, payroll_run_id, employee_id, hours_worked, overtime_hours, base_pay, overtime_pay, bonuses, deductions, net_pay, created_at, updated_at
FROM hr_payroll_entries
WHERE business_id = ? AND payroll_run_id = ? AND employee_id = ?;
