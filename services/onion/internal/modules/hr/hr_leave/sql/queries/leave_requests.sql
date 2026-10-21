-- name: ListLeaveRequests :many
-- Date range uses overlap semantics (crm.sales-style doc in voxie#47): a
-- request is included when it starts on/before to_date and ends on/after
-- from_date, so multi-day leave crossing the boundary is never missed.
SELECT id, owner_id, business_id, employee_id, leave_type_id, start_date, end_date, total_days,
       notes, status, current_approval_level, rejection_reason, created_at, updated_at
FROM hr_leave_requests
WHERE business_id = ?
  AND (? = 0 OR employee_id = ?)
  AND (? = '' OR status = ?)
  AND (sqlc.narg('from_date') IS NULL OR end_date >= sqlc.narg('from_date'))
  AND (sqlc.narg('to_date') IS NULL OR start_date <= sqlc.narg('to_date'))
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: CountLeaveRequests :one
SELECT COUNT(*) AS total
FROM hr_leave_requests
WHERE business_id = ?
  AND (? = 0 OR employee_id = ?)
  AND (? = '' OR status = ?)
  AND (sqlc.narg('from_date') IS NULL OR end_date >= sqlc.narg('from_date'))
  AND (sqlc.narg('to_date') IS NULL OR start_date <= sqlc.narg('to_date'));

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

-- name: AdvanceLeaveApprovalLevel :execresult
UPDATE hr_leave_requests
SET current_approval_level = current_approval_level + 1, updated_at = NOW()
WHERE id = ? AND business_id = ? AND status = 'pending' AND current_approval_level = ?;

-- name: RejectLeaveRequest :execresult
UPDATE hr_leave_requests
SET status = 'rejected', rejection_reason = ?, updated_at = NOW()
WHERE id = ? AND business_id = ? AND status = 'pending';

-- The status = 'pending' guard is a second, DB-level check alongside the
-- service-layer one: RowsAffected = 0 tells the caller the request was no
-- longer pending (approved/rejected/canceled between the read and this
-- write) rather than silently succeeding on a request it should not touch.
-- name: CancelLeaveRequest :execresult
UPDATE hr_leave_requests
SET status = 'canceled', updated_at = NOW()
WHERE id = ? AND business_id = ? AND status = 'pending';

-- Bounded-result exemption: scoped to one employee's own pending/approved
-- requests overlapping the date range of a single new request being
-- validated, called internally from CreateLeaveRequest rather than through
-- any caller-facing pagination. A single employee cannot realistically have
-- more than a handful of overlapping requests, so this is a safety net.
-- name: ListOverlappingLeaveRequests :many
SELECT id, owner_id, business_id, employee_id, leave_type_id, start_date, end_date, total_days,
       notes, status, current_approval_level, rejection_reason, created_at, updated_at
FROM hr_leave_requests
WHERE business_id = ? AND employee_id = ? AND status IN ('pending', 'approved')
  AND start_date <= ? AND end_date >= ?
LIMIT 500;

-- Bounded-result exemption: the calendar handler accepts an open-ended
-- caller-supplied date range with no limit/offset in its own contract, and
-- spans every employee in the business rather than one. Capped generously so
-- a wide or unbounded date range cannot pull the whole company's leave
-- history in one query.
-- name: ListLeaveCalendar :many
SELECT lr.id, lr.owner_id, lr.business_id, lr.employee_id, lr.leave_type_id,
       lr.start_date, lr.end_date, lr.total_days, lr.notes, lr.status,
       lr.current_approval_level, lr.rejection_reason, lr.created_at, lr.updated_at
FROM hr_leave_requests AS lr
LEFT JOIN hr_employees AS employee
  ON employee.id = lr.employee_id
 AND employee.business_id = lr.business_id
WHERE lr.business_id = ? AND lr.status IN ('approved', 'pending')
  AND lr.start_date <= sqlc.arg(date_to) AND lr.end_date >= sqlc.arg(date_from)
  AND (sqlc.narg(location_id) IS NULL OR employee.location_id = sqlc.narg(location_id))
ORDER BY lr.start_date ASC
LIMIT 1000;
