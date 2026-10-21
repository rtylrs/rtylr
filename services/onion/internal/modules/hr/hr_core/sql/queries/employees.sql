-- name: ListEmployees :many
SELECT id, owner_id, business_id, user_id, department_id, location_id, work_schedule_id,
       first_name, last_name, email, phone, employment_type, status,
       salary, hourly_rate, overtime_rate, start_date, end_date,
       address, emergency_contact_name, emergency_contact_phone, notes,
       clock_in_pin_hash, created_at, updated_at
FROM hr_employees
WHERE business_id = ?
  AND (sqlc.narg('department_id') IS NULL OR department_id = sqlc.narg('department_id'))
  AND (sqlc.narg('location_id') IS NULL OR location_id = sqlc.narg('location_id'))
  AND (sqlc.narg('status') IS NULL OR status = sqlc.narg('status'))
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: CountEmployees :one
SELECT COUNT(*) AS total
FROM hr_employees
WHERE business_id = ?
  AND (sqlc.narg('department_id') IS NULL OR department_id = sqlc.narg('department_id'))
  AND (sqlc.narg('location_id') IS NULL OR location_id = sqlc.narg('location_id'))
  AND (sqlc.narg('status') IS NULL OR status = sqlc.narg('status'));

-- name: GetEmployee :one
SELECT id, owner_id, business_id, user_id, department_id, location_id, work_schedule_id,
       first_name, last_name, email, phone, employment_type, status,
       salary, hourly_rate, overtime_rate, start_date, end_date,
       address, emergency_contact_name, emergency_contact_phone, notes,
       clock_in_pin_hash, created_at, updated_at
FROM hr_employees
WHERE business_id = ? AND id = ?;

-- name: CreateEmployee :execresult
INSERT INTO hr_employees (
    owner_id, business_id, user_id, department_id, location_id, work_schedule_id,
    first_name, last_name, email, phone, employment_type, status,
    salary, hourly_rate, overtime_rate, start_date, end_date,
    address, emergency_contact_name, emergency_contact_phone, notes
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: EnsureEmployeeForUser :exec
INSERT INTO hr_employees (owner_id, business_id, user_id, first_name, last_name, email, status)
SELECT ?, ?, ?, ?, ?, ?, 'active'
WHERE NOT EXISTS (
    SELECT 1
    FROM hr_employees e
    WHERE e.user_id = ?
);

-- name: DeleteEmployeeForUser :exec
DELETE FROM hr_employees
WHERE owner_id = ? AND business_id = ? AND user_id = ?;

-- name: ListTerminatedEmployees :many
SELECT DISTINCT business_id, user_id
FROM hr_employees
WHERE status = 'terminated' AND user_id IS NOT NULL
ORDER BY business_id, user_id;

-- name: UpdateEmployee :execrows
UPDATE hr_employees
SET user_id = ?, department_id = ?, location_id = ?, work_schedule_id = ?,
    first_name = ?, last_name = ?, email = ?, phone = ?,
    employment_type = ?, status = ?, salary = ?, hourly_rate = ?,
    overtime_rate = ?, start_date = ?, end_date = ?,
    address = ?, emergency_contact_name = ?, emergency_contact_phone = ?,
    notes = ?, updated_at = NOW()
WHERE id = ? AND business_id = ?;

-- name: DeactivateEmployee :execresult
UPDATE hr_employees
SET status = 'terminated', end_date = ?, updated_at = NOW()
WHERE id = ? AND business_id = ?;

-- name: LinkInviteAcceptance :execresult
UPDATE hr_employees
SET user_id = ?, status = 'active', updated_at = NOW()
WHERE id = ? AND business_id = ? AND status = 'invited';

-- name: CreateLeaveBalance :exec
INSERT INTO hr_leave_balances (owner_id, business_id, employee_id, leave_type_id, year)
VALUES (?, ?, ?, ?, ?);

-- name: GetEmployeeIDByUserID :one
SELECT id FROM hr_employees WHERE business_id = ? AND user_id = ? ORDER BY id ASC LIMIT 1;
