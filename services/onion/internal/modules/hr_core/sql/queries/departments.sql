-- name: ListHRDepartments :many
SELECT id, owner_id, location_id, parent_department_id, name, default_work_schedule_id, created_at, updated_at
FROM hr_departments
WHERE owner_id = ?
ORDER BY name ASC;

-- name: CreateHRDepartment :execresult
INSERT INTO hr_departments (owner_id, location_id, parent_department_id, name, default_work_schedule_id)
VALUES (?, ?, ?, ?, ?);

-- name: UpdateHRDepartment :exec
UPDATE hr_departments
SET location_id = ?, parent_department_id = ?, name = ?, default_work_schedule_id = ?, updated_at = NOW()
WHERE id = ? AND owner_id = ?;
