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
