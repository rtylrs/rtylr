-- name: CreatePayrollRecord :execresult
INSERT INTO hr_payroll_records
(owner_id, business_id, staff_id, period_start, period_end, period_label, regular_hours, overtime_hours, base_pay, overtime_pay, net_pay)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: MarkPayrollPaid :execresult
UPDATE hr_payroll_records
SET status = 'paid', paid_at = NOW()
WHERE id = ? AND business_id = ? AND status != 'paid';

-- cross-schema read from hr_staff — use staff compensation rates for payroll calculations.
-- name: GetStaffPayRates :one
SELECT salary, hourly_rate, overtime_rate FROM hr_staff WHERE id = ? AND business_id = ?;

-- cross-schema read from hr_shifts — aggregate completed attendance for payroll periods.
-- name: GetCompletedShiftsForPeriod :many
SELECT staff_id, SUM(hours_worked) AS total_hours, SUM(overtime_hours) AS total_overtime
FROM hr_shifts
WHERE business_id = ? AND status = 'completed'
  AND date >= ? AND date <= ?
GROUP BY staff_id
LIMIT 200;

-- name: ListPayrollRecords :many
SELECT id, owner_id, business_id, staff_id, period_start, period_end, period_label,
       regular_hours, overtime_hours, base_pay, overtime_pay, bonuses,
       deductions, tax_withheld, net_pay, status, paid_at, notes, created_at
FROM hr_payroll_records
WHERE business_id = ?
ORDER BY period_start DESC, staff_id ASC
LIMIT 200;

-- name: ListPayrollRecordsByPeriod :many
SELECT id, owner_id, business_id, staff_id, period_start, period_end, period_label,
       regular_hours, overtime_hours, base_pay, overtime_pay, bonuses,
       deductions, tax_withheld, net_pay, status, paid_at, notes, created_at
FROM hr_payroll_records
WHERE business_id = ? AND period_label = ?
ORDER BY period_start DESC, staff_id ASC
LIMIT 200;

-- cross-schema read from hr_staff — resolve compensation rates for selected staff.
-- name: GetStaffPayRatesByIDs :many
SELECT id, salary, hourly_rate, overtime_rate
FROM hr_staff
WHERE business_id = ? AND id IN (sqlc.slice('staff_ids'));

-- cross-schema read from hr_overtime_config — apply the configured overtime multiplier.
-- name: GetOvertimeConfigForLegacyPayroll :one
SELECT default_multiplier
FROM hr_overtime_config
WHERE business_id = ?;
