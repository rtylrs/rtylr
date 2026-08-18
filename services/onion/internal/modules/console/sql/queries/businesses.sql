-- name: GetBusinessByID :one
SELECT id, name, owner_id, created_at
FROM businesses
WHERE id = ?
ORDER BY id ASC
LIMIT 1;

-- name: GetActiveBusinessByOwner :one
SELECT id, name, owner_id, created_at
FROM businesses
WHERE owner_id = ? AND COALESCE(is_active, TRUE) = TRUE
ORDER BY created_at ASC
LIMIT 1;

-- name: UpdateBusiness :exec
UPDATE businesses SET name = ? WHERE id = ?;

-- name: DeleteBusiness :exec
DELETE FROM businesses WHERE id = ?;
