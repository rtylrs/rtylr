-- name: GetActiveBusinessOwnerID :one
SELECT owner_id FROM businesses WHERE id = ? AND is_active = TRUE LIMIT 1;

-- name: GetUserDefaultBusinessID :one
SELECT business_id FROM users WHERE id = ? LIMIT 1;

-- name: GetOwnedActiveBusinessByID :one
SELECT id FROM businesses
WHERE id = ? AND owner_id = ? AND COALESCE(is_active, TRUE) = TRUE
LIMIT 1;

-- name: GetFirstOwnedActiveBusiness :one
SELECT id FROM businesses
WHERE owner_id = ? AND COALESCE(is_active, TRUE) = TRUE
ORDER BY created_at ASC
LIMIT 1;
