-- name: ListDepartments :many
SELECT id, owner_id, name, description, manager_staff_id, created_at, updated_at
FROM hr_departments
WHERE owner_id = ?
ORDER BY name ASC;

-- name: CreateDepartment :execresult
INSERT INTO hr_departments (owner_id, name, description)
VALUES (?, ?, ?);

-- name: UpdateDepartment :exec
UPDATE hr_departments
SET name = ?, description = ?, updated_at = NOW()
WHERE id = ? AND owner_id = ?;

-- name: DeleteDepartment :exec
DELETE FROM hr_departments WHERE id = ? AND owner_id = ?;
