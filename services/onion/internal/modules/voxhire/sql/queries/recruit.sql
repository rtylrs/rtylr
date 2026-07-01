-- name: CreateAdmin :execresult
INSERT INTO voxhire_admins (email, password_hash, name) VALUES (?, ?, ?);

-- name: GetAdminByEmail :one
SELECT id, email, COALESCE(name,'') AS name, password_hash, created_at
FROM voxhire_admins WHERE email = ?;

-- name: GetAdminByID :one
SELECT id, email, COALESCE(name,'') AS name, password_hash, created_at
FROM voxhire_admins WHERE id = ?;

-- name: ListAdmins :many
SELECT id, email, COALESCE(name,'') AS name, created_at
FROM voxhire_admins ORDER BY created_at DESC;

-- name: CountAdmins :one
SELECT COUNT(*) AS count FROM voxhire_admins;

-- name: ListOpenings :many
SELECT c.id,
       COALESCE(c.slug, '') AS slug,
       c.title,
       COALESCE(c.department, '') AS department,
       COALESCE(c.location, '') AS location,
       COALESCE(c.employment_type, '') AS employment_type,
       COALESCE(c.work_mode, '') AS work_mode,
       COALESCE(c.seniority, '') AS seniority,
       COALESCE(c.details, '') AS details,
       COALESCE(c.salary_min, 0) AS salary_min,
       COALESCE(c.salary_max, 0) AS salary_max,
       COALESCE(c.salary_currency, 'USD') AS salary_currency,
       c.is_active,
       COALESCE(SUM(CASE WHEN a.status = 'new' THEN 1 ELSE 0 END), 0) AS new_count,
       COALESCE(COUNT(a.id), 0) AS total_count
  FROM careers c
  LEFT JOIN voxire_career_applications a ON a.job_id = c.id
 WHERE c.owner_id = ?
 GROUP BY c.id, c.slug, c.title, c.department, c.location, c.employment_type,
          c.work_mode, c.seniority, c.details, c.salary_min, c.salary_max,
          c.salary_currency, c.is_active
 ORDER BY c.is_active DESC, c.created_at DESC;

-- name: ListCompanyLeads :many
SELECT id, company_name, contact_name, COALESCE(contact_title,'') AS contact_title,
       work_email, COALESCE(phone,'') AS phone, role, COALESCE(headcount,'') AS headcount,
       COALESCE(timeline,'') AS timeline, COALESCE(service,'') AS service,
       COALESCE(message,'') AS message, agency_opt_in, status, created_at
FROM voxhire_company_leads
ORDER BY created_at DESC;

-- name: UpdateApplicationStatus :exec
UPDATE voxire_career_applications SET status = ? WHERE id = ?;

-- name: GetApplicationCVKey :one
SELECT COALESCE(cv_key, '') AS cv_key FROM voxire_career_applications WHERE id = ?;

-- name: InsertApplicationEvent :exec
INSERT INTO voxhire_application_events (application_id, kind, title, detail)
VALUES (?, ?, ?, ?);

-- name: ListApplicationEvents :many
SELECT id, application_id, kind, title, COALESCE(detail,'') AS detail, created_at
FROM voxhire_application_events
WHERE application_id = ?
ORDER BY created_at ASC;

-- name: ListTasksByStatus :many
SELECT id, application_id, title, status, due_at, completed_at, created_at
FROM voxhire_follow_up_tasks
WHERE (? = '' OR status = ?)
ORDER BY due_at IS NULL, due_at ASC, created_at DESC;

-- name: UpsertFollowUpTask :exec
INSERT INTO voxhire_follow_up_tasks (application_id, title, status, due_at)
VALUES (?, ?, ?, ?)
ON DUPLICATE KEY UPDATE status = VALUES(status), due_at = VALUES(due_at);

-- name: UpdateFollowUpTask :exec
UPDATE voxhire_follow_up_tasks
SET title = ?, status = ?, due_at = ?, completed_at = ?
WHERE id = ?;

-- name: GetFollowUpTaskByID :one
SELECT id, application_id, title, status, due_at, completed_at, created_at
FROM voxhire_follow_up_tasks WHERE id = ?;

-- name: ListShortlistsByJob :many
SELECT id, job_id, name, status, COALESCE(notes,'') AS notes, created_at, updated_at
FROM voxhire_shortlists WHERE job_id = ? ORDER BY created_at DESC;

-- name: CreateShortlist :execresult
INSERT INTO voxhire_shortlists (job_id, name, status, notes) VALUES (?, ?, ?, ?);

-- name: AddShortlistItem :execresult
INSERT INTO voxhire_shortlist_items (shortlist_id, application_id, sort_order, notes)
VALUES (?, ?, ?, ?);

-- name: ListShortlistItems :many
SELECT id, shortlist_id, application_id, sort_order, COALESCE(notes,'') AS notes, created_at
FROM voxhire_shortlist_items WHERE shortlist_id = ? ORDER BY sort_order ASC, created_at ASC;

-- name: GetShortlistItemByID :one
SELECT id, shortlist_id, application_id, sort_order, COALESCE(notes,'') AS notes, created_at
FROM voxhire_shortlist_items WHERE id = ?;

-- name: CountCareerApplications :one
SELECT COUNT(*) FROM voxire_career_applications WHERE id = ?;
