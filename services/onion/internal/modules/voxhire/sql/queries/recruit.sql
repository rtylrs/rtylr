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
FROM voxhire_admins ORDER BY created_at DESC LIMIT 100;

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

-- name: ListOpeningsPage :many
WITH opening_page AS (
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
           c.created_at
    FROM careers c
    WHERE c.owner_id = ?
    ORDER BY c.is_active DESC, c.created_at DESC, c.id DESC
    LIMIT ? OFFSET ?
), application_counts AS (
    SELECT a.job_id,
           SUM(a.status = 'new') AS new_count,
           COUNT(*) AS total_count
    FROM voxire_career_applications a
    JOIN opening_page p ON p.id = a.job_id
    GROUP BY a.job_id
)
SELECT p.id, p.slug, p.title, p.department, p.location, p.employment_type,
       p.work_mode, p.seniority, p.details, p.salary_min, p.salary_max,
       p.salary_currency, p.is_active,
       COALESCE(ac.new_count, 0) AS new_count,
       COALESCE(ac.total_count, 0) AS total_count
FROM opening_page p
LEFT JOIN application_counts ac ON ac.job_id = p.id
ORDER BY p.is_active DESC, p.created_at DESC, p.id DESC;

-- name: CountOpenings :one
SELECT COUNT(*) FROM careers WHERE owner_id = ?;

-- name: GetCareerOpeningByID :one
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
 WHERE c.owner_id = ? AND c.id = ?
 GROUP BY c.id, c.slug, c.title, c.department, c.location, c.employment_type,
          c.work_mode, c.seniority, c.details, c.salary_min, c.salary_max,
          c.salary_currency, c.is_active;

-- name: CreateCareer :execresult
INSERT INTO careers (
    owner_id, slug, title, department, location, employment_type,
    work_mode, seniority, details, salary_min, salary_max, salary_currency,
    is_active, is_published
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, TRUE, TRUE);

-- name: UpdateCareer :exec
UPDATE careers
   SET slug = ?,
       title = ?,
       department = ?,
       location = ?,
       employment_type = ?,
       work_mode = ?,
       seniority = ?,
       details = ?,
       salary_min = ?,
       salary_max = ?,
       salary_currency = ?
 WHERE id = ? AND owner_id = ?;

-- name: SetCareerActive :exec
UPDATE careers
   SET is_active = ?,
       is_published = ?
 WHERE id = ? AND owner_id = ?;

-- name: ListCompanyLeads :many
SELECT id, company_name, contact_name, COALESCE(contact_title,'') AS contact_title,
       work_email, COALESCE(phone,'') AS phone, role, COALESCE(headcount,'') AS headcount,
       COALESCE(timeline,'') AS timeline, COALESCE(service,'') AS service,
       COALESCE(message,'') AS message, agency_opt_in, status, created_at
FROM voxhire_company_leads
ORDER BY created_at DESC
LIMIT 100;

-- name: UpdateApplicationStatus :exec
UPDATE voxire_career_applications a
JOIN careers c ON c.id = a.job_id
   SET a.status = ?
 WHERE a.id = ? AND c.owner_id = ?;

-- name: UpdateApplicationReview :exec
UPDATE voxire_career_applications
JOIN careers c ON c.id = voxire_career_applications.job_id
   SET voxire_career_applications.status = CASE WHEN ? <> '' THEN ? ELSE voxire_career_applications.status END,
       voxire_career_applications.review_rating = ?,
       voxire_career_applications.review_notes = ?,
       voxire_career_applications.next_step = ?,
       reviewed_at = CURRENT_TIMESTAMP
 WHERE voxire_career_applications.id = ? AND c.owner_id = ?;

-- name: GetApplicationCVKey :one
SELECT COALESCE(a.cv_key, '') AS cv_key
FROM voxire_career_applications a
JOIN careers c ON c.id = a.job_id AND c.owner_id = ?
WHERE a.id = ?;

-- name: CreateCareerApplication :execresult
INSERT INTO voxire_career_applications (
    job_id, name, email, phone, location, salary_expectation,
    work_history, education, experience_summary, cover_letter, cv_key, status
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'new');

-- name: CountCareerApplications :one
SELECT COUNT(*) FROM voxire_career_applications WHERE id = ?;

-- name: ApplicationExists :one
SELECT COUNT(*) AS count
FROM voxire_career_applications a
JOIN careers c ON c.id = a.job_id AND c.owner_id = ?
WHERE a.id = ?;

-- name: InsertApplicationEvent :exec
INSERT INTO voxhire_application_events (application_id, kind, title, detail)
VALUES (?, ?, ?, ?);

-- name: ListApplicationEvents :many
SELECT e.kind, e.title, COALESCE(e.detail,'') AS detail, e.created_at
FROM voxhire_application_events e
JOIN voxire_career_applications a ON a.id = e.application_id
JOIN careers c ON c.id = a.job_id AND c.owner_id = ?
WHERE e.application_id = ?
ORDER BY e.created_at DESC;

-- name: GetApplicationTimelineMeta :one
SELECT a.created_at,
       COALESCE(a.cv_processing_status, 'pending') AS cv_processing_status,
       COALESCE(a.cv_processing_error, '') AS cv_processing_error,
       a.cv_processed_at,
       a.reviewed_at
FROM voxire_career_applications a
JOIN careers c ON c.id = a.job_id AND c.owner_id = ?
WHERE a.id = ?;

-- name: ListTaskTimelineEvents :many
SELECT t.title, t.status, t.created_at, t.completed_at
FROM voxhire_follow_up_tasks t
JOIN voxire_career_applications a ON a.id = t.application_id
JOIN careers c ON c.id = a.job_id AND c.owner_id = ?
WHERE t.application_id = ?;

-- name: ListShortlistTimelineEvents :many
SELECT s.name, i.sort_order, i.created_at
FROM voxhire_shortlist_items i
JOIN voxhire_shortlists s ON s.id = i.shortlist_id
JOIN voxire_career_applications a ON a.id = i.application_id
JOIN careers c ON c.id = a.job_id AND c.owner_id = ?
WHERE i.application_id = ?;

-- name: ListTasks :many
SELECT t.id, t.application_id, a.job_id, COALESCE(c.title, '') AS opening_title,
       a.name AS candidate_name, a.email AS candidate_email,
       t.title AS title, t.status AS status, t.due_at, t.completed_at, t.created_at AS created_at
FROM voxhire_follow_up_tasks t
JOIN voxire_career_applications a ON a.id = t.application_id
JOIN careers c ON c.id = a.job_id AND c.owner_id = ?
WHERE (sqlc.narg('filter_status') IS NULL OR t.status = sqlc.narg('filter_status'))
ORDER BY
  CASE t.status WHEN 'open' THEN 0 WHEN 'done' THEN 2 ELSE 1 END,
  t.due_at IS NULL,
  t.due_at ASC,
  t.created_at DESC
LIMIT 100;

-- name: GetTaskByID :one
SELECT t.id, t.application_id, a.job_id, COALESCE(c.title, '') AS opening_title,
       a.name AS candidate_name, a.email AS candidate_email,
       t.title AS title, t.status AS status, t.due_at, t.completed_at, t.created_at AS created_at
FROM voxhire_follow_up_tasks t
JOIN voxire_career_applications a ON a.id = t.application_id
JOIN careers c ON c.id = a.job_id AND c.owner_id = ?
WHERE t.id = ?;

-- name: UpsertFollowUpTask :execresult
INSERT INTO voxhire_follow_up_tasks (application_id, title, status, due_at)
VALUES (?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
    status = VALUES(status),
    due_at = VALUES(due_at),
    completed_at = CASE
        WHEN VALUES(status) = 'done' THEN COALESCE(completed_at, CURRENT_TIMESTAMP)
        ELSE NULL
    END,
    id = LAST_INSERT_ID(id);

-- name: UpdateFollowUpTask :exec
UPDATE voxhire_follow_up_tasks t
JOIN voxire_career_applications a ON a.id = t.application_id
JOIN careers c ON c.id = a.job_id
   SET t.title = CASE WHEN ? <> '' THEN ? ELSE t.title END,
       t.status = CASE WHEN ? <> '' THEN ? ELSE t.status END,
       t.due_at = ?,
       t.completed_at = CASE
         WHEN ? = 'done' THEN COALESCE(t.completed_at, CURRENT_TIMESTAMP)
         WHEN ? = 'open' THEN NULL
         ELSE t.completed_at
       END
 WHERE t.id = ? AND c.owner_id = ?;

-- name: CountFollowUpTasks :one
SELECT COUNT(*) FROM voxhire_follow_up_tasks WHERE id = ?;

-- name: ListShortlists :many
SELECT s.id, s.job_id, COALESCE(c.title, '') AS job_title, s.name, s.status,
       COALESCE(s.notes, '') AS notes, s.created_at, s.updated_at
FROM voxhire_shortlists s
JOIN careers c ON c.id = s.job_id AND c.owner_id = ?
WHERE (sqlc.narg('job_id') IS NULL OR s.job_id = sqlc.narg('job_id'))
ORDER BY s.updated_at DESC
LIMIT 100;

-- name: CreateShortlist :execresult
INSERT INTO voxhire_shortlists (job_id, name, status, notes) VALUES (?, ?, ?, ?);

-- name: ShortlistExists :one
SELECT COUNT(*) AS count
FROM voxhire_shortlists s
JOIN careers c ON c.id = s.job_id AND c.owner_id = ?
WHERE s.id = ?;

-- name: UpsertShortlistItem :execresult
INSERT INTO voxhire_shortlist_items (shortlist_id, application_id, sort_order, notes)
VALUES (?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
    sort_order = VALUES(sort_order),
    notes = VALUES(notes),
    id = LAST_INSERT_ID(id);

-- name: ListShortlistItems :many
SELECT i.id, i.shortlist_id, i.application_id, a.name AS candidate_name, a.email AS candidate_email,
       COALESCE(a.status, 'new') AS status, i.sort_order, COALESCE(i.notes, '') AS notes,
       COALESCE(a.cv_profile_json, '') AS cv_profile_json, i.created_at
FROM voxhire_shortlist_items i
JOIN voxhire_shortlists s ON s.id = i.shortlist_id
JOIN careers c ON c.id = s.job_id AND c.owner_id = ?
JOIN voxire_career_applications a ON a.id = i.application_id
WHERE i.shortlist_id = ?
ORDER BY i.sort_order ASC, i.created_at ASC;

-- name: GetShortlistItemByID :one
SELECT i.id, i.shortlist_id, i.application_id, a.name AS candidate_name, a.email AS candidate_email,
       COALESCE(a.status, 'new') AS status, i.sort_order, COALESCE(i.notes, '') AS notes,
       COALESCE(a.cv_profile_json, '') AS cv_profile_json, i.created_at
FROM voxhire_shortlist_items i
JOIN voxhire_shortlists s ON s.id = i.shortlist_id
JOIN careers c ON c.id = s.job_id AND c.owner_id = ?
JOIN voxire_career_applications a ON a.id = i.application_id
WHERE i.id = ?;

-- name: AnalyticsCareerJobTotals :one
SELECT COALESCE(SUM(CASE WHEN is_active = 1 THEN 1 ELSE 0 END), 0) AS open_jobs,
       COALESCE(SUM(CASE WHEN is_active = 0 THEN 1 ELSE 0 END), 0) AS closed_jobs
FROM careers
WHERE owner_id = ?;

-- name: AnalyticsApplicationTotals :one
SELECT COUNT(*) AS applications_total,
       COALESCE(SUM(CASE WHEN a.created_at >= DATE_SUB(UTC_TIMESTAMP(), INTERVAL 7 DAY) THEN 1 ELSE 0 END), 0) AS applications_7d,
       COALESCE(SUM(CASE WHEN a.created_at >= DATE_SUB(UTC_TIMESTAMP(), INTERVAL 30 DAY) THEN 1 ELSE 0 END), 0) AS applications_30d,
       COALESCE(SUM(CASE WHEN LOWER(COALESCE(a.status, 'new')) = 'new' THEN 1 ELSE 0 END), 0) AS new_applications,
       COALESCE(SUM(CASE WHEN a.cv_key IS NOT NULL AND a.cv_key <> '' THEN 1 ELSE 0 END), 0) AS cv_count,
       AVG(NULLIF(a.salary_expectation, 0)) AS avg_salary_expectation
FROM voxire_career_applications a
JOIN careers c ON c.id = a.job_id AND c.owner_id = ?;

-- name: AnalyticsTalentPoolTotals :one
SELECT COUNT(*) AS talent_pool_total,
       COALESCE(SUM(CASE WHEN created_at >= DATE_SUB(UTC_TIMESTAMP(), INTERVAL 7 DAY) THEN 1 ELSE 0 END), 0) AS talent_pool_7d,
       COALESCE(SUM(CASE WHEN created_at >= DATE_SUB(UTC_TIMESTAMP(), INTERVAL 30 DAY) THEN 1 ELSE 0 END), 0) AS talent_pool_30d
FROM voxhire_talent_applications;

-- name: AnalyticsCVProcessingTotals :one
SELECT COALESCE(SUM(CASE WHEN status = 'processed' THEN 1 ELSE 0 END), 0) AS cv_processed_total,
       COALESCE(SUM(CASE WHEN status = 'error' THEN 1 ELSE 0 END), 0) AS cv_error_total,
       COALESCE(SUM(CASE WHEN status NOT IN ('processed', 'error') THEN 1 ELSE 0 END), 0) AS cv_pending_total
FROM (
    SELECT COALESCE(a.cv_processing_status, 'pending') AS status
      FROM voxire_career_applications a
      JOIN careers c ON c.id = a.job_id AND c.owner_id = ?
     WHERE a.cv_key IS NOT NULL AND a.cv_key <> ''
    UNION ALL
    SELECT COALESCE(cv_processing_status, 'pending') AS status
      FROM voxhire_talent_applications
     WHERE cv_key IS NOT NULL AND cv_key <> ''
) cv_statuses;

-- name: AnalyticsLeadTotals :one
SELECT COUNT(*) AS leads_total,
       COALESCE(SUM(CASE WHEN created_at >= DATE_SUB(UTC_TIMESTAMP(), INTERVAL 7 DAY) THEN 1 ELSE 0 END), 0) AS leads_7d,
       COALESCE(SUM(CASE WHEN created_at >= DATE_SUB(UTC_TIMESTAMP(), INTERVAL 30 DAY) THEN 1 ELSE 0 END), 0) AS leads_30d
FROM voxhire_company_leads;

-- name: AnalyticsOpenTaskTotals :one
SELECT COUNT(*) AS open_tasks,
       COALESCE(SUM(CASE WHEN status = 'open' AND due_at IS NOT NULL AND due_at < UTC_TIMESTAMP() THEN 1 ELSE 0 END), 0) AS overdue_tasks
FROM voxhire_follow_up_tasks
JOIN voxire_career_applications a ON a.id = voxhire_follow_up_tasks.application_id
JOIN careers c ON c.id = a.job_id AND c.owner_id = ?
WHERE status <> 'done';

-- name: AnalyticsShortlistTotals :one
SELECT COUNT(DISTINCT s.id) AS shortlists_total,
       COUNT(i.id) AS shortlisted_candidates
FROM voxhire_shortlists s
JOIN careers c ON c.id = s.job_id AND c.owner_id = ?
LEFT JOIN voxhire_shortlist_items i ON i.shortlist_id = s.id;

-- name: AnalyticsApplicationsByStatus :many
SELECT LOWER(COALESCE(NULLIF(a.status, ''), 'new')) AS bucket, COUNT(*) AS count
FROM voxire_career_applications a
JOIN careers c ON c.id = a.job_id AND c.owner_id = ?
WHERE a.created_at >= DATE_SUB(UTC_TIMESTAMP(), INTERVAL 12 MONTH)
GROUP BY bucket
ORDER BY count DESC, bucket ASC;

-- name: AnalyticsCVStatusMix :many
SELECT status AS bucket, COUNT(*) AS count
FROM (
    SELECT LOWER(COALESCE(a.cv_processing_status, 'pending')) AS status
      FROM voxire_career_applications a
      JOIN careers c ON c.id = a.job_id AND c.owner_id = ?
     WHERE a.created_at >= DATE_SUB(UTC_TIMESTAMP(), INTERVAL 12 MONTH)
       AND a.cv_key IS NOT NULL AND a.cv_key <> ''
    UNION ALL
    SELECT LOWER(COALESCE(cv_processing_status, 'pending')) AS status
      FROM voxhire_talent_applications
     WHERE created_at >= DATE_SUB(UTC_TIMESTAMP(), INTERVAL 12 MONTH)
       AND cv_key IS NOT NULL AND cv_key <> ''
) statuses
GROUP BY status
ORDER BY count DESC, status ASC;

-- name: AnalyticsTaskStatusMix :many
SELECT LOWER(COALESCE(NULLIF(status, ''), 'open')) AS bucket, COUNT(*) AS count
FROM voxhire_follow_up_tasks
JOIN voxire_career_applications a ON a.id = voxhire_follow_up_tasks.application_id
JOIN careers c ON c.id = a.job_id AND c.owner_id = ?
WHERE voxhire_follow_up_tasks.created_at >= DATE_SUB(UTC_TIMESTAMP(), INTERVAL 12 MONTH)
GROUP BY bucket
ORDER BY count DESC, bucket ASC;

-- name: AnalyticsLeadServiceMix :many
SELECT LOWER(COALESCE(NULLIF(service, ''), 'unsure')) AS bucket, COUNT(*) AS count
FROM voxhire_company_leads
WHERE created_at >= DATE_SUB(UTC_TIMESTAMP(), INTERVAL 12 MONTH)
GROUP BY bucket
ORDER BY count DESC, bucket ASC;

-- name: AnalyticsLeadTimelineMix :many
SELECT LOWER(COALESCE(NULLIF(timeline, ''), 'unknown')) AS bucket, COUNT(*) AS count
FROM voxhire_company_leads
WHERE created_at >= DATE_SUB(UTC_TIMESTAMP(), INTERVAL 12 MONTH)
GROUP BY bucket
ORDER BY count DESC, bucket ASC;

-- name: AnalyticsApplicationDailyCounts :many
SELECT DATE_FORMAT(a.created_at, '%Y-%m-%d') AS day, COUNT(*) AS count
FROM voxire_career_applications a
JOIN careers c ON c.id = a.job_id AND c.owner_id = ?
WHERE a.created_at >= DATE_SUB(UTC_DATE(), INTERVAL 13 DAY)
GROUP BY day;

-- name: AnalyticsLeadDailyCounts :many
SELECT DATE_FORMAT(created_at, '%Y-%m-%d') AS day, COUNT(*) AS count
FROM voxhire_company_leads
WHERE created_at >= DATE_SUB(UTC_DATE(), INTERVAL 13 DAY)
GROUP BY day;

-- name: AnalyticsTopOpenings :many
SELECT c.id,
       c.title,
       c.is_active,
       COALESCE(COUNT(a.id), 0) AS total_count,
       COALESCE(SUM(CASE WHEN LOWER(COALESCE(a.status, 'new')) = 'new' THEN 1 ELSE 0 END), 0) AS new_count,
       COALESCE(SUM(CASE WHEN a.created_at >= DATE_SUB(UTC_TIMESTAMP(), INTERVAL 30 DAY) THEN 1 ELSE 0 END), 0) AS recent_count,
       MAX(a.created_at) AS last_application_at
FROM careers c
LEFT JOIN voxire_career_applications a ON a.job_id = c.id
WHERE c.owner_id = ?
GROUP BY c.id, c.title, c.is_active, c.created_at
ORDER BY recent_count DESC, total_count DESC, c.created_at DESC
LIMIT 6;

-- name: AnalyticsRecentActivity :many
SELECT kind, id, primary_text, secondary_text, meta, created_at
FROM (
    SELECT 'application' AS kind,
           a.id AS id,
           c.title AS primary_text,
           a.name AS secondary_text,
           COALESCE(a.status, 'new') AS meta,
           a.created_at AS created_at
      FROM voxire_career_applications a
      JOIN careers c ON c.id = a.job_id AND c.owner_id = ?
    UNION ALL
    SELECT 'lead' AS kind,
           id,
           company_name AS primary_text,
           COALESCE(role, '') AS secondary_text,
           COALESCE(timeline, '') AS meta,
           created_at
      FROM voxhire_company_leads
) recent
ORDER BY created_at DESC, id DESC
LIMIT 10;
