-- name: GetAllCareers :many
SELECT
    id,
    title,
    department,
    location,
    employment_type,
    work_mode,
    seniority,
    details,
    is_active
 FROM careers
WHERE is_active = true
ORDER BY created_at DESC;

-- name: CreateCareer :one
INSERT INTO careers (
    title,
    department,
    location,
    employment_type,
    work_mode,
    seniority,
    details,
    is_active
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING *;
