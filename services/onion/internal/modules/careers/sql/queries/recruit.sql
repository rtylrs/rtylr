-- name: ListRecruitCareers :many
SELECT
    id,
    owner_id,
    business_id,
    COALESCE(slug, '') AS slug,
    title,
    COALESCE(department, '') AS department,
    COALESCE(location, '') AS location,
    COALESCE(employment_type, '') AS employment_type,
    COALESCE(work_mode, '') AS work_mode,
    COALESCE(seniority, '') AS seniority,
    details,
    COALESCE(requirements, '') AS requirements,
    COALESCE(salary_min, 0) AS salary_min,
    COALESCE(salary_max, 0) AS salary_max,
    COALESCE(salary_currency, '') AS salary_currency,
    COALESCE(is_published, false) AS is_published,
    is_active,
    created_at,
    updated_at
FROM careers
WHERE business_id = ? AND is_active = true
ORDER BY created_at DESC;

-- name: ListRecruitCareersPublished :many
SELECT
    id,
    owner_id,
    business_id,
    COALESCE(slug, '') AS slug,
    title,
    COALESCE(department, '') AS department,
    COALESCE(location, '') AS location,
    COALESCE(employment_type, '') AS employment_type,
    COALESCE(work_mode, '') AS work_mode,
    COALESCE(seniority, '') AS seniority,
    details,
    COALESCE(requirements, '') AS requirements,
    COALESCE(salary_min, 0) AS salary_min,
    COALESCE(salary_max, 0) AS salary_max,
    COALESCE(salary_currency, '') AS salary_currency,
    COALESCE(is_published, false) AS is_published,
    is_active,
    created_at,
    updated_at
FROM careers
WHERE business_id = ? AND is_active = true AND is_published = true
ORDER BY created_at DESC
LIMIT 100;

-- name: GetRecruitCareerBySlug :one
SELECT
    id,
    owner_id,
    business_id,
    COALESCE(slug, '') AS slug,
    title,
    COALESCE(department, '') AS department,
    COALESCE(location, '') AS location,
    COALESCE(employment_type, '') AS employment_type,
    COALESCE(work_mode, '') AS work_mode,
    COALESCE(seniority, '') AS seniority,
    details,
    COALESCE(requirements, '') AS requirements,
    COALESCE(salary_min, 0) AS salary_min,
    COALESCE(salary_max, 0) AS salary_max,
    COALESCE(salary_currency, '') AS salary_currency,
    COALESCE(is_published, false) AS is_published,
    is_active,
    created_at,
    updated_at
FROM careers
WHERE business_id = ? AND slug = ?
ORDER BY slug ASC
LIMIT 1;

-- name: GetRecruitCareerByID :one
SELECT
    id,
    owner_id,
    business_id,
    COALESCE(slug, '') AS slug,
    title,
    COALESCE(department, '') AS department,
    COALESCE(location, '') AS location,
    COALESCE(employment_type, '') AS employment_type,
    COALESCE(work_mode, '') AS work_mode,
    COALESCE(seniority, '') AS seniority,
    details,
    COALESCE(requirements, '') AS requirements,
    COALESCE(salary_min, 0) AS salary_min,
    COALESCE(salary_max, 0) AS salary_max,
    COALESCE(salary_currency, '') AS salary_currency,
    COALESCE(is_published, false) AS is_published,
    is_active,
    created_at,
    updated_at
FROM careers
WHERE id = ? AND business_id = ?
LIMIT 1;

-- name: CreateRecruitCareer :execresult
INSERT INTO careers (
    owner_id,
    business_id,
    slug,
    title,
    department,
    location,
    employment_type,
    work_mode,
    seniority,
    details,
    requirements,
    salary_min,
    salary_max,
    salary_currency,
    is_published,
    is_active
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, true);

-- name: UpdateRecruitCareer :exec
UPDATE careers
SET
    slug = ?,
    title = ?,
    department = ?,
    location = ?,
    employment_type = ?,
    work_mode = ?,
    seniority = ?,
    details = ?,
    requirements = ?,
    salary_min = ?,
    salary_max = ?,
    salary_currency = ?,
    is_published = ?
WHERE id = ? AND business_id = ?;

-- name: SoftDeleteRecruitCareer :exec
UPDATE careers
SET is_active = false, is_published = false
WHERE id = ? AND business_id = ?;

-- name: GetCustomizationByBusiness :one
SELECT
    id,
    owner_id,
    business_id,
    slug,
    brand_name,
    COALESCE(tagline, '') AS tagline,
    COALESCE(about_blurb, '') AS about_blurb,
    COALESCE(logo_url, '') AS logo_url,
    COALESCE(primary_color, '') AS primary_color,
    COALESCE(accent_color, '') AS accent_color,
    COALESCE(bg_color, '') AS bg_color,
    COALESCE(text_color, '') AS text_color,
    COALESCE(muted_text_color, '') AS muted_text_color,
    COALESCE(font_pairing, '') AS font_pairing,
    COALESCE(hero_image_url, '') AS hero_image_url,
    COALESCE(favicon_url, '') AS favicon_url,
    template,
    COALESCE(header_html, '') AS header_html,
    COALESCE(meta_title, '') AS meta_title,
    COALESCE(meta_description, '') AS meta_description,
    COALESCE(privacy_url, '') AS privacy_url,
    COALESCE(terms_url, '') AS terms_url,
    COALESCE(ga_measurement_id, '') AS ga_measurement_id,
    COALESCE(meta_pixel_id, '') AS meta_pixel_id,
    require_consent,
    COALESCE(notification_email, '') AS notification_email,
    email_notifications,
    created_at,
    updated_at
FROM careers_customizations
WHERE business_id = ?
ORDER BY business_id ASC
LIMIT 1;

-- name: GetCustomizationBySlug :one
SELECT
    id,
    owner_id,
    business_id,
    slug,
    brand_name,
    COALESCE(tagline, '') AS tagline,
    COALESCE(about_blurb, '') AS about_blurb,
    COALESCE(logo_url, '') AS logo_url,
    COALESCE(primary_color, '') AS primary_color,
    COALESCE(accent_color, '') AS accent_color,
    COALESCE(bg_color, '') AS bg_color,
    COALESCE(text_color, '') AS text_color,
    COALESCE(muted_text_color, '') AS muted_text_color,
    COALESCE(font_pairing, '') AS font_pairing,
    COALESCE(hero_image_url, '') AS hero_image_url,
    COALESCE(favicon_url, '') AS favicon_url,
    template,
    COALESCE(header_html, '') AS header_html,
    COALESCE(meta_title, '') AS meta_title,
    COALESCE(meta_description, '') AS meta_description,
    COALESCE(privacy_url, '') AS privacy_url,
    COALESCE(terms_url, '') AS terms_url,
    COALESCE(ga_measurement_id, '') AS ga_measurement_id,
    COALESCE(meta_pixel_id, '') AS meta_pixel_id,
    require_consent,
    COALESCE(notification_email, '') AS notification_email,
    email_notifications,
    created_at,
    updated_at
FROM careers_customizations
WHERE slug = ?
ORDER BY slug ASC
LIMIT 1;

-- name: UpsertCustomization :exec
INSERT INTO careers_customizations (
    owner_id,
    business_id,
    slug,
    brand_name,
    tagline,
    about_blurb,
    logo_url,
    primary_color,
    accent_color,
    bg_color,
    text_color,
    muted_text_color,
    font_pairing,
    hero_image_url,
    favicon_url,
    template,
    header_html,
    meta_title,
    meta_description,
    privacy_url,
    terms_url,
    ga_measurement_id,
    meta_pixel_id,
    require_consent,
    notification_email,
    email_notifications
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
    slug = VALUES(slug),
    brand_name = VALUES(brand_name),
    tagline = VALUES(tagline),
    about_blurb = VALUES(about_blurb),
    logo_url = VALUES(logo_url),
    primary_color = VALUES(primary_color),
    accent_color = VALUES(accent_color),
    bg_color = VALUES(bg_color),
    text_color = VALUES(text_color),
    muted_text_color = VALUES(muted_text_color),
    font_pairing = VALUES(font_pairing),
    hero_image_url = VALUES(hero_image_url),
    favicon_url = VALUES(favicon_url),
    template = VALUES(template),
    header_html = VALUES(header_html),
    meta_title = VALUES(meta_title),
    meta_description = VALUES(meta_description),
    privacy_url = VALUES(privacy_url),
    terms_url = VALUES(terms_url),
    ga_measurement_id = VALUES(ga_measurement_id),
    meta_pixel_id = VALUES(meta_pixel_id),
    require_consent = VALUES(require_consent),
    notification_email = VALUES(notification_email),
    email_notifications = VALUES(email_notifications);

-- name: CreateCareerSubmission :execresult
INSERT INTO careers_submissions (
    career_id,
    owner_id,
    business_id,
    candidate_name,
    email,
    phone,
    resume_url,
    cover_letter,
    answers_json
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: ListCareerSubmissions :many
SELECT
    s.id,
    s.career_id,
    s.owner_id,
    s.business_id,
    s.candidate_name,
    s.email,
    COALESCE(s.phone, '') AS phone,
    COALESCE(s.resume_url, '') AS resume_url,
    COALESCE(s.cover_letter, '') AS cover_letter,
    s.status,
    COALESCE(CAST(s.answers_json AS CHAR), '') AS answers_json,
    s.created_at,
    s.updated_at,
    COALESCE(c.title, '') AS career_title
FROM careers_submissions s
LEFT JOIN careers c ON c.id = s.career_id
WHERE s.business_id = ?
ORDER BY s.created_at DESC
LIMIT 100;

-- name: ListCareerSubmissionsByStatus :many
SELECT
    s.id,
    s.career_id,
    s.owner_id,
    s.business_id,
    s.candidate_name,
    s.email,
    COALESCE(s.phone, '') AS phone,
    COALESCE(s.resume_url, '') AS resume_url,
    COALESCE(s.cover_letter, '') AS cover_letter,
    s.status,
    COALESCE(CAST(s.answers_json AS CHAR), '') AS answers_json,
    s.created_at,
    s.updated_at,
    COALESCE(c.title, '') AS career_title
FROM careers_submissions s
LEFT JOIN careers c ON c.id = s.career_id
WHERE s.business_id = ? AND s.status = ?
ORDER BY s.created_at DESC
LIMIT 100;

-- The paged variants below back reads that must not silently truncate at the
-- UI page size (the submissions CSV export walks them until a short page).
-- id breaks created_at ties so OFFSET paging cannot skip or repeat rows.
-- name: ListCareerSubmissionsPage :many
SELECT
    s.id,
    s.career_id,
    s.owner_id,
    s.business_id,
    s.candidate_name,
    s.email,
    COALESCE(s.phone, '') AS phone,
    COALESCE(s.resume_url, '') AS resume_url,
    COALESCE(s.cover_letter, '') AS cover_letter,
    s.status,
    COALESCE(CAST(s.answers_json AS CHAR), '') AS answers_json,
    s.created_at,
    s.updated_at,
    COALESCE(c.title, '') AS career_title
FROM careers_submissions s
LEFT JOIN careers c ON c.id = s.career_id
WHERE s.business_id = ?
ORDER BY s.created_at DESC, s.id DESC
LIMIT ? OFFSET ?;

-- name: ListCareerSubmissionsByStatusPage :many
SELECT
    s.id,
    s.career_id,
    s.owner_id,
    s.business_id,
    s.candidate_name,
    s.email,
    COALESCE(s.phone, '') AS phone,
    COALESCE(s.resume_url, '') AS resume_url,
    COALESCE(s.cover_letter, '') AS cover_letter,
    s.status,
    COALESCE(CAST(s.answers_json AS CHAR), '') AS answers_json,
    s.created_at,
    s.updated_at,
    COALESCE(c.title, '') AS career_title
FROM careers_submissions s
LEFT JOIN careers c ON c.id = s.career_id
WHERE s.business_id = ? AND s.status = ?
ORDER BY s.created_at DESC, s.id DESC
LIMIT ? OFFSET ?;

-- name: GetCareerSubmissionResumeURL :one
SELECT COALESCE(resume_url, '') AS resume_url
FROM careers_submissions
WHERE id = ? AND business_id = ?
LIMIT 1;

-- name: UpdateCareerSubmissionStatus :exec
UPDATE careers_submissions
SET status = ?
WHERE id = ? AND business_id = ?;
