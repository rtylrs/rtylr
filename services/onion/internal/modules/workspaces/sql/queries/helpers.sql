-- name: GetCommentTaskID :one
SELECT task_id
FROM flow_task_comments
WHERE id = ?
LIMIT 1;

-- name: GetTimeLogTaskID :one
SELECT task_id
FROM flow_time_logs
WHERE id = ?
LIMIT 1;

-- name: DeleteTimeLog :exec
DELETE FROM flow_time_logs
WHERE id = ? AND user_id = ?;

-- name: RecalcActualHours :exec
UPDATE flow_tasks
SET actual_hours = (
    SELECT COALESCE(SUM(duration_minutes), 0) / 60.0
    FROM flow_time_logs
    WHERE task_id = ?
)
WHERE flow_tasks.id = ?
  AND (
    EXISTS (SELECT 1 FROM flow_workspaces w WHERE w.id = flow_tasks.workspace_id AND w.owner_id = ?)
    OR EXISTS (SELECT 1 FROM flow_workspace_members wm WHERE wm.workspace_id = flow_tasks.workspace_id AND wm.user_id = ?)
  );

-- cross-schema read from users — email lookup for project invites
-- name: LookupUserByEmail :one
SELECT id
FROM users
WHERE email = ?
LIMIT 1;

-- name: IsWorkspaceMember :one
SELECT EXISTS(
    SELECT 1 FROM flow_workspaces w
    WHERE w.id = ? AND w.owner_id = ?
) OR EXISTS(
    SELECT 1 FROM flow_workspace_members wm
    WHERE wm.workspace_id = ? AND wm.user_id = ?
) AS is_member;

-- name: GetMyTasksProject :one
SELECT id, name
FROM flow_projects
WHERE owner_id = ? AND workspace_id = ? AND name IN ('My Tasks', 'Inbox')
ORDER BY (name = 'My Tasks') DESC, created_at ASC
LIMIT 1;

-- name: RenameProject :exec
UPDATE flow_projects
SET name = ?
WHERE id = ? AND owner_id = ? AND workspace_id = ?;

-- name: GetProjectWorkspace :one
SELECT workspace_id
FROM flow_projects
WHERE id = ?
LIMIT 1;

-- name: GetColumnProject :one
SELECT project_id
FROM flow_project_columns
WHERE id = ?
LIMIT 1;

-- name: IsProjectMember :one
SELECT EXISTS(
    SELECT 1 FROM flow_project_members pm
    WHERE pm.project_id = ? AND pm.user_id = ?
) OR EXISTS(
    SELECT 1
    FROM flow_projects p
    LEFT JOIN flow_workspace_members wm
      ON wm.workspace_id = p.workspace_id AND wm.user_id = ?
    WHERE p.id = ? AND (p.owner_id = ? OR wm.user_id IS NOT NULL)
) AS is_member;
