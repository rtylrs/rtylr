-- name: GetBusinessByID :one
SELECT id, name, owner_id, created_at
FROM businesses
WHERE id = ?
ORDER BY id ASC
LIMIT 1;

-- name: ListActiveBusinessesByOwner :many
SELECT b.id, b.name, b.owner_id, b.created_at
FROM businesses b
LEFT JOIN users u ON u.id = b.owner_id
WHERE b.owner_id = ? AND COALESCE(b.is_active, TRUE) = TRUE
ORDER BY (b.id = u.business_id) DESC, b.id ASC
LIMIT 100;

-- name: UpdateBusiness :exec
UPDATE businesses SET name = ? WHERE id = ?;

-- name: DeleteBusiness :exec
DELETE FROM businesses WHERE id = ?;
