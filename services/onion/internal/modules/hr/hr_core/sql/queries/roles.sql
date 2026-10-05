-- The generated list contract has no pagination fields; keep the endpoint
-- bounded until the corresponding voxie contract can expose a page cursor.
-- name: ListRoles :many
SELECT id, owner_id, business_id, name, description, permissions, is_system, created_at, updated_at
FROM hr_roles
WHERE business_id = ?
ORDER BY is_system DESC, name ASC
LIMIT 500;

-- name: GetRole :one
SELECT id, owner_id, business_id, name, description, permissions, is_system, created_at, updated_at
FROM hr_roles
WHERE business_id = ? AND id = ?;

-- name: CreateRole :execresult
INSERT INTO hr_roles (owner_id, business_id, name, description, permissions, is_system)
VALUES (?, ?, ?, ?, ?, ?);

-- name: InsertRoleIfMissing :execresult
INSERT IGNORE INTO hr_roles (owner_id, business_id, name, description, permissions, is_system)
VALUES (?, ?, ?, ?, ?, ?);

-- name: UpdateRole :execresult
UPDATE hr_roles
SET name = ?, description = ?, permissions = ?, updated_at = NOW()
WHERE id = ? AND business_id = ?;

-- name: AssignRole :exec
INSERT INTO hr_employee_roles (owner_id, business_id, employee_id, role_id)
VALUES (?, ?, ?, ?)
ON DUPLICATE KEY UPDATE role_id = VALUES(role_id), updated_at = NOW();

-- AssignDefaultOwnerRole gives a newly onboarded business owner the system HR
-- Admin role after their employee row has been created. INSERT IGNORE keeps an
-- existing custom assignment intact when provisioning is retried.
-- name: AssignDefaultOwnerRole :exec
INSERT IGNORE INTO hr_employee_roles (owner_id, business_id, employee_id, role_id)
SELECT e.owner_id, e.business_id, e.id, r.id
FROM hr_employees e
JOIN hr_roles r ON r.business_id = e.business_id
WHERE e.business_id = ?
  AND e.user_id = ?
  AND r.name = 'HR Admin'
  AND r.is_system = TRUE;

-- name: GetEmployeeRoles :many
SELECT r.id, r.owner_id, r.business_id, r.name, r.description, r.permissions, r.is_system, r.created_at, r.updated_at
FROM hr_roles r
JOIN hr_employee_roles er ON er.role_id = r.id AND er.business_id = r.business_id
WHERE er.business_id = ? AND er.employee_id = ?;

-- name: HasEmployeeRole :one
SELECT COUNT(*)
FROM hr_employee_roles
WHERE business_id = ? AND employee_id = ? AND role_id = ?;
