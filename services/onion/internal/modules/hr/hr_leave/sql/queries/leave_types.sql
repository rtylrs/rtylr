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

-- name: UpdateLeaveType :execresult
UPDATE hr_leave_types
SET name = ?, is_paid = ?, requires_approval = ?, accrual_rule = ?, accrual_amount = ?,
    max_carryover_days = ?, min_notice_days = ?, requires_documentation_after_days = ?,
    is_active = ?, updated_at = NOW()
WHERE id = ? AND business_id = ?;

-- Partial update: an absent optional field keeps its stored value. Explicit
-- false and zero values remain valid when their parameter is present.
-- name: PatchLeaveType :execresult
UPDATE hr_leave_types
SET name = COALESCE(sqlc.narg(name), name),
    is_paid = COALESCE(sqlc.narg(is_paid), is_paid),
    requires_approval = COALESCE(sqlc.narg(requires_approval), requires_approval),
    accrual_rule = COALESCE(sqlc.narg(accrual_rule), accrual_rule),
    accrual_amount = COALESCE(sqlc.narg(accrual_amount), accrual_amount),
    max_carryover_days = COALESCE(sqlc.narg(max_carryover_days), max_carryover_days),
    min_notice_days = COALESCE(sqlc.narg(min_notice_days), min_notice_days),
    requires_documentation_after_days = COALESCE(sqlc.narg(requires_documentation_after_days), requires_documentation_after_days),
    is_active = COALESCE(sqlc.narg(is_active), is_active),
    updated_at = NOW()
WHERE id = ? AND business_id = ?;

-- Bounded-result exemption: called once per owner by the monthly accrual job,
-- not through any caller-facing pagination. A business's own active leave
-- types are a small, admin-configured list, so a generous hard cap is a
-- safety net rather than a real limit in practice.
-- name: ListAllActiveLeaveTypesForOwner :many
SELECT id, owner_id, business_id, name, is_paid, requires_approval, accrual_rule, accrual_amount,
       max_carryover_days, min_notice_days, requires_documentation_after_days,
       is_active, created_at, updated_at
FROM hr_leave_types
WHERE business_id = ? AND is_active = TRUE
LIMIT 500;

-- Bounded-result exemption: enumerates every tenant across the whole platform
-- for the monthly accrual cron job, so it cannot take a caller-supplied
-- limit/offset the way a per-tenant list can. Capped generously so a single
-- run stays bounded; if the platform ever approaches this many businesses
-- with active leave types, this needs real keyset pagination in the job
-- instead of raising the constant.
-- name: ListDistinctOwnerIDs :many
SELECT DISTINCT business_id FROM hr_leave_types WHERE is_active = TRUE
LIMIT 5000;
