-- The generated list contract has no pagination fields; keep the endpoint
-- bounded until the corresponding voxie contract can expose a page cursor.
-- name: ListLocations :many
SELECT id, owner_id, business_id, name, slug, address, city, country, timezone, latitude, longitude, is_default, is_active, created_at, updated_at
FROM hr_locations
WHERE business_id = ?
ORDER BY name ASC
LIMIT 500;

-- name: GetLocation :one
SELECT id, owner_id, business_id, name, slug, address, city, country, timezone, latitude, longitude, is_default, is_active, created_at, updated_at
FROM hr_locations
WHERE business_id = ? AND id = ?;

-- name: CreateLocation :execresult
INSERT INTO hr_locations (owner_id, business_id, name, address, city, country, timezone, latitude, longitude, is_default)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: UpdateLocation :execresult
UPDATE hr_locations
SET name = ?, address = ?, city = ?, country = ?, timezone = ?, latitude = ?, longitude = ?, is_default = ?, updated_at = NOW()
WHERE id = ? AND business_id = ?;
