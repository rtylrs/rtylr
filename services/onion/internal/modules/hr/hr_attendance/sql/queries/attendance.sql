-- name: CreateAttendanceRecord :execresult
INSERT INTO hr_attendance_records (
    owner_id, business_id, employee_id, location_id, date, clock_in_at,
    clock_in_method, clock_in_lat, clock_in_lng, is_late, late_minutes, status
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: UpdateAttendanceRecord :exec
UPDATE hr_attendance_records
SET clock_out_at = ?, duration_minutes = ?, status = ?, updated_at = NOW()
WHERE id = ? AND business_id = ?;

-- name: GetActiveClockIn :one
SELECT id, owner_id, business_id, employee_id, location_id, date, clock_in_at, clock_out_at,
       duration_minutes, clock_in_method, clock_in_lat, clock_in_lng,
       is_late, late_minutes, status, created_at, updated_at
FROM hr_attendance_records
WHERE business_id = ? AND employee_id = ? AND clock_out_at IS NULL
ORDER BY clock_in_at DESC
LIMIT 1;

-- name: ListAttendance :many
SELECT id, owner_id, business_id, employee_id, location_id, date, clock_in_at, clock_out_at,
       duration_minutes, clock_in_method, clock_in_lat, clock_in_lng,
       is_late, late_minutes, status, created_at, updated_at
FROM hr_attendance_records
WHERE business_id = ? AND employee_id = ? AND date >= ? AND date <= ?
ORDER BY date DESC, clock_in_at DESC
LIMIT ? OFFSET ?;

-- name: CountAttendance :one
SELECT COUNT(*) AS total FROM hr_attendance_records
WHERE business_id = ? AND employee_id = ? AND date >= ? AND date <= ?;

-- name: CountAllAttendance :one
SELECT COUNT(*) AS total FROM hr_attendance_records
WHERE business_id = ? AND date >= ? AND date <= ?;

-- name: ListAllAttendance :many
SELECT id, owner_id, business_id, employee_id, location_id, date, clock_in_at, clock_out_at,
       duration_minutes, clock_in_method, clock_in_lat, clock_in_lng,
       is_late, late_minutes, status, created_at, updated_at
FROM hr_attendance_records
WHERE business_id = ? AND date >= ? AND date <= ?
ORDER BY date DESC, clock_in_at DESC
LIMIT ? OFFSET ?;

-- name: GetDailyAttendance :many
SELECT id, owner_id, business_id, employee_id, location_id, date, clock_in_at, clock_out_at,
       duration_minutes, clock_in_method, clock_in_lat, clock_in_lng,
       is_late, late_minutes, status, created_at, updated_at
FROM hr_attendance_records
WHERE business_id = ? AND location_id = ? AND date = ?
ORDER BY clock_in_at ASC
LIMIT ? OFFSET ?;

-- name: GetAttendanceConfig :one
SELECT id, owner_id, business_id, app_enabled, qr_enabled, gps_enabled, pin_enabled,
       facial_enabled, geofence_radius_m, default_location_lat, default_location_lng
FROM hr_attendance_config
WHERE business_id = ?;

-- cross-schema read from hr_core — verify the employee's attendance PIN.
-- name: GetEmployeePINHash :one
SELECT clock_in_pin_hash
FROM hr_employees
WHERE id = ? AND business_id = ?;

-- name: SetEmployeePINHash :execresult
UPDATE hr_employees
SET clock_in_pin_hash = ?, updated_at = NOW()
WHERE id = ? AND business_id = ?;

-- cross-schema read from hr_core — resolve the attendance location's geofence coordinates.
-- name: GetLocationCoordinates :one
SELECT latitude, longitude
FROM hr_locations
WHERE id = ? AND business_id = ?;
