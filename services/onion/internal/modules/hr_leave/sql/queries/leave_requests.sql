-- name: ListLeaveRequests :many
SELECT id, owner_id, business_id, employee_id, leave_type_id, start_date, end_date, total_days,
       notes, status, current_approval_level, rejection_reason, created_at, updated_at
FROM hr_leave_requests
WHERE business_id = ?
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: CountLeaveRequests :one
SELECT COUNT(*) AS total FROM hr_leave_requests WHERE business_id = ?;

-- name: GetLeaveRequest :one
SELECT id, owner_id, business_id, employee_id, leave_type_id, start_date, end_date, total_days,
       notes, status, current_approval_level, rejection_reason, created_at, updated_at
FROM hr_leave_requests
WHERE business_id = ? AND id = ?;

-- name: CreateLeaveRequest :execresult
INSERT INTO hr_leave_requests (
    owner_id, business_id, employee_id, leave_type_id, start_date, end_date, total_days,
    notes, status, current_approval_level
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: ApproveLeaveRequest :exec
UPDATE hr_leave_requests
SET status = 'approved', updated_at = NOW()
WHERE id = ? AND business_id = ?;

-- name: AdvanceLeaveApprovalLevel :exec
UPDATE hr_leave_requests
SET current_approval_level = current_approval_level + 1, updated_at = NOW()
WHERE id = ? AND business_id = ?;

-- name: RejectLeaveRequest :exec
UPDATE hr_leave_requests
SET status = 'rejected', rejection_reason = ?, updated_at = NOW()
WHERE id = ? AND business_id = ?;

-- name: ListOverlappingLeaveRequests :many
SELECT id, owner_id, business_id, employee_id, leave_type_id, start_date, end_date, total_days,
       notes, status, current_approval_level, rejection_reason, created_at, updated_at
FROM hr_leave_requests
WHERE business_id = ? AND employee_id = ? AND status IN ('pending', 'approved')
  AND start_date <= ? AND end_date >= ?;

-- name: ListLeaveCalendar :many
SELECT id, owner_id, business_id, employee_id, leave_type_id, start_date, end_date, total_days,
       notes, status, current_approval_level, rejection_reason, created_at, updated_at
FROM hr_leave_requests
WHERE business_id = ? AND status IN ('approved', 'pending')
  AND start_date <= ? AND end_date >= ?
ORDER BY start_date ASC;
