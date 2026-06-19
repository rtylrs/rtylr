-- name: ListLeaveTypes :many
SELECT id, owner_id, business_id, name, is_paid, requires_approval, accrual_rule, accrual_amount,
       max_carryover_days, min_notice_days, requires_documentation_after_days,
       is_active, created_at, updated_at
FROM hr_leave_types
WHERE business_id = ? AND is_active = TRUE
ORDER BY name ASC
LIMIT ? OFFSET ?;

-- name: CountLeaveTypes :one
SELECT COUNT(*) AS total FROM hr_leave_types
WHERE business_id = ? AND is_active = TRUE;

-- name: GetLeaveType :one
SELECT id, owner_id, business_id, name, is_paid, requires_approval, accrual_rule, accrual_amount,
       max_carryover_days, min_notice_days, requires_documentation_after_days,
       is_active, created_at, updated_at
FROM hr_leave_types
WHERE business_id = ? AND id = ?;

-- name: CreateLeaveType :execresult
INSERT INTO hr_leave_types (
    owner_id, business_id, name, is_paid, requires_approval, accrual_rule, accrual_amount,
    max_carryover_days, min_notice_days, requires_documentation_after_days, is_active
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: UpdateLeaveType :exec
UPDATE hr_leave_types
SET name = ?, is_paid = ?, requires_approval = ?, accrual_rule = ?, accrual_amount = ?,
    max_carryover_days = ?, min_notice_days = ?, requires_documentation_after_days = ?,
    is_active = ?, updated_at = NOW()
WHERE id = ? AND business_id = ?;

-- name: ListAllActiveLeaveTypesForOwner :many
SELECT id, owner_id, business_id, name, is_paid, requires_approval, accrual_rule, accrual_amount,
       max_carryover_days, min_notice_days, requires_documentation_after_days,
       is_active, created_at, updated_at
FROM hr_leave_types
WHERE business_id = ? AND is_active = TRUE;

-- name: ListDistinctOwnerIDs :many
SELECT DISTINCT business_id FROM hr_leave_types WHERE is_active = TRUE;
