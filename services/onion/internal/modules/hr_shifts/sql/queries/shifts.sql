-- name: ListShifts :many
SELECT id, owner_id, business_id, location_id, name, start_time, end_time, color, created_at, updated_at
FROM hr_shifts
WHERE business_id = ?
ORDER BY name ASC
LIMIT ? OFFSET ?;

-- name: CountShifts :one
SELECT COUNT(*) AS total FROM hr_shifts WHERE business_id = ?;

-- name: GetShift :one
SELECT id, owner_id, business_id, location_id, name, start_time, end_time, color, created_at, updated_at
FROM hr_shifts
WHERE business_id = ? AND id = ?;

-- name: CreateShift :execresult
INSERT INTO hr_shifts (owner_id, business_id, location_id, name, start_time, end_time, color)
VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: UpdateShift :exec
UPDATE hr_shifts
SET location_id = ?, name = ?, start_time = ?, end_time = ?, color = ?, updated_at = NOW()
WHERE id = ? AND business_id = ?;

-- name: ListShiftsByLocation :many
SELECT id, owner_id, business_id, location_id, name, start_time, end_time, color, created_at, updated_at
FROM hr_shifts
WHERE business_id = ? AND location_id = ?
ORDER BY start_time ASC;
