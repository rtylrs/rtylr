-- name: CreatePayrollRecord :execresult
INSERT INTO hr_payroll_records
(owner_id, business_id, staff_id, period_start, period_end, period_label, regular_hours, overtime_hours, base_pay, overtime_pay, net_pay)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: MarkPayrollPaid :execresult
UPDATE hr_payroll_records
SET status = 'paid', paid_at = NOW()
WHERE id = ? AND business_id = ? AND status != 'paid';

-- A payroll record's dates identify the period it belongs to, so settling one
-- record has to know which period to lock before it can look for duplicates.
-- name: GetPayrollRecordPeriod :one
SELECT period_start, period_end FROM hr_payroll_records WHERE id = ? AND business_id = ?;

-- Locking read over one period. It serialises a re-run against a settlement
-- and a settlement against a re-run, and it is how both callers detect the
-- duplicate pair a re-run used to leave behind.
-- name: ListPayrollPeriodRowsForUpdate :many
SELECT id, staff_id, status FROM hr_payroll_records
WHERE business_id = ? AND period_start = ? AND period_end = ?
FOR UPDATE;

-- Only ever deletes a period's unpaid rows: a paid row is money already out
-- the door and is never removed by a recalculation.
-- name: DeleteUnpaidPayrollRecords :exec
DELETE FROM hr_payroll_records
WHERE business_id = ? AND period_label = ? AND status <> 'paid';

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
