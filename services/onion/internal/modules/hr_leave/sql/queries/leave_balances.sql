-- name: ListLeaveBalances :many
SELECT id, owner_id, business_id, employee_id, leave_type_id, year, entitled_days, used_days,
       pending_days, last_accrual_month, created_at, updated_at
FROM hr_leave_balances
WHERE business_id = ? AND employee_id = ? AND year = ?
ORDER BY leave_type_id ASC;

-- name: GetLeaveBalance :one
SELECT id, owner_id, business_id, employee_id, leave_type_id, year, entitled_days, used_days,
       pending_days, last_accrual_month, created_at, updated_at
FROM hr_leave_balances
WHERE business_id = ? AND employee_id = ? AND leave_type_id = ? AND year = ?;

-- name: UpsertLeaveBalance :execresult
INSERT INTO hr_leave_balances (owner_id, business_id, employee_id, leave_type_id, year, entitled_days, used_days, pending_days, last_accrual_month)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE entitled_days = VALUES(entitled_days), used_days = VALUES(used_days),
    pending_days = VALUES(pending_days), last_accrual_month = VALUES(last_accrual_month), updated_at = NOW();

-- name: AddToEntitledDays :exec
UPDATE hr_leave_balances
SET entitled_days = entitled_days + ?, last_accrual_month = ?, updated_at = NOW()
WHERE business_id = ? AND employee_id = ? AND leave_type_id = ? AND year = ?;

-- name: AddToUsedDays :execresult
UPDATE hr_leave_balances
SET used_days = used_days + ?, pending_days = GREATEST(pending_days - ?, 0), updated_at = NOW()
WHERE business_id = ? AND employee_id = ? AND leave_type_id = ? AND year = ?;

-- name: AddToPendingDays :exec
UPDATE hr_leave_balances
SET pending_days = pending_days + ?, updated_at = NOW()
WHERE business_id = ? AND employee_id = ? AND leave_type_id = ? AND year = ?;

-- name: SubtractPendingDays :exec
UPDATE hr_leave_balances
SET pending_days = GREATEST(pending_days - ?, 0), updated_at = NOW()
WHERE business_id = ? AND employee_id = ? AND leave_type_id = ? AND year = ?;

-- name: ListBalancesForAccrual :many
SELECT id, owner_id, business_id, employee_id, leave_type_id, year, entitled_days, used_days,
       pending_days, last_accrual_month, created_at, updated_at
FROM hr_leave_balances
WHERE business_id = ? AND year = ? AND last_accrual_month < ?;
