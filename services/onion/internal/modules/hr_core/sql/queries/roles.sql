-- name: ListRoles :many
SELECT id, owner_id, business_id, name, description, permissions, is_system, created_at, updated_at
FROM hr_roles
WHERE business_id = ?
ORDER BY is_system DESC, name ASC;

-- name: GetRole :one
SELECT id, owner_id, business_id, name, description, permissions, is_system, created_at, updated_at
FROM hr_roles
WHERE business_id = ? AND id = ?;

-- name: CreateRole :execresult
INSERT INTO hr_roles (owner_id, business_id, name, description, permissions, is_system)
VALUES (?, ?, ?, ?, ?, ?);

-- name: UpsertRole :execresult
INSERT INTO hr_roles (owner_id, business_id, name, description, permissions, is_system)
VALUES (?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE description = VALUES(description), permissions = VALUES(permissions), updated_at = NOW();

-- name: UpdateRole :exec
UPDATE hr_roles
SET name = ?, description = ?, permissions = ?, updated_at = NOW()
WHERE id = ? AND business_id = ?;

-- name: AssignRole :exec
INSERT INTO hr_employee_roles (owner_id, business_id, employee_id, role_id)
VALUES (?, ?, ?, ?)
ON DUPLICATE KEY UPDATE role_id = VALUES(role_id), updated_at = NOW();

-- name: GetEmployeeRoles :many
SELECT r.id, r.owner_id, r.business_id, r.name, r.description, r.permissions, r.is_system, r.created_at, r.updated_at
FROM hr_roles r
JOIN hr_employee_roles er ON er.role_id = r.id AND er.business_id = r.business_id
WHERE er.business_id = ? AND er.employee_id = ?;
