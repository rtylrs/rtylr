-- V2 business-scoped project queries for the workspace-removal contracts.
-- Curation rule: without bypass, callers see only projects they created or are
-- members of; the My Tasks bucket is always excluded from shared lists because
-- GetMyTasksProject addresses it directly.

-- name: GetProjectV2ByID :one
SELECT
    p.id, p.workspace_id, p.owner_id, p.business_id, p.name,
    COALESCE(p.description, '') AS description,
    COALESCE(p.icon, '') AS icon,
    COALESCE(p.color, '') AS color,
    COALESCE(p.status, 'active') AS status,
    COALESCE(p.visibility, 'private') AS visibility,
    p.is_personal,
    p.start_date, p.due_date,
    p.created_by, p.created_at, p.updated_at,
    p.archived_at, p.purge_at,
    (SELECT COUNT(*) FROM flow_tasks t WHERE t.project_id = p.id) AS task_count,
    (SELECT COUNT(*) FROM flow_tasks t WHERE t.project_id = p.id AND t.status = 'done') AS completed_count,
    (SELECT COUNT(*) FROM flow_project_members pm WHERE pm.project_id = p.id) AS member_count
FROM flow_projects p
WHERE p.id = ? AND p.business_id = ?;

-- name: CountProjectsV2 :one
SELECT COUNT(*) FROM flow_projects p
WHERE p.business_id = ?
  AND p.is_personal = FALSE
  AND p.archived_at IS NULL
  AND (? = 1 OR p.created_by = ? OR EXISTS (
        SELECT 1 FROM flow_project_members pm
        WHERE pm.project_id = p.id AND pm.user_id = ?))
  AND (? = '' OR p.status = ?)
  AND (? = '' OR p.visibility = ?)
  AND (? = '' OR p.name LIKE CONCAT('%', ?, '%'));

-- name: ListProjectsV2 :many
SELECT
    p.id, p.workspace_id, p.owner_id, p.business_id, p.name,
    COALESCE(p.description, '') AS description,
    COALESCE(p.icon, '') AS icon,
    COALESCE(p.color, '') AS color,
    COALESCE(p.status, 'active') AS status,
    COALESCE(p.visibility, 'private') AS visibility,
    p.is_personal,
    p.start_date, p.due_date,
    p.created_by, p.created_at, p.updated_at,
    p.archived_at, p.purge_at,
    (SELECT COUNT(*) FROM flow_tasks t WHERE t.project_id = p.id) AS task_count,
    (SELECT COUNT(*) FROM flow_tasks t WHERE t.project_id = p.id AND t.status = 'done') AS completed_count,
    (SELECT COUNT(*) FROM flow_project_members pm WHERE pm.project_id = p.id) AS member_count
FROM flow_projects p
WHERE p.business_id = ?
  AND p.is_personal = FALSE
  AND p.archived_at IS NULL
  AND (? = 1 OR p.created_by = ? OR EXISTS (
        SELECT 1 FROM flow_project_members pm
        WHERE pm.project_id = p.id AND pm.user_id = ?))
  AND (? = '' OR p.status = ?)
  AND (? = '' OR p.visibility = ?)
  AND (? = '' OR p.name LIKE CONCAT('%', ?, '%'))
ORDER BY p.created_at DESC, p.id DESC
LIMIT ? OFFSET ?;

-- name: CountArchivedProjectsV2 :one
SELECT COUNT(*) FROM flow_projects p
WHERE p.business_id = ?
  AND p.archived_at IS NOT NULL
  AND (? = 1 OR p.created_by = ? OR EXISTS (
        SELECT 1 FROM flow_project_members pm
        WHERE pm.project_id = p.id AND pm.user_id = ?));

-- name: ListArchivedProjectsV2 :many
SELECT
    p.id, p.workspace_id, p.owner_id, p.business_id, p.name,
    COALESCE(p.description, '') AS description,
    COALESCE(p.icon, '') AS icon,
    COALESCE(p.color, '') AS color,
    COALESCE(p.status, 'active') AS status,
    COALESCE(p.visibility, 'private') AS visibility,
    p.is_personal,
    p.start_date, p.due_date,
    p.created_by, p.created_at, p.updated_at,
    p.archived_at, p.purge_at,
    (SELECT COUNT(*) FROM flow_tasks t WHERE t.project_id = p.id) AS task_count,
    (SELECT COUNT(*) FROM flow_tasks t WHERE t.project_id = p.id AND t.status = 'done') AS completed_count,
    (SELECT COUNT(*) FROM flow_project_members pm WHERE pm.project_id = p.id) AS member_count
FROM flow_projects p
WHERE p.business_id = ?
  AND p.archived_at IS NOT NULL
  AND (? = 1 OR p.created_by = ? OR EXISTS (
        SELECT 1 FROM flow_project_members pm
        WHERE pm.project_id = p.id AND pm.user_id = ?))
ORDER BY p.purge_at ASC, p.id DESC
LIMIT ? OFFSET ?;

-- name: UpdateProjectVisibility :execresult
UPDATE flow_projects SET visibility = ?, updated_at = NOW()
WHERE id = ? AND business_id = ?;

-- name: RestoreProjectV2 :execresult
UPDATE flow_projects SET status = 'active', archived_at = NULL, purge_at = NULL, updated_at = NOW()
WHERE id = ? AND business_id = ?;

-- name: GetDashboardStatsV2 :one
-- Business-scoped dashboard counts: no workspace dimension.
SELECT
    COUNT(CASE WHEN status = 'todo' THEN 1 END) AS open_tasks,
    COUNT(CASE WHEN status = 'in_progress' THEN 1 END) AS in_progress_tasks,
    COUNT(CASE WHEN status = 'done' AND completed_at >= ? AND completed_at < ? THEN 1 END) AS completed_this_week,
    COUNT(CASE WHEN status NOT IN ('done','cancelled') AND due_date < NOW() THEN 1 END) AS overdue_tasks
FROM flow_tasks WHERE business_id = ?;

-- name: ListLabelsV2 :many
SELECT id, business_id, name, color, created_at
FROM flow_labels WHERE business_id = ?
ORDER BY name ASC LIMIT ? OFFSET ?;

-- name: CountLabelsV2 :one
SELECT COUNT(*) FROM flow_labels WHERE business_id = ?;

-- name: CreateLabelV2 :execresult
INSERT INTO flow_labels (business_id, name, color) VALUES (?, ?, ?);

-- name: GetInsightsTimeByMemberV2 :many
SELECT CAST(u.id AS CHAR) AS member_id,
       COALESCE(u.full_name, u.email) AS name,
       COALESCE(SUM(tl.duration_minutes) / 60.0, 0) AS hours
FROM flow_time_logs tl
JOIN users u ON u.id = tl.user_id
WHERE tl.business_id = ?
GROUP BY u.id, u.full_name, u.email
ORDER BY hours DESC;

-- name: CreateCycleV2 :execresult
INSERT INTO flow_cycles (business_id, project_id, name, start_date, end_date)
VALUES (?, ?, ?, ?, ?);

-- name: ListMilestonesByBusiness :many
SELECT m.id, m.project_id, m.name,
       COALESCE(m.target_date, '') AS target_date,
       COALESCE(m.status, 'upcoming') AS status,
       COALESCE(m.description, '') AS description,
       (SELECT COUNT(*) FROM flow_milestone_tasks mt WHERE mt.milestone_id = m.id) AS task_count,
       (SELECT COUNT(*) FROM flow_milestone_tasks mt JOIN flow_tasks t ON t.id = mt.task_id
        WHERE mt.milestone_id = m.id AND t.status = 'done') AS done_count
FROM flow_milestones m
WHERE m.business_id = ?
  AND (? = 0 OR m.project_id = ?)
ORDER BY m.created_at DESC
LIMIT ? OFFSET ?;

-- name: CreateMilestoneV2 :execresult
INSERT INTO flow_milestones (business_id, project_id, name, description, status)
VALUES (?, ?, ?, ?, 'planning');

-- name: ListViewsV2 :many
SELECT id, business_id, name, COALESCE(icon,'') AS icon,
       COALESCE(filters,'{}') AS filters,
       COALESCE(group_by,'') AS group_by,
       COALESCE(sort_by,'') AS sort_by,
       COALESCE(sort_dir,'asc') AS sort_dir
FROM flow_views WHERE business_id = ?
ORDER BY created_at DESC LIMIT ? OFFSET ?;

-- name: CreateViewV2 :execresult
INSERT INTO flow_views (business_id, name, icon, filters, group_by, sort_by, sort_dir, created_by)
VALUES (?, ?, ?, ?, ?, ?, ?, ?);

-- ListPublicProjectsByOwnerV2 backs the profile page. Deliberately narrower
-- than ListProjectsV2: it returns metadata only, never task-level fields,
-- because "public" means discoverable, not openable. Membership is not
-- consulted at all - a public project shows on its owner's profile to every
-- colleague in the business, and opening it still goes through the normal
-- project gate.
--
-- name: ListPublicProjectsByOwnerV2 :many
SELECT
    p.id, p.name,
    COALESCE(p.icon, '') AS icon,
    COALESCE(p.color, '') AS color,
    COALESCE(p.description, '') AS description,
    COALESCE(p.status, 'active') AS status,
    (SELECT COUNT(*) FROM flow_tasks t WHERE t.project_id = p.id) AS task_count,
    (SELECT COUNT(*) FROM flow_tasks t WHERE t.project_id = p.id AND t.status = 'done') AS completed_count
FROM flow_projects p
WHERE p.business_id = ?
  AND p.created_by = ?
  AND p.visibility = 'public'
  AND p.is_personal = FALSE
  AND p.archived_at IS NULL
ORDER BY p.created_at DESC, p.id DESC
LIMIT ?;

-- name: GetProjectProxyFacts :one
-- The four facts the dept-manager visibility proxy needs about a project,
-- looked up by id alone. projectRoleFor resolves a role without knowing the
-- business, and the proxy must not trust a business passed in from a caller.
SELECT p.business_id, p.owner_id, p.created_by, p.is_personal
FROM flow_projects p
WHERE p.id = ?;
