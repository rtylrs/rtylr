-- name: CreatePayrollRecord :execresult
INSERT INTO hr_payroll_records
(owner_id, business_id, staff_id, period_start, period_end, period_label, regular_hours, overtime_hours, base_pay, overtime_pay, net_pay)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: MarkPayrollPaid :execresult
UPDATE hr_payroll_records
SET status = 'paid', paid_at = NOW()
WHERE id = ? AND business_id = ? AND status != 'paid';

-- name: GetStaffPayRates :one
SELECT salary, hourly_rate, overtime_rate FROM hr_staff WHERE id = ? AND business_id = ?;

-- name: GetCompletedShiftsForPeriod :many
SELECT staff_id, SUM(hours_worked) AS total_hours, SUM(overtime_hours) AS total_overtime
FROM hr_shifts
WHERE business_id = ? AND status = 'completed'
  AND date >= ? AND date <= ?
GROUP BY staff_id;
