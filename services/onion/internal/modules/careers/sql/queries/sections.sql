-- name: ListSections :many
SELECT
    id,
    owner_id,
    business_id,
    kind,
    COALESCE(title, '') AS title,
    COALESCE(body, '') AS body,
    COALESCE(media_url, '') AS media_url,
    COALESCE(CAST(items_json AS CHAR), '') AS items_json,
    position,
    enabled,
    created_at,
    updated_at
FROM careers_sections
WHERE business_id = ?
ORDER BY position ASC, id ASC
LIMIT 100;

-- name: ListEnabledSections :many
SELECT
    id,
    owner_id,
    business_id,
    kind,
    COALESCE(title, '') AS title,
    COALESCE(body, '') AS body,
    COALESCE(media_url, '') AS media_url,
    COALESCE(CAST(items_json AS CHAR), '') AS items_json,
    position,
    enabled,
    created_at,
    updated_at
FROM careers_sections
WHERE business_id = ? AND enabled = true
ORDER BY position ASC, id ASC
LIMIT 100;

-- name: GetSection :one
SELECT
    id,
    owner_id,
    business_id,
    kind,
    COALESCE(title, '') AS title,
    COALESCE(body, '') AS body,
    COALESCE(media_url, '') AS media_url,
    COALESCE(CAST(items_json AS CHAR), '') AS items_json,
    position,
    enabled,
    created_at,
    updated_at
FROM careers_sections
WHERE id = ? AND business_id = ?
ORDER BY id ASC
LIMIT 1;

-- name: CreateSection :execresult
INSERT INTO careers_sections (
    owner_id,
    business_id,
    kind,
    title,
    body,
    media_url,
    items_json,
    position,
    enabled
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: UpdateSection :exec
UPDATE careers_sections
SET
    kind = ?,
    title = ?,
    body = ?,
    media_url = ?,
    items_json = ?,
    position = ?,
    enabled = ?
WHERE id = ? AND business_id = ?;

-- name: UpdateSectionPosition :exec
UPDATE careers_sections
SET position = ?
WHERE id = ? AND business_id = ?;

-- name: DeleteSection :execresult
DELETE FROM careers_sections
WHERE id = ? AND business_id = ?;
