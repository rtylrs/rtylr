-- name: ListCategories :many
SELECT id, owner_id, business_id, name, description, created_at, updated_at
FROM erp_categories
WHERE business_id = ?
ORDER BY name ASC;

-- name: CreateCategory :execresult
INSERT INTO erp_categories (owner_id, business_id, name, description)
VALUES (?, ?, ?, ?);

-- name: UpdateCategory :exec
UPDATE erp_categories
SET name = ?, description = ?
WHERE business_id = ? AND id = ?;

-- name: DeleteCategory :exec
DELETE FROM erp_categories WHERE business_id = ? AND id = ?;

-- name: ListCategoriesPaginated :many
SELECT id, owner_id, business_id, name, description, created_at, updated_at
FROM erp_categories
WHERE business_id = ?
ORDER BY name ASC
LIMIT ? OFFSET ?;
