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
ORDER BY created_at DESC
LIMIT 100;

-- name: CreateCareer :execresult
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
    ?, ?, ?, ?, ?, ?, ?, ?
);

-- name: SaveCareerApplication :execresult
INSERT INTO voxire_career_applications
    (job_id, name, email, phone, work_history, education, experience_summary, cover_letter, cv_key)
VALUES
    (?, ?, ?, ?, ?, ?, ?, ?, ?);
