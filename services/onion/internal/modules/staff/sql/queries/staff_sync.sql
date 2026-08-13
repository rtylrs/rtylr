-- name: ListStaffWithDepartment :many
SELECT s.id, s.owner_id, s.business_id, s.user_id, s.department_id,
       COALESCE(d.name, '') AS department_name,
       s.first_name, s.last_name, s.email, s.phone,
       s.role, s.employment_type, s.status, s.salary, s.hourly_rate, s.overtime_rate,
       s.start_date, s.end_date, s.address, s.emergency_contact_name, s.emergency_contact_phone,
       s.apps, s.notes, s.created_at, s.updated_at
FROM hr_staff s
LEFT JOIN hr_departments d ON d.id = s.department_id AND d.business_id = s.business_id
WHERE s.business_id = ?
  AND (? = '' OR s.department_id = ?)
  AND (? = '' OR s.status = ?)
ORDER BY s.last_name ASC, s.first_name ASC;

-- name: GetStaffEmployeeSyncRow :one
SELECT owner_id, business_id, id, employee_id, user_id, department_id, first_name, last_name,
       email, phone, employment_type, status, salary, hourly_rate, overtime_rate,
       start_date, end_date, address, emergency_contact_name, emergency_contact_phone, notes
FROM hr_staff
WHERE business_id = ? AND id = ?;

-- cross-schema read from hr_core — find linked employee by user
-- name: FindEmployeeIDByUser :one
SELECT id FROM hr_employees
WHERE business_id = ? AND user_id = ?
ORDER BY id ASC
LIMIT 1;

-- cross-schema read from hr_core — find linked employee by email
-- name: FindEmployeeIDByEmail :one
SELECT id FROM hr_employees
WHERE business_id = ? AND email = ?
ORDER BY id ASC
LIMIT 1;

-- cross-schema write to hr_core — create employee from staff record
-- name: InsertEmployeeFromStaff :execresult
INSERT INTO hr_employees
    (owner_id, business_id, user_id, department_id, first_name, last_name, email, phone,
     employment_type, status, salary, hourly_rate, overtime_rate, start_date,
     end_date, address, emergency_contact_name, emergency_contact_phone, notes)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- cross-schema write to hr_core — sync employee fields from staff
-- name: UpdateEmployeeFromStaff :exec
UPDATE hr_employees
SET user_id = COALESCE(?, user_id),
    department_id = ?, first_name = ?, last_name = ?, email = ?, phone = ?,
    employment_type = ?, status = ?, salary = ?, hourly_rate = ?,
    overtime_rate = ?, start_date = ?, end_date = ?, address = ?,
    emergency_contact_name = ?, emergency_contact_phone = ?, notes = ?,
    updated_at = NOW()
WHERE id = ? AND business_id = ?;

-- name: LinkStaffEmployee :exec
UPDATE hr_staff SET employee_id = ? WHERE id = ? AND business_id = ?;
