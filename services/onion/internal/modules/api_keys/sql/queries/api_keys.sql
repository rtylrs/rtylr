-- name: ListAPIKeys :many
SELECT id, owner_id, name, key_prefix, scopes, last_used_at, expires_at, is_active, created_at
FROM hub_api_keys
WHERE owner_id = ? AND is_active = TRUE
ORDER BY created_at DESC;

-- name: CreateAPIKey :execresult
INSERT INTO hub_api_keys (owner_id, name, key_hash, key_prefix, scopes)
VALUES (?, ?, ?, ?, ?);

-- name: RevokeAPIKey :exec
UPDATE hub_api_keys
SET is_active = FALSE
WHERE id = ? AND owner_id = ?;
