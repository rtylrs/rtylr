-- name: ListDepartments :many
SELECT id, owner_id, business_id, name, description, manager_staff_id, created_at, updated_at
FROM hr_departments
WHERE business_id = ?
ORDER BY name ASC
LIMIT 100;

-- name: CreateDepartment :execresult
INSERT INTO hr_departments (owner_id, business_id, name, description)
VALUES (?, ?, ?, ?);

-- name: UpdateDepartment :exec
UPDATE hr_departments
SET name = ?, description = ?, updated_at = NOW()
WHERE id = ? AND business_id = ?;

-- name: DeleteDepartment :exec
DELETE FROM hr_departments WHERE id = ? AND business_id = ?;
