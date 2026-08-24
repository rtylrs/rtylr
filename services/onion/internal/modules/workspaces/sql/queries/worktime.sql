-- Flow daily time logging: work days, their lines and the work-type tree.
--
-- Every statement scopes on owner_id AND business_id. Scoping on owner alone is
-- what turns the older tables' `business_id DEFAULT 0` into a cross-tenant read,
-- and these tables deliberately have no such default to fall back on.

-- name: GetWorkDay :one
SELECT id, employee_id, date, timezone, work_mode, start_time,
         end_time, break_minutes, status, logged_by, approved_by,
         approved_at, reopened_by, reopened_at, created_at, updated_at
FROM flow_work_days
WHERE owner_id = ? AND business_id = ? AND employee_id = ? AND date = ?
LIMIT 1;

-- name: GetWorkDayByID :one
SELECT id, employee_id, date, timezone, work_mode, start_time,
         end_time, break_minutes, status, logged_by, approved_by,
         approved_at, reopened_by, reopened_at, created_at, updated_at
FROM flow_work_days
WHERE id = ? AND owner_id = ? AND business_id = ?
LIMIT 1;

-- Upsert, because a day is identified by who and when rather than by an id the
-- client holds: the first save of a day and its later corrections are the same
-- call. Leans on uk_work_day (owner, business, employee, date).
-- name: UpsertWorkDay :execresult
INSERT INTO flow_work_days (
    owner_id, business_id, employee_id, date, timezone,
    work_mode, start_time, end_time, break_minutes, logged_by
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
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
WHERE owner_id = ? AND business_id = ? AND employee_id = ?
  AND date >= ? AND date <= ?
ORDER BY date DESC
LIMIT ? OFFSET ?;

-- name: CountWorkDays :one
SELECT COUNT(*) FROM flow_work_days
WHERE owner_id = ? AND business_id = ? AND employee_id = ?
  AND date >= ? AND date <= ?;

-- name: InsertWorkDayAction :execresult
INSERT INTO flow_work_day_actions (
    owner_id, business_id, work_day_id, actor_id, action, comment
) VALUES (?, ?, ?, ?, ?, ?);

-- ── Lines ──────────────────────────────────────────────────────────────────
-- Lines are hr_time_entries: that table is what HR payroll reads, and every row
-- written here carries source='flow' and a task_id.
--
-- Flow does keep a second store of hours. flow_time_logs holds the ad-hoc
-- per-task durations written by CreateTimeLog in workspaces.sql, which payroll
-- never reads. Note that the per-task rollups over there (RecalcActualHours,
-- GetInsightsTimeByMember) read that store alone, so the lines written here do
-- not reach a task's actual_hours: tryrtylr/rtylr#903. Do not fix that by
-- summing both stores, since the same hour can legitimately be in each.

-- name: ListWorkDayLines :many
SELECT id, work_day_id, project_id, task_id, work_type_id,
           work_mode, description, duration_minutes, is_billable, status,
           client_key, logged_by, created_at, updated_at
FROM hr_time_entries
WHERE owner_id = ? AND business_id = ? AND work_day_id = ?
ORDER BY id ASC;

-- name: GetTimeEntryByClientKey :one
SELECT id, work_day_id, project_id, task_id, work_type_id,
           work_mode, description, duration_minutes, is_billable, status,
           client_key, logged_by, created_at, updated_at
FROM hr_time_entries
WHERE owner_id = ? AND business_id = ? AND client_key = ?
LIMIT 1;

-- name: GetTimeEntry :one
SELECT id, work_day_id, project_id, task_id, work_type_id,
           work_mode, description, duration_minutes, is_billable, status,
           client_key, logged_by, created_at, updated_at
FROM hr_time_entries
WHERE id = ? AND owner_id = ? AND business_id = ?
LIMIT 1;

-- name: CreateWorkDayLine :execresult
INSERT INTO hr_time_entries (
    owner_id, business_id, employee_id, work_day_id, date,
    project_id, task_id, work_type_id, work_mode, timezone,
    duration_minutes, description, is_billable, client_key, logged_by, source
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'flow');

-- name: UpdateWorkDayLine :execresult
UPDATE hr_time_entries
SET project_id = ?, task_id = ?, work_type_id = ?, work_mode = ?,
    duration_minutes = ?, description = ?, is_billable = ?
WHERE id = ? AND owner_id = ? AND business_id = ?;

-- name: DeleteWorkDayLine :execresult
DELETE FROM hr_time_entries
WHERE id = ? AND owner_id = ? AND business_id = ?;

-- The day's logged total, used for the clocked-versus-logged gap. Summed in SQL
-- rather than in Go so a paginated read cannot report a partial total.
-- name: SumWorkDayLineMinutes :one
SELECT COALESCE(SUM(duration_minutes), 0) FROM hr_time_entries
WHERE owner_id = ? AND business_id = ? AND work_day_id = ?;

-- ── Work types ─────────────────────────────────────────────────────────────

-- name: ListWorkTypes :many
SELECT id, parent_id, name, description, hr_category_id,
         sort_order, archived
FROM flow_work_types
WHERE owner_id = ? AND business_id = ?
  AND (archived = FALSE OR ? = TRUE)
ORDER BY sort_order ASC, name ASC;

-- name: GetWorkType :one
SELECT id, parent_id, name, description, hr_category_id,
         sort_order, archived
FROM flow_work_types
WHERE id = ? AND owner_id = ? AND business_id = ?
LIMIT 1;

-- name: CreateWorkType :execresult
INSERT INTO flow_work_types (
    owner_id, business_id, parent_id, name, description, hr_category_id, sort_order
) VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: UpdateWorkType :execresult
UPDATE flow_work_types
SET name = ?, description = ?, parent_id = ?, hr_category_id = ?, sort_order = ?, archived = ?
WHERE id = ? AND owner_id = ? AND business_id = ?;

-- name: DeleteWorkType :execresult
DELETE FROM flow_work_types
WHERE id = ? AND owner_id = ? AND business_id = ?;

-- Guards the two-level rule: a node may only be nested under a top-level type.
-- name: CountWorkTypeChildren :one
SELECT COUNT(*) FROM flow_work_types
WHERE owner_id = ? AND business_id = ? AND parent_id = ?;

-- The only link between a workspace member (a user id) and a work day (keyed
-- on employee id). One employee may in principle share a user_id across
-- multiple business rows in other schemas, so business_id stays in the filter
-- even though this table's own scope is already owner_id-bound.
-- name: GetEmployeeIDsByUserIDs :many
SELECT id, user_id FROM hr_employees
WHERE owner_id = ? AND business_id = ? AND user_id IN (sqlc.slice('user_ids'));

-- name: ListWorkDaysByEmployeeIDs :many
SELECT id, employee_id, date, timezone, work_mode, start_time,
         end_time, break_minutes, status, logged_by, approved_by,
         approved_at, reopened_by, reopened_at, created_at, updated_at
FROM flow_work_days
WHERE owner_id = ? AND business_id = ?
  AND employee_id IN (sqlc.slice('employee_ids'))
  AND date >= ? AND date <= ?
ORDER BY employee_id, date DESC
LIMIT ? OFFSET ?;

-- name: CountWorkDaysByEmployeeIDs :one
SELECT COUNT(*) FROM flow_work_days
WHERE owner_id = ? AND business_id = ?
  AND employee_id IN (sqlc.slice('employee_ids'))
  AND date >= ? AND date <= ?;
