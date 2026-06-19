-- name: ListTimesheets :many
SELECT id, owner_id, business_id, employee_id, period_start, period_end, status,
       current_approval_level, submitted_at, total_minutes, rejection_reason,
       created_at, updated_at
FROM hr_timesheets
WHERE business_id = ? AND employee_id = ?
ORDER BY period_start DESC
LIMIT ? OFFSET ?;

-- name: CountTimesheets :one
SELECT COUNT(*) AS total FROM hr_timesheets
WHERE business_id = ? AND employee_id = ?;

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

-- name: UpdateTimesheetStatus :exec
UPDATE hr_timesheets
SET status = ?, updated_at = NOW()
WHERE id = ? AND business_id = ?;

-- name: SubmitTimesheet :exec
UPDATE hr_timesheets
SET status = 'submitted', submitted_at = NOW(), current_approval_level = 1, updated_at = NOW()
WHERE id = ? AND business_id = ?;

-- name: AdvanceApprovalLevel :exec
UPDATE hr_timesheets
SET current_approval_level = current_approval_level + 1, updated_at = NOW()
WHERE id = ? AND business_id = ?;

-- name: ApproveTimesheet :exec
UPDATE hr_timesheets
SET status = 'approved', updated_at = NOW()
WHERE id = ? AND business_id = ?;

-- name: LockTimesheet :exec
UPDATE hr_timesheets
SET status = 'locked', updated_at = NOW()
WHERE id = ? AND business_id = ?;

-- name: RejectTimesheet :exec
UPDATE hr_timesheets
SET status = 'rejected', rejection_reason = ?, updated_at = NOW()
WHERE id = ? AND business_id = ?;

-- name: UpdateTimesheetTotalMinutes :exec
UPDATE hr_timesheets
SET total_minutes = ?, updated_at = NOW()
WHERE id = ? AND business_id = ?;

-- name: CreateApprovalAction :exec
INSERT INTO hr_timesheet_approval_actions (
    owner_id, business_id, timesheet_id, approver_id, action, level, comment
) VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: ListUnsubmittedEmployees :many
SELECT DISTINCT e.employee_id
FROM hr_time_entries e
LEFT JOIN hr_timesheets t ON t.business_id = e.business_id
    AND t.employee_id = e.employee_id
    AND t.period_start = ?
WHERE e.business_id = ?
    AND e.date >= ? AND e.date <= ?
    AND (t.id IS NULL OR t.status = 'draft');
