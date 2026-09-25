-- The generated list contract has no pagination fields; keep the endpoint
-- bounded until the corresponding voxie contract can expose a page cursor.
-- name: ListHRDepartments :many
SELECT id, owner_id, business_id, location_id, parent_department_id, name, default_work_schedule_id, created_at, updated_at
FROM hr_departments
WHERE business_id = ?
  AND (sqlc.arg('location_id') = 0 OR location_id = sqlc.arg('location_id'))
ORDER BY name ASC
LIMIT 500;

-- name: GetHRDepartment :one
SELECT id, owner_id, business_id, location_id, parent_department_id, name, default_work_schedule_id, created_at, updated_at
FROM hr_departments
WHERE business_id = ? AND id = ?;

-- name: CreateHRDepartment :execresult
INSERT INTO hr_departments (owner_id, business_id, location_id, parent_department_id, name, default_work_schedule_id)
VALUES (?, ?, ?, ?, ?, ?);

-- name: UpdateHRDepartment :execresult
UPDATE hr_departments
SET location_id = ?, parent_department_id = ?, name = ?, default_work_schedule_id = ?, updated_at = NOW()
WHERE id = ? AND business_id = ?;
