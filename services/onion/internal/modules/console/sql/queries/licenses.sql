-- name: GetSelfHostLicenseByKey :one
SELECT id, license_key, owner_id, status, created_at
FROM self_host_licenses
WHERE license_key = ?
LIMIT 1;
