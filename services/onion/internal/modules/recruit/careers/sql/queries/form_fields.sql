-- name: ListFormFields :many
SELECT
    id,
    career_id,
    owner_id,
    business_id,
    kind,
    label,
    COALESCE(placeholder, '') AS placeholder,
    required,
    COALESCE(CAST(options_json AS CHAR), '') AS options_json,
    position,
    created_at,
    updated_at
FROM careers_form_fields
WHERE business_id = ? AND career_id = ?
ORDER BY position ASC, id ASC
LIMIT 100;

-- name: ListFormFieldsByCareer :many
SELECT
    id,
    career_id,
    owner_id,
    business_id,
    kind,
    label,
    COALESCE(placeholder, '') AS placeholder,
    required,
    COALESCE(CAST(options_json AS CHAR), '') AS options_json,
    position,
    created_at,
    updated_at
FROM careers_form_fields
WHERE career_id = ?
ORDER BY position ASC, id ASC
LIMIT 100;

-- name: GetFormField :one
SELECT
    id,
    career_id,
    owner_id,
    business_id,
    kind,
    label,
    COALESCE(placeholder, '') AS placeholder,
    required,
    COALESCE(CAST(options_json AS CHAR), '') AS options_json,
    position,
    created_at,
    updated_at
FROM careers_form_fields
WHERE id = ? AND business_id = ?
ORDER BY id ASC
LIMIT 1;

-- name: CreateFormField :execresult
INSERT INTO careers_form_fields (
    career_id,
    owner_id,
    business_id,
    kind,
    label,
    placeholder,
    required,
    options_json,
    position
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: UpdateFormField :exec
UPDATE careers_form_fields
SET
    kind = ?,
    label = ?,
    placeholder = ?,
    required = ?,
    options_json = ?,
    position = ?
WHERE id = ? AND business_id = ?;

-- name: DeleteFormField :execresult
DELETE FROM careers_form_fields
WHERE id = ? AND business_id = ?;
