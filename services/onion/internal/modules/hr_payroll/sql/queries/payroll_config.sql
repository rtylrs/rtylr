-- name: GetPayrollConfig :one
SELECT id, owner_id, business_id, period_type, pay_day_of_month, overtime_enabled, created_at, updated_at
FROM hr_payroll_config
WHERE business_id = ?;

-- ListAllPayrollConfigs loads every business's payroll config in one query
-- (see the caller for why), so it stays unpaginated by design - bounded by
-- the same generous hard LIMIT as ListDistinctPayrollOwnerIDs above it.
-- name: ListAllPayrollConfigs :many
SELECT id, owner_id, business_id, period_type, pay_day_of_month, overtime_enabled, created_at, updated_at
FROM hr_payroll_config
ORDER BY business_id
LIMIT 5000;

-- name: UpsertPayrollConfig :execresult
INSERT INTO hr_payroll_config (owner_id, business_id, period_type, pay_day_of_month, overtime_enabled)
VALUES (?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
    period_type = VALUES(period_type),
    pay_day_of_month = VALUES(pay_day_of_month),
    overtime_enabled = VALUES(overtime_enabled);
