-- name: ListLeaveRequests :many
SELECT id, owner_id, business_id, employee_id, leave_type_id, start_date, end_date, total_days,
       notes, status, current_approval_level, rejection_reason, created_at, updated_at
FROM hr_leave_requests
WHERE business_id = ?
  AND (? = 0 OR employee_id = ?)
  AND (? = '' OR status = ?)
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: CountLeaveRequests :one
SELECT COUNT(*) AS total
FROM hr_leave_requests
WHERE business_id = ?
  AND (? = 0 OR employee_id = ?)
  AND (? = '' OR status = ?);

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

-- name: ApproveLeaveRequest :execresult
UPDATE hr_leave_requests
SET status = 'approved', updated_at = NOW()
WHERE id = ? AND business_id = ? AND status = 'pending';

-- name: AdvanceLeaveApprovalLevel :exec
UPDATE hr_leave_requests
SET current_approval_level = current_approval_level + 1, updated_at = NOW()
WHERE id = ? AND business_id = ?;

-- name: RejectLeaveRequest :exec
UPDATE hr_leave_requests
SET status = 'rejected', rejection_reason = ?, updated_at = NOW()
WHERE id = ? AND business_id = ?;

-- The status = 'pending' guard is a second, DB-level check alongside the
-- service-layer one: RowsAffected = 0 tells the caller the request was no
-- longer pending (approved/rejected/canceled between the read and this
-- write) rather than silently succeeding on a request it should not touch.
-- name: CancelLeaveRequest :execresult
UPDATE hr_leave_requests
SET status = 'canceled', updated_at = NOW()
WHERE id = ? AND business_id = ? AND status = 'pending';

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
