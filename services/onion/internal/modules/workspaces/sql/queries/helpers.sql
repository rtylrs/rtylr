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

-- A task's actual hours come from BOTH stores of hours, not just Flow's own.
--
-- hr_time_entries rows carry a task_id, so a work-day line is attributable to
-- the task it was logged against. Summing flow_time_logs alone meant every
-- work-day line counted as zero here, which is what tryrtylr/rtylr#903 reports.
--
-- flow_time_logs rows with time_entry_id set are skipped: that column names the
-- hr_time_entries row they duplicate, so counting both sides would double the
-- same hour. Nothing writes it yet, so today this is a plain sum of the two.
--
-- Restricted to source='flow'. A task's hours are Flow's own; an HR-authored
-- entry that happens to carry a task_id belongs to HR's timesheet flow and its
-- employee_id is an hr_employees.id rather than a users.id. See the note on
-- GetInsightsTimeByMember in features.sql.
--
-- name: RecalcActualHours :exec
UPDATE flow_tasks
SET actual_hours = (
    SELECT COALESCE(SUM(minutes), 0) / 60.0
    FROM (
        SELECT te.duration_minutes AS minutes
        FROM hr_time_entries te
        WHERE te.task_id = ? AND te.source = 'flow'
        UNION ALL
        SELECT tl.duration_minutes AS minutes
        FROM flow_time_logs tl
        WHERE tl.task_id = ? AND tl.time_entry_id IS NULL
    ) AS both_stores
)
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

-- cross-schema read from users — email lookup for project invites
-- name: LookupUserByEmail :one
SELECT id
FROM users
WHERE email = ?
LIMIT 1;

-- cross-schema read from users and user_business_access: resolve a mention
-- inside the task's business without exposing same-named users from another
-- tenant
-- name: LookupUserByUsername :one
SELECT u.id
FROM users u
WHERE u.username = ?
  AND (
      u.business_id = ?
      OR EXISTS (
          SELECT 1
          FROM user_business_access uba
          WHERE uba.user_id = u.id AND uba.business_id = ?
      )
  )
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
