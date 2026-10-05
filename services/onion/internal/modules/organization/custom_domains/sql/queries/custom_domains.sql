-- name: GetCustomDomain :one
SELECT id, owner_id, business_id, domain, verification_token, status, ssl_status, verified_at, created_at, updated_at
FROM custom_domains WHERE business_id = ? LIMIT 1;

-- name: GetCustomDomainByDomain :one
SELECT id, owner_id, business_id, domain, verification_token, status, ssl_status, verified_at, created_at, updated_at
FROM custom_domains WHERE domain = ? LIMIT 1;

-- name: RegisterCustomDomain :execresult
INSERT INTO custom_domains (owner_id, business_id, domain, verification_token)
VALUES (?, ?, ?, ?);

-- name: RotateCustomDomainChallenge :exec
UPDATE custom_domains SET verification_token = ?, status = 'pending_verification', ssl_status = 'pending', verified_at = NULL
WHERE business_id = ? AND domain = ?;

-- name: VerifyCustomDomain :execresult
UPDATE custom_domains SET status = 'verified', ssl_status = 'pending', verified_at = CURRENT_TIMESTAMP
WHERE business_id = ? AND verification_token = ?;

-- name: RemoveCustomDomain :exec
DELETE FROM custom_domains WHERE business_id = ?;
