-- name: ListEmployees :many
SELECT id, owner_id, user_id, department_id, location_id, work_schedule_id,
       first_name, last_name, email, phone, employment_type, status,
       salary, hourly_rate, overtime_rate, start_date, end_date,
       address, emergency_contact_name, emergency_contact_phone, notes,
       clock_in_pin_hash, created_at, updated_at
FROM hr_employees
WHERE owner_id = ?
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: CountEmployees :one
SELECT COUNT(*) AS total FROM hr_employees WHERE owner_id = ?;

-- name: GetEmployee :one
SELECT id, owner_id, user_id, department_id, location_id, work_schedule_id,
       first_name, last_name, email, phone, employment_type, status,
       salary, hourly_rate, overtime_rate, start_date, end_date,
       address, emergency_contact_name, emergency_contact_phone, notes,
       clock_in_pin_hash, created_at, updated_at
FROM hr_employees
WHERE owner_id = ? AND id = ?;

-- name: CreateEmployee :execresult
INSERT INTO hr_employees (
    owner_id, user_id, department_id, location_id, work_schedule_id,
    first_name, last_name, email, phone, employment_type, status,
    salary, hourly_rate, overtime_rate, start_date, end_date,
    address, emergency_contact_name, emergency_contact_phone, notes
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: UpdateEmployee :exec
UPDATE hr_employees
SET department_id = ?, location_id = ?, work_schedule_id = ?,
    first_name = ?, last_name = ?, email = ?, phone = ?,
    employment_type = ?, status = ?, salary = ?, hourly_rate = ?,
    overtime_rate = ?, start_date = ?, end_date = ?,
    address = ?, emergency_contact_name = ?, emergency_contact_phone = ?,
    notes = ?, updated_at = NOW()
WHERE id = ? AND owner_id = ?;

-- name: DeleteEmployee :exec
UPDATE hr_employees SET status = 'terminated', updated_at = NOW()
WHERE id = ? AND owner_id = ?;

-- name: ActivateEmployee :exec
UPDATE hr_employees SET status = 'active', user_id = ?, updated_at = NOW()
WHERE id = ? AND owner_id = ?;

-- name: CreateLeaveBalance :exec
INSERT INTO hr_leave_balances (owner_id, employee_id, leave_type_id, year)
VALUES (?, ?, ?, ?);
