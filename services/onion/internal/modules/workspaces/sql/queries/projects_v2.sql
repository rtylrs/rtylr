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
