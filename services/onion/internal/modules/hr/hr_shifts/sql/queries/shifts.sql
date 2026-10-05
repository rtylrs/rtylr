-- hr_shifts holds two kinds of row, told apart by whether staff_id is set:
-- template rows (location_id, name, color; staff_id NULL) and attendance rows
-- (staff_id, date, clock/hours columns; location_id and name NULL). Every query
-- below selects the full column set so one row struct covers both kinds.

-- name: ListShifts :many
SELECT id, owner_id, business_id, staff_id, department_id, date, start_time, end_time,
       break_duration_minutes, status, clock_in_at, clock_out_at,
       hours_worked, overtime_hours, notes, location_id, name, color,
       created_at, updated_at
FROM hr_shifts
WHERE business_id = ?
  AND (sqlc.narg('staff_id') IS NULL OR staff_id = sqlc.narg('staff_id'))
  AND (sqlc.narg('date') IS NULL OR date = sqlc.narg('date'))
  AND (sqlc.narg('location_id') IS NULL OR location_id = sqlc.narg('location_id'))
  AND (sqlc.narg('kind') IS NULL
       OR (sqlc.narg('kind') = 'template' AND staff_id IS NULL)
       OR (sqlc.narg('kind') = 'attendance' AND staff_id IS NOT NULL))
ORDER BY date DESC, start_time ASC, name ASC
LIMIT ? OFFSET ?;

-- name: CountShifts :one
SELECT COUNT(*) AS total
FROM hr_shifts
WHERE business_id = ?
  AND (sqlc.narg('staff_id') IS NULL OR staff_id = sqlc.narg('staff_id'))
  AND (sqlc.narg('date') IS NULL OR date = sqlc.narg('date'))
  AND (sqlc.narg('location_id') IS NULL OR location_id = sqlc.narg('location_id'))
  AND (sqlc.narg('kind') IS NULL
       OR (sqlc.narg('kind') = 'template' AND staff_id IS NULL)
       OR (sqlc.narg('kind') = 'attendance' AND staff_id IS NOT NULL));

-- name: GetShift :one
SELECT id, owner_id, business_id, staff_id, department_id, date, start_time, end_time,
       break_duration_minutes, status, clock_in_at, clock_out_at,
       hours_worked, overtime_hours, notes, location_id, name, color,
       created_at, updated_at
FROM hr_shifts
WHERE business_id = ? AND id = ?;

-- GetShiftForUpdate row-locks the target before an update or delete so a
-- concurrent write to the same shift serialises behind it.
-- name: GetShiftForUpdate :one
SELECT id, owner_id, business_id, staff_id, department_id, date, start_time, end_time,
       break_duration_minutes, status, clock_in_at, clock_out_at,
       hours_worked, overtime_hours, notes, location_id, name, color,
       created_at, updated_at
FROM hr_shifts
WHERE business_id = ? AND id = ?
FOR UPDATE;

-- name: CreateShift :execresult
INSERT INTO hr_shifts (
    owner_id, business_id, staff_id, department_id, date, start_time, end_time,
    break_duration_minutes, status, notes, location_id, name, color
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: UpdateShift :exec
UPDATE hr_shifts
SET staff_id = ?, department_id = ?, date = ?, start_time = ?, end_time = ?,
    break_duration_minutes = ?, status = ?, notes = ?,
    location_id = ?, name = ?, color = ?, updated_at = NOW()
WHERE id = ? AND business_id = ?;

-- name: DeleteShift :exec
DELETE FROM hr_shifts WHERE id = ? AND business_id = ?;

-- name: ListShiftsByLocation :many
SELECT id, owner_id, business_id, staff_id, department_id, date, start_time, end_time,
       break_duration_minutes, status, clock_in_at, clock_out_at,
       hours_worked, overtime_hours, notes, location_id, name, color,
       created_at, updated_at
FROM hr_shifts
WHERE business_id = ? AND location_id = ?
ORDER BY start_time ASC
LIMIT ? OFFSET ?;
