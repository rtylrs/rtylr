-- name: ListRoles :many
SELECT id, owner_id, business_id, name, permissions, is_system, created_at, updated_at
FROM hub_roles
WHERE business_id = ?
ORDER BY is_system DESC, name ASC, id ASC;

-- name: CreateRole :execresult
INSERT INTO hub_roles (owner_id, business_id, name, permissions) VALUES (?, ?, ?, ?);

-- name: UpdateRole :execresult
UPDATE hub_roles
SET name = ?, permissions = ?, updated_at = NOW()
WHERE id = ? AND business_id = ? AND is_system = FALSE;

-- name: DeleteRole :execresult
DELETE FROM hub_roles WHERE id = ? AND business_id = ? AND is_system = FALSE;
