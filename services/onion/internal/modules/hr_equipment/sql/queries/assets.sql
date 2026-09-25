-- name: CreateAsset :execresult
INSERT INTO hr_assets (owner_id, business_id, name, category, serial_number, location_id, status, `condition`, purchase_date, purchase_price, notes)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetAsset :one
SELECT id, owner_id, business_id, name, category, serial_number, location_id, assigned_to, status, `condition`,
       purchase_date, purchase_price, notes, created_at, updated_at
FROM hr_assets
WHERE business_id = ? AND id = ?;

-- name: ListAssets :many
SELECT id, owner_id, business_id, name, category, serial_number, location_id, assigned_to, status, `condition`,
       purchase_date, purchase_price, notes, created_at, updated_at
FROM hr_assets
WHERE business_id = ?
ORDER BY name ASC
LIMIT ? OFFSET ?;

-- name: ListAssetsFiltered :many
SELECT id, owner_id, business_id, name, category, serial_number, location_id, assigned_to, status, `condition`,
       purchase_date, purchase_price, notes, created_at, updated_at
FROM hr_assets
WHERE business_id = sqlc.arg(business_id)
  AND (sqlc.narg(location_id) IS NULL OR location_id = sqlc.narg(location_id))
  AND (sqlc.narg(status) IS NULL OR status = sqlc.narg(status))
  AND (sqlc.narg(condition) IS NULL OR `condition` = sqlc.narg(condition))
  AND (sqlc.narg(assigned_to) IS NULL OR assigned_to = sqlc.narg(assigned_to))
ORDER BY name ASC
LIMIT ? OFFSET ?;

-- name: CountAssetsFiltered :one
SELECT COUNT(*) AS total
FROM hr_assets
WHERE business_id = sqlc.arg(business_id)
  AND (sqlc.narg(location_id) IS NULL OR location_id = sqlc.narg(location_id))
  AND (sqlc.narg(status) IS NULL OR status = sqlc.narg(status))
  AND (sqlc.narg(condition) IS NULL OR `condition` = sqlc.narg(condition))
  AND (sqlc.narg(assigned_to) IS NULL OR assigned_to = sqlc.narg(assigned_to));

-- name: ListAssetsByStatus :many
SELECT id, owner_id, business_id, name, category, serial_number, location_id, assigned_to, status, `condition`,
       purchase_date, purchase_price, notes, created_at, updated_at
FROM hr_assets
WHERE business_id = ? AND status = ?
ORDER BY name ASC
LIMIT ? OFFSET ?;

-- name: ListAssetsByLocation :many
SELECT id, owner_id, business_id, name, category, serial_number, location_id, assigned_to, status, `condition`,
       purchase_date, purchase_price, notes, created_at, updated_at
FROM hr_assets
WHERE business_id = ? AND location_id = ?
ORDER BY name ASC
LIMIT ? OFFSET ?;

-- name: ListAssetsByLocationAndStatus :many
SELECT id, owner_id, business_id, name, category, serial_number, location_id, assigned_to, status, `condition`,
       purchase_date, purchase_price, notes, created_at, updated_at
FROM hr_assets
WHERE business_id = ? AND location_id = ? AND status = ?
ORDER BY name ASC
LIMIT ? OFFSET ?;

-- name: AssignAsset :exec
UPDATE hr_assets
SET assigned_to = ?, status = 'assigned'
WHERE business_id = ? AND id = ?;

-- name: UnassignAsset :exec
UPDATE hr_assets
SET assigned_to = NULL, status = 'available'
WHERE business_id = ? AND id = ?;
