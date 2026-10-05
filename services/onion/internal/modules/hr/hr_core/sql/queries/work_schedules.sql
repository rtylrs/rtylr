-- name: ListWorkSchedules :many
SELECT id, owner_id, business_id, name, schedule_data, is_default, created_at, updated_at
FROM hr_work_schedules
WHERE business_id = ?
ORDER BY name ASC
LIMIT ? OFFSET ?;

-- name: CountWorkSchedules :one
SELECT COUNT(*) FROM hr_work_schedules WHERE business_id = ?;

-- name: GetWorkScheduleForEmployee :one
SELECT ws.id, ws.owner_id, ws.business_id, ws.name, ws.schedule_data, ws.is_default, ws.created_at, ws.updated_at
FROM hr_work_schedules ws
JOIN hr_employees e ON e.work_schedule_id = ws.id AND e.business_id = ws.business_id
WHERE ws.business_id = ? AND e.id = ?;

-- name: GetWorkSchedule :one
SELECT id, owner_id, business_id, name, schedule_data, is_default, created_at, updated_at
FROM hr_work_schedules
WHERE business_id = ? AND id = ?;

-- name: CreateWorkSchedule :execresult
INSERT INTO hr_work_schedules (owner_id, business_id, name, schedule_data, is_default)
VALUES (?, ?, ?, ?, ?);

-- name: UpdateWorkSchedule :execresult
UPDATE hr_work_schedules
SET name = ?, schedule_data = ?, is_default = ?, updated_at = NOW()
WHERE id = ? AND business_id = ?;
