-- Flow owns its work-day workflow, while HR Time owns every persisted time
-- entry. These narrow operations keep the Flow-source slice inside its owner.

-- name: ListFlowWorkdayEntries :many
SELECT id, work_day_id, project_id, task_id, work_type_id,
       work_mode, description, duration_minutes, is_billable, status,
       client_key, logged_by, created_at, updated_at
FROM hr_time_entries
WHERE owner_id = ? AND business_id = ? AND work_day_id = ? AND source = 'flow'
ORDER BY id ASC;

-- name: GetFlowWorkdayEntry :one
SELECT id, work_day_id, project_id, task_id, work_type_id,
       work_mode, description, duration_minutes, is_billable, status,
       client_key, logged_by, created_at, updated_at
FROM hr_time_entries
WHERE id = ? AND owner_id = ? AND business_id = ? AND source = 'flow'
LIMIT 1;

-- name: GetFlowWorkdayEntryOwner :one
SELECT work_day_id, employee_id, task_id
FROM hr_time_entries
WHERE id = ? AND owner_id = ? AND business_id = ? AND source = 'flow'
LIMIT 1;

-- The client's stable key makes offline retries idempotent. LAST_INSERT_ID on
-- the duplicate path returns the original row id without touching its data.
-- name: CreateFlowWorkdayEntry :execresult
INSERT INTO hr_time_entries (
    owner_id, business_id, employee_id, work_day_id, date,
    project_id, task_id, work_type_id, work_mode, timezone,
    duration_minutes, description, is_billable, client_key, logged_by, source
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'flow')
ON DUPLICATE KEY UPDATE id = LAST_INSERT_ID(id);

-- name: UpdateFlowWorkdayEntry :execresult
UPDATE hr_time_entries
SET project_id = ?, task_id = ?, work_type_id = ?, work_mode = ?,
    duration_minutes = ?, description = ?, is_billable = ?
WHERE id = ? AND owner_id = ? AND business_id = ? AND source = 'flow';

-- name: DeleteFlowWorkdayEntry :execresult
DELETE FROM hr_time_entries
WHERE id = ? AND owner_id = ? AND business_id = ? AND source = 'flow';

-- name: SumFlowWorkdayEntryMinutes :one
SELECT COALESCE(SUM(duration_minutes), 0)
FROM hr_time_entries
WHERE owner_id = ? AND business_id = ? AND work_day_id = ? AND source = 'flow';

-- name: SumFlowWorkdayEntryMinutesByDayIDs :many
SELECT work_day_id, COALESCE(SUM(duration_minutes), 0) AS logged_minutes
FROM hr_time_entries
WHERE business_id = ? AND source = 'flow'
  AND work_day_id IN (sqlc.slice('work_day_ids'))
GROUP BY work_day_id;

-- name: SumFlowWorkdayEntryMinutesByTaskIDs :many
SELECT task_id, CAST(COALESCE(SUM(duration_minutes), 0) AS SIGNED) AS logged_minutes
FROM hr_time_entries
WHERE business_id = ? AND source = 'flow' AND task_id IS NOT NULL
  AND task_id IN (sqlc.slice('task_ids'))
GROUP BY task_id;

-- name: SumFlowWorkdayEntryMinutesByMemberAndTaskIDs :many
SELECT employee_id, CAST(COALESCE(SUM(duration_minutes), 0) AS SIGNED) AS logged_minutes
FROM hr_time_entries
WHERE business_id = ? AND source = 'flow' AND task_id IS NOT NULL
  AND task_id IN (sqlc.slice('task_ids'))
GROUP BY employee_id;
