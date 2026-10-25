-- Flow daily time logging: work days, their lines and the work-type tree.
--
-- Every read and write is scoped to a business.

-- name: GetWorkDay :one
SELECT id, employee_id, date, timezone, work_mode, start_time,
         end_time, break_minutes, status, logged_by, approved_by,
         approved_at, reopened_by, reopened_at, created_at, updated_at
FROM flow_work_days
WHERE business_id = ? AND employee_id = ? AND date = ?
LIMIT 1;

-- name: GetWorkDayByID :one
SELECT id, employee_id, date, timezone, work_mode, start_time,
         end_time, break_minutes, status, logged_by, approved_by,
         approved_at, reopened_by, reopened_at, created_at, updated_at
FROM flow_work_days
WHERE id = ? AND business_id = ?
LIMIT 1;

-- Upsert, because a day is identified by who and when rather than by an id the
-- client holds: the first save of a day and its later corrections are the same
-- call. Leans on uk_work_day (business, employee, date).
-- name: UpsertWorkDay :execresult
INSERT INTO flow_work_days (
    business_id, employee_id, date, timezone,
    work_mode, start_time, end_time, break_minutes, logged_by
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
    timezone = VALUES(timezone),
    work_mode = VALUES(work_mode),
    start_time = VALUES(start_time),
    end_time = VALUES(end_time),
    break_minutes = VALUES(break_minutes),
    updated_at = CURRENT_TIMESTAMP;

-- name: ListWorkDays :many
SELECT id, employee_id, date, timezone, work_mode, start_time,
         end_time, break_minutes, status, logged_by, approved_by,
         approved_at, reopened_by, reopened_at, created_at, updated_at
FROM flow_work_days
WHERE business_id = ? AND employee_id = ?
  AND date >= ? AND date <= ?
ORDER BY date DESC
LIMIT ? OFFSET ?;

-- name: CountWorkDays :one
SELECT COUNT(*) FROM flow_work_days
WHERE business_id = ? AND employee_id = ?
  AND date >= ? AND date <= ?;

-- ── Work types ─────────────────────────────────────────────────────────────

-- name: ListWorkTypes :many
SELECT id, parent_id, name, description, hr_category_id,
         sort_order, archived
FROM flow_work_types
WHERE business_id = ?
  AND (archived = FALSE OR ? = TRUE)
ORDER BY sort_order ASC, name ASC;

-- name: GetWorkType :one
SELECT id, parent_id, name, description, hr_category_id,
         sort_order, archived
FROM flow_work_types
WHERE id = ? AND business_id = ?
LIMIT 1;

-- name: CreateWorkType :execresult
INSERT INTO flow_work_types (
    workspace_id, business_id, parent_id, name, description, hr_category_id, sort_order
) VALUES (0, ?, ?, ?, ?, ?, ?);

-- name: UpdateWorkType :execresult
UPDATE flow_work_types
SET name = ?, description = ?, parent_id = ?, hr_category_id = ?, sort_order = ?, archived = ?
WHERE id = ? AND business_id = ?;

-- Archives rather than deletes, which is what the archived column was added for.
-- There is no foreign key on hr_time_entries.work_type_id, so a hard DELETE left
-- every entry that used the type pointing at an id that no longer exists, and
-- timesheets that were already submitted and approved silently lost their
-- classification. Those entries feed approvals and Finance's billable pull, so
-- the classification has to stay readable.
--
-- ListWorkTypes already hides archived rows unless a caller asks for them, so
-- the type disappears from Log Hours exactly as it did before.
--
-- archived = FALSE keeps the row count honest: archiving an already-archived
-- type affects nothing, and the caller is told it was not found rather than
-- that it was deleted a second time.
-- name: DeleteWorkType :execresult
UPDATE flow_work_types
SET archived = TRUE
WHERE id = ? AND business_id = ? AND archived = FALSE;

-- Guards the two-level rule: a node may only be nested under a top-level type.
-- name: CountWorkTypeChildren :one
SELECT COUNT(*) FROM flow_work_types
WHERE business_id = ? AND parent_id = ?;

-- name: ListWorkDaysByEmployeeIDs :many
SELECT id, employee_id, date, timezone, work_mode, start_time,
         end_time, break_minutes, status, logged_by, approved_by,
         approved_at, reopened_by, reopened_at, created_at, updated_at
FROM flow_work_days
WHERE business_id = ?
  AND employee_id IN (sqlc.slice('employee_ids'))
  AND date >= ? AND date <= ?
ORDER BY employee_id, date DESC
LIMIT ? OFFSET ?;

-- name: CountWorkDaysByEmployeeIDs :one
SELECT COUNT(*) FROM flow_work_days
WHERE business_id = ?
  AND employee_id IN (sqlc.slice('employee_ids'))
  AND date >= ? AND date <= ?;

-- Flow reads only its work-day records. The roster contract supplies the
-- authorized organization-member IDs and active HR employee IDs.
-- name: ListTeamWorkDaysByRoster :many
SELECT id, employee_id, date, timezone, work_mode, start_time,
       end_time, break_minutes, status, logged_by, approved_by,
       approved_at, reopened_by, reopened_at, created_at, updated_at
FROM flow_work_days
WHERE business_id = ?
  AND date >= ? AND date <= ?
  AND ((employee_id IN (sqlc.slice('employee_ids'))
        AND (employee_id <> logged_by OR employee_id NOT IN (sqlc.slice('hidden_ids'))))
       OR (employee_id = logged_by AND employee_id IN (sqlc.slice('user_ids'))))
ORDER BY employee_id, date DESC
LIMIT ? OFFSET ?;

-- name: CountTeamWorkDaysByRoster :one
SELECT COUNT(*) FROM flow_work_days
WHERE business_id = ?
  AND date >= ? AND date <= ?
  AND ((employee_id IN (sqlc.slice('employee_ids'))
        AND (employee_id <> logged_by OR employee_id NOT IN (sqlc.slice('hidden_ids'))))
       OR (employee_id = logged_by AND employee_id IN (sqlc.slice('user_ids'))));
