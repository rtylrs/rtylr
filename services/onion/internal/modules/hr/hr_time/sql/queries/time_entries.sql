-- Every filter besides business_id is optional and they combine with AND, which
-- is what hr.time_entry.v1 ListTimeEntriesRequest documents. employee_id was a
-- mandatory equality before tryrtylr/rtylr#855, so "all employees over this
-- range" was not answerable however the caller asked.
--
-- name: ListTimeEntries :many
SELECT id, owner_id, business_id, employee_id, timesheet_id, project_id, task_id, category_id,
       date, duration_minutes, start_time, end_time, description, is_billable,
       status, rejection_reason, approved_by, source, billable_invoice_id,
       created_at, updated_at
FROM hr_time_entries
WHERE business_id = ?
  AND (sqlc.narg('employee_id') IS NULL OR employee_id = sqlc.narg('employee_id'))
  AND (sqlc.narg('from_date') IS NULL OR date >= sqlc.narg('from_date'))
  AND (sqlc.narg('to_date') IS NULL OR date <= sqlc.narg('to_date'))
  AND (sqlc.narg('project_id') IS NULL OR project_id = sqlc.narg('project_id'))
  AND (sqlc.narg('is_billable') IS NULL OR is_billable = sqlc.narg('is_billable'))
  AND (sqlc.narg('status') IS NULL OR status = sqlc.narg('status'))
ORDER BY date DESC, created_at DESC
LIMIT ? OFFSET ?;

-- Filters must match ListTimeEntries exactly, or total describes a different
-- set from the page it is reported alongside.
--
-- name: CountTimeEntries :one
SELECT COUNT(*) AS total FROM hr_time_entries
WHERE business_id = ?
  AND (sqlc.narg('employee_id') IS NULL OR employee_id = sqlc.narg('employee_id'))
  AND (sqlc.narg('from_date') IS NULL OR date >= sqlc.narg('from_date'))
  AND (sqlc.narg('to_date') IS NULL OR date <= sqlc.narg('to_date'))
  AND (sqlc.narg('project_id') IS NULL OR project_id = sqlc.narg('project_id'))
  AND (sqlc.narg('is_billable') IS NULL OR is_billable = sqlc.narg('is_billable'))
  AND (sqlc.narg('status') IS NULL OR status = sqlc.narg('status'));

-- name: ListUnbilledBillableTimeEntries :many
SELECT id, owner_id, business_id, employee_id, timesheet_id, project_id, task_id, category_id,
       date, duration_minutes, start_time, end_time, description, is_billable,
       status, rejection_reason, approved_by, source, billable_invoice_id,
       created_at, updated_at
FROM hr_time_entries
WHERE business_id = ?
  AND is_billable = TRUE
  AND status = 'approved'
  AND billable_invoice_id IS NULL
ORDER BY date ASC, id ASC
LIMIT ? OFFSET ?;

-- name: CountUnbilledBillableTimeEntries :one
SELECT COUNT(*) AS total,
       CAST(COALESCE(SUM(duration_minutes), 0) AS SIGNED) AS total_minutes
FROM hr_time_entries
WHERE business_id = ?
  AND is_billable = TRUE
  AND status = 'approved'
  AND billable_invoice_id IS NULL;

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

-- The reminder job has no pagination contract; keep the endpoint bounded
-- until it can page through businesses instead of scanning them all at once.
-- name: ListTimesheetBusinessIDs :many
SELECT DISTINCT business_id FROM hr_time_entries
UNION
SELECT DISTINCT business_id FROM hr_timesheets
ORDER BY business_id ASC
LIMIT 500;

-- Final timesheet approval owns the entry workflow transition. Entries created
-- through the self-service API may not have a timesheet_id yet, so unlinked
-- rows are matched by employee and period; explicitly linked rows remain bound
-- to their timesheet. This runs in the same transaction as the final lock.
-- name: ApproveTimesheetEntries :execrows
UPDATE hr_time_entries
SET timesheet_id = COALESCE(timesheet_id, ?),
    status = 'approved',
    approved_by = ?,
    rejection_reason = NULL,
    updated_at = NOW()
WHERE business_id = ?
  AND employee_id = ?
  AND (timesheet_id = ? OR (timesheet_id IS NULL AND date >= ? AND date <= ?))
  AND status IN ('draft', 'rejected');

-- name: MarkEntryInvoiced :execrows
-- The invoice-side half of the billable loop (#836/#791), applied per entry
-- by the repository inside one transaction so the whole batch lands or none
-- of it does. Idempotent (already-invoiced rows do not match), scoped to the
-- caller's business, and restricted to approved billable work so a stray id
-- can never stamp an unrelated row.
UPDATE hr_time_entries
SET billable_invoice_id = ?, updated_at = NOW()
WHERE business_id = ?
  AND id = ?
  AND is_billable = TRUE
  AND status = 'approved'
  AND billable_invoice_id IS NULL;
