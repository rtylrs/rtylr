-- name: CreateAsset :execresult
INSERT INTO hr_assets (owner_id, business_id, name, category, serial_number, location_id, status, purchase_date, purchase_price, notes)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetAsset :one
SELECT id, owner_id, business_id, name, category, serial_number, location_id, assigned_to, status,
       purchase_date, purchase_price, notes, created_at, updated_at
FROM hr_assets
WHERE business_id = ? AND id = ?;

-- name: ListAssets :many
SELECT id, owner_id, business_id, name, category, serial_number, location_id, assigned_to, status,
       purchase_date, purchase_price, notes, created_at, updated_at
FROM hr_assets
WHERE business_id = ?
ORDER BY name ASC;

-- name: ListAssetsByStatus :many
SELECT id, owner_id, business_id, name, category, serial_number, location_id, assigned_to, status,
       purchase_date, purchase_price, notes, created_at, updated_at
FROM hr_assets
WHERE business_id = ? AND status = ?
ORDER BY name ASC;

-- name: ListAssetsByLocation :many
SELECT id, owner_id, business_id, name, category, serial_number, location_id, assigned_to, status,
       purchase_date, purchase_price, notes, created_at, updated_at
FROM hr_assets
WHERE business_id = ? AND location_id = ?
ORDER BY name ASC;

-- name: ListAssetsByLocationAndStatus :many
SELECT id, owner_id, business_id, name, category, serial_number, location_id, assigned_to, status,
       purchase_date, purchase_price, notes, created_at, updated_at
FROM hr_assets
WHERE business_id = ? AND location_id = ? AND status = ?
ORDER BY name ASC;

-- name: AssignAsset :exec
UPDATE hr_assets
SET assigned_to = ?, status = 'assigned'
WHERE business_id = ? AND id = ?;

-- name: UnassignAsset :exec
UPDATE hr_assets
SET assigned_to = NULL, status = 'available'
WHERE business_id = ? AND id = ?;
