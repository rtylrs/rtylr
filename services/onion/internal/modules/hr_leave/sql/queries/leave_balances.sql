-- Bounded-result exemption: scoped to one employee's balances for one year,
-- naturally bounded by the business's own configured leave type count. No
-- caller paginates this (GetLeaveBalances returns the whole set for the
-- employee/year pair), so a generous hard cap is a safety net rather than a
-- real limit in practice.
-- name: ListLeaveBalances :many
SELECT id, owner_id, business_id, employee_id, leave_type_id, year, entitled_days, used_days,
       pending_days, last_accrual_month, created_at, updated_at
FROM hr_leave_balances
WHERE business_id = ? AND employee_id = ? AND year = ?
ORDER BY leave_type_id ASC
LIMIT 500;

-- name: GetLeaveBalance :one
SELECT id, owner_id, business_id, employee_id, leave_type_id, year, entitled_days, used_days,
       pending_days, last_accrual_month, created_at, updated_at
FROM hr_leave_balances
WHERE business_id = ? AND employee_id = ? AND leave_type_id = ? AND year = ?;

-- name: GetLeaveBalanceForUpdate :one
SELECT id, owner_id, business_id, employee_id, leave_type_id, year, entitled_days, used_days,
       pending_days, last_accrual_month, created_at, updated_at
FROM hr_leave_balances
WHERE business_id = ? AND employee_id = ? AND leave_type_id = ? AND year = ?
FOR UPDATE;

-- name: UpsertLeaveBalance :execresult
INSERT INTO hr_leave_balances (owner_id, business_id, employee_id, leave_type_id, year, entitled_days, used_days, pending_days, last_accrual_month)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE entitled_days = VALUES(entitled_days), used_days = VALUES(used_days),
    pending_days = VALUES(pending_days), last_accrual_month = VALUES(last_accrual_month), updated_at = NOW();

-- name: AddToEntitledDays :execresult
UPDATE hr_leave_balances
SET entitled_days = entitled_days + ?, last_accrual_month = ?, updated_at = NOW()
WHERE business_id = ? AND employee_id = ? AND leave_type_id = ? AND year = ?
  AND last_accrual_month < ?;

-- Carry forward the unused, non-pending portion of the prior year's balance
-- once at the start of the new year. last_accrual_month = 0 is the seeded
-- row's rollover marker; switching it to -1 makes the operation idempotent
-- before the January accrual advances it to month 1.
-- name: ApplyYearEndCarryover :execresult
INSERT INTO hr_leave_balances (
    owner_id, business_id, employee_id, leave_type_id, year,
    entitled_days, used_days, pending_days, last_accrual_month
)
SELECT previous.owner_id, previous.business_id, previous.employee_id, previous.leave_type_id, ?,
       LEAST(GREATEST(previous.entitled_days - previous.used_days - previous.pending_days, 0), leave_type.max_carryover_days),
       0, 0, -1
FROM hr_leave_balances AS previous
JOIN hr_leave_types AS leave_type
  ON leave_type.business_id = previous.business_id
 AND leave_type.id = previous.leave_type_id
WHERE previous.business_id = ? AND previous.year = ? AND leave_type.is_active = TRUE
ON DUPLICATE KEY UPDATE
    entitled_days = entitled_days + IF(last_accrual_month = 0, VALUES(entitled_days), 0),
    last_accrual_month = IF(last_accrual_month = 0, -1, last_accrual_month),
    updated_at = NOW();

-- name: AddToUsedDays :execresult
UPDATE hr_leave_balances
SET used_days = used_days + ?, pending_days = GREATEST(pending_days - ?, 0), updated_at = NOW()
WHERE business_id = ? AND employee_id = ? AND leave_type_id = ? AND year = ?;

-- name: AddToPendingDays :execresult
UPDATE hr_leave_balances
SET pending_days = pending_days + ?, updated_at = NOW()
WHERE business_id = ? AND employee_id = ? AND leave_type_id = ? AND year = ?;

-- name: SubtractPendingDays :execresult
UPDATE hr_leave_balances
SET pending_days = GREATEST(pending_days - ?, 0), updated_at = NOW()
WHERE business_id = ? AND employee_id = ? AND leave_type_id = ? AND year = ?;

-- Bounded-result exemption: called once per owner by the monthly accrual job
-- to sweep every balance row still owed an accrual for the year, so it takes
-- no caller-supplied limit/offset. Capped generously rather than left
-- unbounded; any rows left over past the cap keep their un-advanced
-- last_accrual_month and are picked up again on the job's next run, so a
-- capped sweep degrades to "caught up a run late" instead of dropping rows.
-- name: ListBalancesForAccrual :many
SELECT id, owner_id, business_id, employee_id, leave_type_id, year, entitled_days, used_days,
       pending_days, last_accrual_month, created_at, updated_at
FROM hr_leave_balances
WHERE business_id = ? AND year = ? AND last_accrual_month < ?
LIMIT 2000;
