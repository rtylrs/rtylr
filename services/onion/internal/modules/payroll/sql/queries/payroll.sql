-- name: CreatePayrollRecord :execresult
INSERT INTO hr_payroll_records
(owner_id, staff_id, period_start, period_end, period_label, regular_hours, overtime_hours, base_pay, overtime_pay, net_pay)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: MarkPayrollPaid :execresult
UPDATE hr_payroll_records
SET status = 'paid', paid_at = NOW()
WHERE id = ? AND owner_id = ? AND status != 'paid';

-- name: GetStaffPayRates :one
SELECT salary, hourly_rate, overtime_rate FROM hr_staff WHERE id = ? AND owner_id = ?;
