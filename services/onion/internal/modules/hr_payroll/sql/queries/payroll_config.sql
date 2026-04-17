-- name: GetPayrollConfig :one
SELECT id, owner_id, period_type, pay_day_of_month, overtime_enabled, created_at, updated_at
FROM hr_payroll_config
WHERE owner_id = ?;

-- name: UpsertPayrollConfig :execresult
INSERT INTO hr_payroll_config (owner_id, period_type, pay_day_of_month, overtime_enabled)
VALUES (?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
    period_type = VALUES(period_type),
    pay_day_of_month = VALUES(pay_day_of_month),
    overtime_enabled = VALUES(overtime_enabled);
