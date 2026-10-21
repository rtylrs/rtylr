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

-- name: GetAttachmentTaskID :one
SELECT task_id
FROM flow_task_attachments
WHERE id = ?
LIMIT 1;

-- Self-scoped like the delete itself: only the author's own time log can
-- match, so an admin-tier caller deleting someone else's log reports the
-- denial instead of a success that deleted nothing. :execresult is what lets
-- the repository tell the two apart via requireAffected.
-- name: DeleteTimeLog :execresult
DELETE FROM flow_time_logs
WHERE id = ? AND user_id = ?;

-- name: GetTaskBusinessID :one
SELECT w.business_id
FROM flow_tasks t
JOIN flow_workspaces w ON w.id = t.workspace_id
WHERE t.id = ?
LIMIT 1;

-- name: SumTaskTimeLogMinutes :one
SELECT CAST(COALESCE(SUM(duration_minutes), 0) AS SIGNED)
FROM flow_time_logs
WHERE task_id = ? AND time_entry_id IS NULL;

-- Flow stores its time logs locally. HR Time provides the work-day entry
-- aggregate through the injected owner port; Flow only persists the resulting
-- task projection here.
-- name: RecalcActualHours :exec
UPDATE flow_tasks
SET actual_hours = ?
WHERE flow_tasks.id = ?
  AND (
    EXISTS (SELECT 1 FROM flow_project_members pm
            WHERE pm.project_id = flow_tasks.project_id AND pm.user_id = ?)
    OR EXISTS (SELECT 1 FROM flow_projects p
               WHERE p.id = flow_tasks.project_id
                 AND (p.owner_id = ? OR p.created_by = ?))
    OR EXISTS (SELECT 1 FROM flow_workspaces w WHERE w.id = flow_tasks.workspace_id AND w.owner_id = ?)
    OR EXISTS (SELECT 1 FROM flow_workspace_members wm WHERE wm.workspace_id = flow_tasks.workspace_id AND wm.user_id = ?)
  );

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
