-- name: GetSelfHostLicenseByID :one
SELECT id, owner_email, customer_name, plan, features, seat_limit, expires_at, grace_until,
       status, license_key, key_hash, created_by, updated_by, created_at, updated_at,
       regenerated_at, revoked_at
FROM self_host_licenses
WHERE id = ?
ORDER BY id ASC
LIMIT 1;

-- name: GetSelfHostLicenseByOwnerEmail :one
SELECT id, owner_email, customer_name, plan, features, seat_limit, expires_at, grace_until,
       status, license_key, key_hash, created_by, updated_by, created_at, updated_at,
       regenerated_at, revoked_at
FROM self_host_licenses
WHERE owner_email = ?
ORDER BY owner_email ASC
LIMIT 1;

-- name: CreateSelfHostLicense :execresult
INSERT INTO self_host_licenses (
    owner_email, customer_name, plan, features, seat_limit, expires_at, grace_until,
    license_key, key_hash, created_by, updated_by
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: RegenerateSelfHostLicenseKey :execrows
UPDATE self_host_licenses
SET license_key = ?, key_hash = ?, status = 'active', updated_by = ?,
    regenerated_at = CURRENT_TIMESTAMP, revoked_at = NULL
WHERE id = ? AND owner_email = ?;

-- name: RevokeSelfHostLicense :execrows
UPDATE self_host_licenses
SET status = 'revoked', updated_by = ?, revoked_at = CURRENT_TIMESTAMP
WHERE id = ? AND owner_email = ?;
