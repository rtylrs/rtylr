-- name: ListStaffAvailability :many
SELECT id, owner_id, business_id, staff_id, day_of_week, specific_date, start_time, end_time, is_available, created_at, updated_at, day_key, date_key
FROM hr_staff_availability
WHERE business_id = ?
  AND (sqlc.narg(staff_id) IS NULL OR staff_id = sqlc.narg(staff_id))
  AND (sqlc.narg(day_of_week) IS NULL OR day_of_week = sqlc.narg(day_of_week))
  AND (sqlc.narg(date_from) IS NULL OR specific_date IS NULL OR specific_date >= sqlc.narg(date_from))
  AND (sqlc.narg(date_to) IS NULL OR specific_date IS NULL OR specific_date <= sqlc.narg(date_to))
ORDER BY staff_id ASC, day_of_week ASC, specific_date ASC, start_time ASC
LIMIT ? OFFSET ?;

-- name: CountStaffAvailability :one
SELECT COUNT(*)
FROM hr_staff_availability
WHERE business_id = ?
  AND (sqlc.narg(staff_id) IS NULL OR staff_id = sqlc.narg(staff_id))
  AND (sqlc.narg(day_of_week) IS NULL OR day_of_week = sqlc.narg(day_of_week))
  AND (sqlc.narg(date_from) IS NULL OR specific_date IS NULL OR specific_date >= sqlc.narg(date_from))
  AND (sqlc.narg(date_to) IS NULL OR specific_date IS NULL OR specific_date <= sqlc.narg(date_to));

-- name: GetStaffAvailability :one
SELECT id, owner_id, business_id, staff_id, day_of_week, specific_date, start_time, end_time, is_available, created_at, updated_at, day_key, date_key
FROM hr_staff_availability
WHERE business_id = ? AND id = ?;

-- name: GetStaffAvailabilityByNaturalKey :one
SELECT id, owner_id, business_id, staff_id, day_of_week, specific_date, start_time, end_time, is_available, created_at, updated_at, day_key, date_key
FROM hr_staff_availability
WHERE business_id = ?
  AND staff_id = ?
  AND IFNULL(day_of_week, -1) = IFNULL(sqlc.narg(day_of_week), -1)
  AND IFNULL(specific_date, '1000-01-01') = IFNULL(sqlc.narg(specific_date), '1000-01-01')
  AND start_time = ?;

-- name: CreateStaffAvailability :execresult
INSERT INTO hr_staff_availability (owner_id, business_id, staff_id, day_of_week, specific_date, start_time, end_time, is_available)
VALUES (?, ?, ?, ?, ?, ?, ?, ?);

-- name: UpdateStaffAvailability :exec
UPDATE hr_staff_availability
SET day_of_week = ?, specific_date = ?, start_time = ?, end_time = ?, is_available = ?, updated_at = NOW()
WHERE business_id = ? AND id = ?;

-- name: DeleteStaffAvailability :execresult
DELETE FROM hr_staff_availability
WHERE business_id = ? AND id = ?;

-- name: ListStaffAvailabilityForDate :many
SELECT id, owner_id, business_id, staff_id, day_of_week, specific_date, start_time, end_time, is_available, created_at, updated_at, day_key, date_key
FROM hr_staff_availability
WHERE business_id = ?
  AND staff_id = ?
  AND (specific_date = sqlc.arg(on_date) OR (specific_date IS NULL AND day_of_week = sqlc.arg(day_of_week)))
ORDER BY specific_date IS NULL ASC, start_time ASC
LIMIT ? OFFSET ?;
