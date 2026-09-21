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
        WHERE pm.project_id = p.id AND pm.user_id = ?)
    -- Dept-manager visibility proxy (M3). A caller who manages nobody is passed
    -- a single 0: sqlc expands an empty slice into an invalid IN (), and no user
    -- has id 0, so IN (0) matches nothing.
    --
    -- The surrounding WHERE already has p.is_personal = FALSE, which is what
    -- keeps a report's My Tasks bucket out, as Project.is_personal's contract
    -- requires.
    OR p.created_by IN (sqlc.slice('managed_creators'))
    OR p.owner_id IN (sqlc.slice('managed_owners'))
    OR EXISTS (
        SELECT 1 FROM flow_project_members pmr
        WHERE pmr.project_id = p.id AND pmr.user_id IN (sqlc.slice('managed_members'))))
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
        WHERE pm.project_id = p.id AND pm.user_id = ?)
    -- Dept-manager visibility proxy (M3). A caller who manages nobody is passed
    -- a single 0: sqlc expands an empty slice into an invalid IN (), and no user
    -- has id 0, so IN (0) matches nothing.
    --
    -- The surrounding WHERE already has p.is_personal = FALSE, which is what
    -- keeps a report's My Tasks bucket out, as Project.is_personal's contract
    -- requires.
    OR p.created_by IN (sqlc.slice('managed_creators'))
    OR p.owner_id IN (sqlc.slice('managed_owners'))
    OR EXISTS (
        SELECT 1 FROM flow_project_members pmr
        WHERE pmr.project_id = p.id AND pmr.user_id IN (sqlc.slice('managed_members'))))
  AND (? = '' OR p.status = ?)
  AND (? = '' OR p.visibility = ?)
  AND (? = '' OR p.name LIKE CONCAT('%', ?, '%'))
ORDER BY p.created_at DESC, p.id DESC
LIMIT ? OFFSET ?;

-- The two archived queries below are deliberately NOT widened for the
-- dept-manager proxy. The recycle bin is where projects are restored and
-- purged, so widening read there widens a destructive surface, and P9 asks for
-- a sidebar group of a report's active work. A manager who needs something
-- restored can ask its owner.
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
--
-- workspace_id is supplied as a literal 0 because the column is NOT NULL with
-- no default, unlike its siblings owner_id and business_id which both default
-- to 0. Omitting it made every v2 cycle creation fail outright with
-- "Error 1364: Field 'workspace_id' doesn't have a default value".
--
-- 0 is the same value a v2 project carries. The column is dropped in Phase 4,
-- at which point this literal comes out with it.
INSERT INTO flow_cycles (business_id, workspace_id, project_id, name, start_date, end_date)
VALUES (?, 0, ?, ?, ?, ?);

-- name: ListMilestonesByBusiness :many
SELECT m.id, m.project_id, m.name,
       m.target_date,
       COALESCE(m.status, 'upcoming') AS status,
       COALESCE(m.description, '') AS description,
       p.name AS project_name, p.color AS project_color,
       (SELECT COUNT(*) FROM flow_milestone_tasks mt WHERE mt.milestone_id = m.id) AS task_count,
       (SELECT COUNT(*) FROM flow_milestone_tasks mt JOIN flow_tasks t ON t.id = mt.task_id
        WHERE mt.milestone_id = m.id AND t.status = 'done') AS done_count
FROM flow_milestones m
LEFT JOIN flow_projects p ON p.id = m.project_id
WHERE m.business_id = ?
  AND (? = 0 OR m.project_id = ?)
  AND (? = 1 OR EXISTS (SELECT 1 FROM flow_projects p
        WHERE p.id = m.project_id
          AND (p.created_by = ? OR p.owner_id = ?
               OR EXISTS (SELECT 1 FROM flow_project_members pm
                          WHERE pm.project_id = p.id AND pm.user_id = ?))))
ORDER BY m.created_at DESC
LIMIT ? OFFSET ?;

-- name: CreateMilestoneV2 :execresult
INSERT INTO flow_milestones (business_id, project_id, name, description, target_date, status)
VALUES (?, ?, ?, ?, ?, 'planning');

-- name: ListViewsV2 :many
SELECT id, business_id, name, COALESCE(icon,'') AS icon,
       COALESCE(filters,'{}') AS filters,
       COALESCE(group_by,'') AS group_by,
       COALESCE(sort_by,'') AS sort_by,
       COALESCE(sort_dir,'asc') AS sort_dir,
       COALESCE(visibility,'shared') AS visibility
FROM flow_views
WHERE business_id = ?
  -- A private view is visible only to its creator (S2). Everything else is
  -- shared business-wide, which is how every view behaved before the column.
  AND (COALESCE(visibility,'shared') <> 'private' OR created_by = ?)
ORDER BY created_at DESC LIMIT ? OFFSET ?;

-- name: CreateViewV2 :execresult
INSERT INTO flow_views (business_id, name, icon, filters, group_by, sort_by, sort_dir, visibility, created_by)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);

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
