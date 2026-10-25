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

-- name: LinkStaffEmployeeForUser :exec
UPDATE hr_staff s
JOIN hr_employees e
  ON e.business_id = s.business_id
 AND e.user_id = s.user_id
SET s.employee_id = e.id
WHERE s.business_id = ?
  AND s.user_id = ?
  AND (s.employee_id IS NULL OR s.employee_id = 0);

-- name: EnsureStaffForUser :exec
INSERT INTO hr_staff
    (owner_id, business_id, user_id, employee_id, department_id, first_name, last_name, email,
     phone, role, employment_type, status, salary, hourly_rate, overtime_rate, start_date,
     end_date, address, emergency_contact_name, emergency_contact_phone, apps, notes, photo_url, bio)
SELECT e.business_id, e.business_id, e.user_id, e.id, e.department_id, e.first_name, e.last_name, e.email,
       e.phone, ?,
       CASE e.employment_type
           WHEN 'part_time' THEN 'part_time'
           WHEN 'contractor' THEN 'contractor'
           WHEN 'intern' THEN 'intern'
           ELSE 'full_time'
       END,
       'active', e.salary, e.hourly_rate, e.overtime_rate, COALESCE(e.start_date, CURDATE()),
       e.end_date, e.address, e.emergency_contact_name, e.emergency_contact_phone,
       JSON_ARRAY(), e.notes, NULL, NULL
FROM hr_employees e
WHERE e.business_id = ?
  AND e.user_id = ?
  AND NOT EXISTS (
      SELECT 1
      FROM hr_staff s
      WHERE s.business_id = e.business_id
        AND s.user_id = e.user_id
  );

-- name: DeleteStaffForUser :exec
DELETE FROM hr_staff
WHERE business_id = ? AND user_id = ?;
