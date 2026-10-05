-- name: ListOvertimeRequests :many
SELECT id, owner_id, business_id, employee_id, date, planned_extra_minutes, reason, status,
       current_approval_level, rejection_reason, approved_by, multiplier_applied, created_at, updated_at
FROM hr_overtime_requests
WHERE business_id = ?
  AND (sqlc.narg('employee_id') IS NULL OR employee_id = sqlc.narg('employee_id'))
  AND (sqlc.narg('status') IS NULL OR status = sqlc.narg('status'))
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: CountOvertimeRequests :one
SELECT COUNT(*) AS total
FROM hr_overtime_requests
WHERE business_id = ?
  AND (sqlc.narg('employee_id') IS NULL OR employee_id = sqlc.narg('employee_id'))
  AND (sqlc.narg('status') IS NULL OR status = sqlc.narg('status'));

-- name: GetOvertimeRequest :one
SELECT id, owner_id, business_id, employee_id, date, planned_extra_minutes, reason, status,
       current_approval_level, rejection_reason, approved_by, multiplier_applied, created_at, updated_at
FROM hr_overtime_requests
WHERE business_id = ? AND id = ?;

-- name: CreateOvertimeRequest :execresult
INSERT INTO hr_overtime_requests (
    owner_id, business_id, employee_id, date, planned_extra_minutes, reason, status, current_approval_level
) VALUES (?, ?, ?, ?, ?, ?, ?, ?);

-- name: ApproveOvertimeRequest :execresult
UPDATE hr_overtime_requests
SET status = 'approved', approved_by = ?, multiplier_applied = ?, rejection_reason = NULL, updated_at = NOW()
WHERE id = ? AND business_id = ? AND status = 'pending';

-- name: AdvanceOvertimeApprovalLevel :execresult
UPDATE hr_overtime_requests
SET current_approval_level = current_approval_level + 1, updated_at = NOW()
WHERE id = ? AND business_id = ? AND status = 'pending' AND current_approval_level = ?;

-- name: RejectOvertimeRequest :execresult
UPDATE hr_overtime_requests
SET status = 'rejected', rejection_reason = ?, updated_at = NOW()
WHERE id = ? AND business_id = ? AND status = 'pending';

-- name: CreateFlaggedOvertimeRequest :execresult
INSERT INTO hr_overtime_requests (
    owner_id, business_id, employee_id, date, planned_extra_minutes, reason, status, current_approval_level
) SELECT ?, ?, ?, ?, ?, 'auto-detected overtime', 'flagged', 0
WHERE NOT EXISTS (
    SELECT 1 FROM hr_overtime_requests AS existing
    WHERE existing.business_id = ? AND existing.employee_id = ? AND existing.date = ?
      AND existing.status IN ('flagged', 'pending', 'approved')
);

-- name: ListFlaggedForDate :many
SELECT id, owner_id, business_id, employee_id, date, planned_extra_minutes, reason, status,
       current_approval_level, rejection_reason, approved_by, multiplier_applied, created_at, updated_at
FROM hr_overtime_requests
WHERE business_id = ? AND date = ? AND status IN ('flagged', 'pending', 'approved');
-- The detector must treat an existing pending/approved request as a dedupe
-- hit too; otherwise rerunning the job creates a second request after a flag
-- has already entered the approval flow.

-- name: ListOvertimeCandidates :many
SELECT employee_id, date, SUM(duration_minutes) - ? AS extra_minutes
FROM hr_time_entries
WHERE business_id = ? AND date = ? AND status IN ('submitted', 'approved')
GROUP BY employee_id, date
HAVING SUM(duration_minutes) > ?
ORDER BY employee_id ASC;
