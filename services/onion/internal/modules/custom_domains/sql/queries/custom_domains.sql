-- name: GetCustomDomain :one
SELECT id, owner_id, business_id, domain, verification_token, status, ssl_status, verified_at, created_at, updated_at
FROM custom_domains WHERE business_id = ? LIMIT 1;

-- name: RegisterCustomDomain :execresult
INSERT INTO custom_domains (owner_id, business_id, domain, verification_token)
VALUES (?, ?, ?, ?)
ON DUPLICATE KEY UPDATE domain = VALUES(domain), verification_token = VALUES(verification_token), status = 'pending_verification', ssl_status = 'pending', verified_at = NULL;

-- name: VerifyCustomDomain :exec
UPDATE custom_domains SET status = 'verified', ssl_status = 'pending', verified_at = CURRENT_TIMESTAMP
WHERE business_id = ? AND verification_token = ?;

-- name: RemoveCustomDomain :exec
DELETE FROM custom_domains WHERE business_id = ?;
