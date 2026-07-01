-- name: ListTimeEntries :many
SELECT id, owner_id, business_id, employee_id, timesheet_id, project_id, task_id, category_id,
       date, duration_minutes, start_time, end_time, description, is_billable,
       status, rejection_reason, approved_by, source, billable_invoice_id,
       created_at, updated_at
FROM hr_time_entries
WHERE business_id = ? AND employee_id = ? AND date >= ? AND date <= ?
ORDER BY date DESC, created_at DESC
LIMIT ? OFFSET ?;

-- name: CountTimeEntries :one
SELECT COUNT(*) AS total FROM hr_time_entries
WHERE business_id = ? AND employee_id = ? AND date >= ? AND date <= ?;

-- name: GetTimeEntry :one
SELECT id, owner_id, business_id, employee_id, timesheet_id, project_id, task_id, category_id,
       date, duration_minutes, start_time, end_time, description, is_billable,
       status, rejection_reason, approved_by, source, billable_invoice_id,
       created_at, updated_at
FROM hr_time_entries
WHERE business_id = ? AND id = ?;

-- name: CreateTimeEntry :execresult
INSERT INTO hr_time_entries (
    owner_id, business_id, employee_id, timesheet_id, project_id, task_id, category_id,
    date, duration_minutes, start_time, end_time, description, is_billable,
    status, rejection_reason, approved_by, source
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: UpdateTimeEntry :exec
UPDATE hr_time_entries
SET employee_id = ?, timesheet_id = ?, project_id = ?, task_id = ?, category_id = ?,
    date = ?, duration_minutes = ?, start_time = ?, end_time = ?, description = ?,
    is_billable = ?, status = ?, rejection_reason = ?, approved_by = ?, source = ?,
    updated_at = NOW()
WHERE id = ? AND business_id = ?;

-- name: DeleteTimeEntry :exec
DELETE FROM hr_time_entries WHERE id = ? AND business_id = ?;

-- name: GetTimeEntryBySourceAndTask :one
SELECT id, owner_id, business_id, employee_id, timesheet_id, project_id, task_id, category_id,
       date, duration_minutes, start_time, end_time, description, is_billable,
       status, rejection_reason, approved_by, source, billable_invoice_id,
       created_at, updated_at
FROM hr_time_entries
WHERE business_id = ? AND employee_id = ? AND source = ? AND task_id = ?
LIMIT 1;

-- name: ListApprovedBillableEntries :many
SELECT id, owner_id, business_id, employee_id, timesheet_id, project_id, task_id, category_id,
       date, duration_minutes, start_time, end_time, description, is_billable,
       status, rejection_reason, approved_by, source, billable_invoice_id,
       created_at, updated_at
FROM hr_time_entries
WHERE business_id = ? AND is_billable = TRUE AND status = 'approved'
      AND billable_invoice_id IS NULL
ORDER BY date ASC;

-- name: MarkEntryExported :exec
UPDATE hr_time_entries
SET billable_invoice_id = ?, updated_at = NOW()
WHERE id = ? AND business_id = ?;

-- name: GetExistingTaskIDsBySource :many
SELECT task_id FROM hr_time_entries
WHERE business_id = ? AND employee_id = ? AND source = ?
  AND task_id IN (sqlc.slice('task_ids'));

-- name: ListTimesheetBusinessIDs :many
SELECT DISTINCT business_id FROM hr_time_entries
UNION
SELECT DISTINCT business_id FROM hr_timesheets;
