-- name: ListShifts :many
SELECT id, owner_id, business_id, staff_id, department_id, date, start_time, end_time,
       break_duration_minutes, status, clock_in_at, clock_out_at,
       hours_worked, overtime_hours, notes, created_at, updated_at
FROM hr_shifts
WHERE business_id = ?
  AND staff_id IS NOT NULL
  AND (sqlc.narg('staff_id') IS NULL OR staff_id = sqlc.narg('staff_id'))
  AND (sqlc.narg('date') IS NULL OR date = sqlc.narg('date'))
ORDER BY date DESC, start_time ASC;

-- name: CreateShift :execresult
INSERT INTO hr_shifts (owner_id, business_id, staff_id, department_id, date, start_time, end_time, break_duration_minutes, notes)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: UpdateShift :exec
UPDATE hr_shifts
SET department_id = ?, date = ?, start_time = ?, end_time = ?, break_duration_minutes = ?,
    status = ?, notes = ?, updated_at = NOW()
WHERE id = ? AND business_id = ?;

-- name: DeleteShift :exec
DELETE FROM hr_shifts WHERE id = ? AND business_id = ?;
