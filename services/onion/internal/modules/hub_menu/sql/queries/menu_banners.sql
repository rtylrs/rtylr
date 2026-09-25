-- name: ListMenuBanners :many
SELECT
    id,
    owner_id,
    business_id,
    title,
    COALESCE(body, '') AS body,
    COALESCE(image_url, '') AS image_url,
    COALESCE(link_url, '') AS link_url,
    position,
    is_active,
    starts_at,
    ends_at,
    created_at,
    updated_at
FROM hub_menu_banners
WHERE business_id = ?
ORDER BY position ASC, id ASC
LIMIT 200;

-- name: GetMenuBanner :one
SELECT
    id,
    owner_id,
    business_id,
    title,
    COALESCE(body, '') AS body,
    COALESCE(image_url, '') AS image_url,
    COALESCE(link_url, '') AS link_url,
    position,
    is_active,
    starts_at,
    ends_at,
    created_at,
    updated_at
FROM hub_menu_banners
WHERE business_id = ? AND id = ?
ORDER BY id ASC
LIMIT 1;

-- name: CreateMenuBanner :execresult
INSERT INTO hub_menu_banners (
    owner_id,
    business_id,
    title,
    body,
    image_url,
    link_url,
    position,
    is_active,
    starts_at,
    ends_at
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
);

-- name: UpdateMenuBanner :exec
UPDATE hub_menu_banners
SET title = ?,
    body = ?,
    image_url = ?,
    link_url = ?,
    position = ?,
    is_active = ?,
    starts_at = ?,
    ends_at = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE business_id = ? AND id = ?;

-- name: DeleteMenuBanner :exec
DELETE FROM hub_menu_banners
WHERE business_id = ? AND id = ?;

-- name: GetHubMenuBusinessOwnerID :one
SELECT owner_id FROM businesses WHERE id = ? AND is_active = TRUE;
