-- name: ListRoles :many
SELECT id, owner_id, name, permissions, is_system, created_at, updated_at
FROM hub_roles
WHERE owner_id = ?
ORDER BY is_system DESC, name ASC;

-- name: CreateRole :execresult
INSERT INTO hub_roles (owner_id, name, permissions) VALUES (?, ?, ?);

-- name: UpdateRole :exec
UPDATE hub_roles
SET name = ?, permissions = ?, updated_at = NOW()
WHERE id = ? AND owner_id = ? AND is_system = FALSE;

-- name: DeleteRole :exec
DELETE FROM hub_roles WHERE id = ? AND owner_id = ? AND is_system = FALSE;
