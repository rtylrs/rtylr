-- name: GetBusinessProfile :one
SELECT COALESCE(site_name,'') AS site_name, COALESCE(logo_url,'') AS logo_url, COALESCE(contact_email,'') AS contact_email, COALESCE(contact_number,'') AS contact_number
FROM config
WHERE owner_id = ?
ORDER BY owner_id ASC
LIMIT 1;

-- name: GetBusinessProfileByBusiness :one
SELECT
    b.id,
    b.name,
    COALESCE(NULLIF(c.site_name, ''), b.name) AS site_name,
    COALESCE(NULLIF(c.logo_url, ''), COALESCE(b.logo_url, '')) AS logo_url,
    COALESCE(c.contact_email, '') AS contact_email,
    COALESCE(c.contact_number, '') AS contact_number,
    COALESCE(b.currency, 'USD') AS currency,
    COALESCE(b.timezone, 'UTC') AS timezone
FROM businesses b
LEFT JOIN config c ON c.business_id = b.id
WHERE (b.id = sqlc.arg(owner_id) OR b.owner_id = sqlc.arg(owner_id)) AND b.is_active = TRUE
ORDER BY (b.id = sqlc.arg(owner_id)) DESC, b.id ASC
LIMIT 1;

-- name: GetPublicBusiness :one
SELECT id, name, COALESCE(slug, '') AS slug, COALESCE(logo_url, '') AS logo_url,
       COALESCE(currency, 'USD') AS currency, COALESCE(timezone, 'UTC') AS timezone
FROM businesses
WHERE id = ? AND is_active = TRUE
ORDER BY id ASC
LIMIT 1;

-- name: LookupBusinessBySlug :one
SELECT id, name, COALESCE(slug, '') AS slug, COALESCE(logo_url, '') AS logo_url,
       COALESCE(currency, 'USD') AS currency, COALESCE(timezone, 'UTC') AS timezone,
       COALESCE(owner_id, 0) AS owner_id
FROM businesses
WHERE slug = ? AND is_active = TRUE
LIMIT 1;

-- cross-schema read from hr_core — expose active HR locations in the public business profile.
-- hr_locations is business-scoped (hr_core writes owner_id = business_id), so
-- a storefront must list its own business_id, not every location of its owner.
-- name: ListPublicLocationsByBusiness :many
SELECT id, name, COALESCE(slug, '') AS slug, COALESCE(address, '') AS address,
       COALESCE(timezone, '') AS timezone, is_active
FROM hr_locations
WHERE business_id = ? AND is_active = TRUE
ORDER BY is_default DESC, name ASC
LIMIT 100;

-- name: UpdateBusinessSlug :exec
UPDATE businesses SET slug = ? WHERE id = ?;

-- name: RenameBusiness :execresult
UPDATE businesses
SET name = ?
WHERE id = ? AND owner_id = ? AND is_active = TRUE;

-- name: SetBusinessLogo :execresult
UPDATE businesses
SET logo_url = ?
WHERE id = ? AND owner_id = ? AND is_active = TRUE;

-- name: DeactivateBusiness :execresult
UPDATE businesses
SET is_active = FALSE
WHERE id = ? AND owner_id = ? AND is_active = TRUE;

-- name: CountActiveOwnedBusinesses :one
SELECT COUNT(*) AS cnt
FROM businesses
WHERE owner_id = ? AND is_active = TRUE;

-- name: ListUserBusinesses :many
SELECT b.id, b.name, COALESCE(b.logo_url, '') AS logo_url, COALESCE(b.parent_id, 0) AS parent_id,
       b.is_active, (b.owner_id = ?) AS is_owned
FROM businesses b
WHERE b.is_active = TRUE
  AND (b.owner_id = ? OR EXISTS (
    SELECT 1 FROM user_business_access uba WHERE uba.user_id = ? AND uba.business_id = b.id
  ) OR EXISTS (
    SELECT 1
    FROM user_app_permissions uap
    WHERE uap.user_id = ?
      AND uap.business_id = b.id
      AND uap.is_active = TRUE
  ))
ORDER BY b.id ASC
LIMIT 100;

-- name: InsertSubBusiness :execresult
INSERT INTO businesses (name, owner_id, parent_id) VALUES (?, ?, ?);

-- name: GrantUserBusinessAccess :exec
INSERT INTO user_business_access (user_id, business_id, role) VALUES (?, ?, 'owner');

-- name: CheckSubBusinessOwnership :one
SELECT COUNT(*) AS cnt FROM businesses WHERE id = ? AND owner_id = ? AND is_active = TRUE;
