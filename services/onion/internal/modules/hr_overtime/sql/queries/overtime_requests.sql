-- name: ListOvertimeRequests :many
SELECT id, owner_id, employee_id, date, planned_extra_minutes, reason, status,
       current_approval_level, approved_by, multiplier_applied, created_at, updated_at
FROM hr_overtime_requests
WHERE owner_id = ?
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: CountOvertimeRequests :one
SELECT COUNT(*) AS total FROM hr_overtime_requests WHERE owner_id = ?;

-- name: GetOvertimeRequest :one
SELECT id, owner_id, employee_id, date, planned_extra_minutes, reason, status,
       current_approval_level, approved_by, multiplier_applied, created_at, updated_at
FROM hr_overtime_requests
WHERE owner_id = ? AND id = ?;

-- name: CreateOvertimeRequest :execresult
INSERT INTO hr_overtime_requests (
    owner_id, employee_id, date, planned_extra_minutes, reason, status, current_approval_level
) VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: ApproveOvertimeRequest :exec
UPDATE hr_overtime_requests
SET status = 'approved', approved_by = ?, multiplier_applied = ?, updated_at = NOW()
WHERE id = ? AND owner_id = ?;

-- name: AdvanceOvertimeApprovalLevel :exec
UPDATE hr_overtime_requests
SET current_approval_level = current_approval_level + 1, updated_at = NOW()
WHERE id = ? AND owner_id = ?;

-- name: RejectOvertimeRequest :exec
UPDATE hr_overtime_requests
SET status = 'rejected', updated_at = NOW()
WHERE id = ? AND owner_id = ?;

-- name: CreateFlaggedOvertimeRequest :execresult
INSERT INTO hr_overtime_requests (
    owner_id, employee_id, date, planned_extra_minutes, reason, status, current_approval_level
) VALUES (?, ?, ?, ?, 'auto-detected overtime', 'flagged', 0);

-- name: ListFlaggedForDate :many
SELECT id, owner_id, employee_id, date, planned_extra_minutes, reason, status,
       current_approval_level, approved_by, multiplier_applied, created_at, updated_at
FROM hr_overtime_requests
WHERE owner_id = ? AND date = ? AND status = 'flagged';
