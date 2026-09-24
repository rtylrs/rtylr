-- name: ListTimesheets :many
SELECT id, owner_id, business_id, employee_id, period_start, period_end, status,
       current_approval_level, submitted_at, total_minutes, rejection_reason,
       created_at, updated_at
FROM hr_timesheets
WHERE business_id = ?
  AND (sqlc.narg('employee_id') IS NULL OR employee_id = sqlc.narg('employee_id'))
  AND (sqlc.narg('status') IS NULL OR status = sqlc.narg('status'))
ORDER BY period_start DESC
LIMIT ? OFFSET ?;

-- name: CountTimesheets :one
SELECT COUNT(*) AS total FROM hr_timesheets
WHERE business_id = ?
  AND (sqlc.narg('employee_id') IS NULL OR employee_id = sqlc.narg('employee_id'))
  AND (sqlc.narg('status') IS NULL OR status = sqlc.narg('status'));

-- name: GetTimesheet :one
SELECT id, owner_id, business_id, employee_id, period_start, period_end, status,
       current_approval_level, submitted_at, total_minutes, rejection_reason,
       created_at, updated_at
FROM hr_timesheets
WHERE business_id = ? AND id = ?;

-- name: GetTimesheetForPeriod :one
SELECT id, owner_id, business_id, employee_id, period_start, period_end, status,
       current_approval_level, submitted_at, total_minutes, rejection_reason,
       created_at, updated_at
FROM hr_timesheets
WHERE business_id = ? AND employee_id = ? AND period_start = ?;

-- name: CreateTimesheet :execresult
INSERT INTO hr_timesheets (
    owner_id, business_id, employee_id, period_start, period_end, status,
    current_approval_level, total_minutes
) VALUES (?, ?, ?, ?, ?, ?, ?, ?);

-- name: SubmitTimesheet :exec
UPDATE hr_timesheets
SET status = 'submitted', submitted_at = NOW(), current_approval_level = 1, updated_at = NOW()
WHERE id = ? AND business_id = ?;

-- name: AdvanceApprovalLevel :execresult
-- The status and level predicates make the increment conditional on the state
-- the caller read: an overlapping or retried approval (carrying the level it
-- observed) affects 0 rows instead of advancing twice and skipping a level,
-- and a draft/rejected/locked timesheet cannot be advanced at all. Mirrors
-- hr_overtime and hr_leave, whose advance queries carry the same guard.
UPDATE hr_timesheets
SET current_approval_level = current_approval_level + 1, updated_at = NOW()
WHERE id = ? AND business_id = ? AND status = 'submitted' AND current_approval_level = ?;

-- name: ApproveTimesheet :exec
UPDATE hr_timesheets
SET status = 'approved', updated_at = NOW()
WHERE id = ? AND business_id = ?;

-- name: LockTimesheet :execresult
-- Final approval may only lock a submitted timesheet: a second, overlapping
-- final approval then affects 0 rows and is reported as a conflict instead of
-- writing a duplicate final_approved action.
UPDATE hr_timesheets
SET status = 'locked', updated_at = NOW()
WHERE id = ? AND business_id = ? AND status = 'submitted';

-- name: RejectTimesheet :execrows
UPDATE hr_timesheets
SET status = 'rejected', rejection_reason = ?, updated_at = NOW()
WHERE id = ? AND business_id = ? AND status = 'submitted';

-- name: CreateApprovalAction :exec
INSERT INTO hr_timesheet_approval_actions (
    owner_id, business_id, timesheet_id, approver_id, action, level, comment
) VALUES (?, ?, ?, ?, ?, ?, ?);

-- The reminder job has no pagination contract; keep the endpoint bounded
-- until it can page through employees instead of scanning them all at once.
-- name: ListUnsubmittedEmployees :many
SELECT DISTINCT e.employee_id
FROM hr_time_entries e
LEFT JOIN hr_timesheets t ON t.business_id = e.business_id
    AND t.employee_id = e.employee_id
    AND t.period_start = ?
WHERE e.business_id = ?
    AND e.date >= ? AND e.date <= ?
    AND (t.id IS NULL OR t.status = 'draft')
ORDER BY e.employee_id ASC
LIMIT 500;
