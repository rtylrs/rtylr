-- name: CountCycles :one
SELECT COUNT(*) FROM flow_cycles WHERE workspace_id = ?;

-- name: ListCycles :many
SELECT
    c.id, c.workspace_id, c.project_id, c.name, c.description, c.status,
    c.start_date, c.end_date, c.created_by, c.created_at,
    COUNT(ct.task_id) AS scope,
    COUNT(CASE WHEN t.status = 'done' THEN 1 END) AS completed
FROM flow_cycles c
LEFT JOIN flow_cycle_tasks ct ON ct.cycle_id = c.id
LEFT JOIN flow_tasks t ON t.id = ct.task_id
WHERE c.workspace_id = ?
GROUP BY c.id
ORDER BY c.start_date DESC
LIMIT ? OFFSET ?;

-- name: CreateCycle :execresult
INSERT INTO flow_cycles (owner_id, business_id, workspace_id, project_id, name, description, start_date, end_date, created_by)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: CountMilestonesByWorkspace :one
SELECT COUNT(*) FROM flow_milestones WHERE workspace_id = ?;

-- name: ListMilestonesByWorkspace :many
SELECT
    m.id, m.project_id, m.workspace_id, m.name, m.description, m.status,
    m.due_date, m.completed_at, m.created_by, m.created_at,
    p.name AS project_name, p.color AS project_color,
    COUNT(mt.task_id) AS task_count,
    COUNT(CASE WHEN t.status = 'done' THEN 1 END) AS done_count
FROM flow_milestones m
LEFT JOIN flow_projects p ON p.id = m.project_id
LEFT JOIN flow_milestone_tasks mt ON mt.milestone_id = m.id
LEFT JOIN flow_tasks t ON t.id = mt.task_id
WHERE m.workspace_id = ?
GROUP BY m.id
ORDER BY m.due_date ASC
LIMIT ? OFFSET ?;

-- name: CountMilestonesByProject :one
SELECT COUNT(*) FROM flow_milestones WHERE workspace_id = ? AND project_id = ?;

-- name: ListMilestonesByProject :many
SELECT
    m.id, m.project_id, m.workspace_id, m.name, m.description, m.status,
    m.due_date, m.completed_at, m.created_by, m.created_at,
    p.name AS project_name, p.color AS project_color,
    COUNT(mt.task_id) AS task_count,
    COUNT(CASE WHEN t.status = 'done' THEN 1 END) AS done_count
FROM flow_milestones m
LEFT JOIN flow_projects p ON p.id = m.project_id
LEFT JOIN flow_milestone_tasks mt ON mt.milestone_id = m.id
LEFT JOIN flow_tasks t ON t.id = mt.task_id
WHERE m.workspace_id = ? AND m.project_id = ?
GROUP BY m.id
ORDER BY m.due_date ASC
LIMIT ? OFFSET ?;

-- name: CreateMilestone :execresult
INSERT INTO flow_milestones (owner_id, business_id, project_id, workspace_id, name, description, due_date, created_by)
VALUES (?, ?, ?, ?, ?, ?, ?, ?);

-- name: CountViewsByWorkspace :one
SELECT COUNT(*) FROM flow_views WHERE business_id = ? AND workspace_id = ?;

-- name: ListViews :many
-- A private view is visible only to whoever created it (S2). Everything else is
-- shared business-wide, which is how every view behaved before the column
-- existed, so an unset visibility keeps its old reach.
--
-- created_by is nullable, and a private view with a NULL creator would be
-- reachable by nobody. That is the safe direction for a visibility filter: a
-- view nobody can see is recoverable, a private view leaked to colleagues is
-- not. CreateView always sets it.
SELECT id, workspace_id, name, icon, filters, group_by, sort_by, sort_dir,
       COALESCE(visibility, 'shared') AS visibility, created_by, created_at
FROM flow_views
WHERE business_id = ? AND workspace_id = ?
  AND (COALESCE(visibility, 'shared') <> 'private' OR created_by = ?)
ORDER BY created_at ASC
LIMIT ? OFFSET ?;

-- name: CreateView :execresult
INSERT INTO flow_views (owner_id, business_id, workspace_id, name, icon, filters, group_by, sort_by, sort_dir, visibility, created_by)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: UpdateView :execresult
UPDATE flow_views
SET name = ?, icon = ?, filters = ?, group_by = ?, sort_by = ?, sort_dir = ?, visibility = ?
WHERE flow_views.id = ?
  AND flow_views.business_id = ?
  AND (
    EXISTS (SELECT 1 FROM flow_workspaces w WHERE w.id = flow_views.workspace_id AND w.owner_id = ?)
    OR EXISTS (SELECT 1 FROM flow_workspace_members wm WHERE wm.workspace_id = flow_views.workspace_id AND wm.user_id = ?)
  );

-- name: DeleteView :execresult
DELETE FROM flow_views
WHERE flow_views.id = ?
  AND flow_views.business_id = ?
  AND (
    EXISTS (SELECT 1 FROM flow_workspaces w WHERE w.id = flow_views.workspace_id AND w.owner_id = ?)
    OR EXISTS (SELECT 1 FROM flow_workspace_members wm WHERE wm.workspace_id = flow_views.workspace_id AND wm.user_id = ?)
  );

-- name: GetInsightsTaskCounts :one
SELECT
    COUNT(*) AS total_tasks,
    COALESCE(SUM(status = 'done'), 0) AS done_tasks
FROM flow_tasks
WHERE workspace_id = ? AND parent_task_id IS NULL;

-- name: GetInsightsCycleTime :one
SELECT AVG(DATEDIFF(completed_at, created_at)) AS avg_days
FROM flow_tasks
WHERE workspace_id = ? AND status = 'done' AND completed_at IS NOT NULL;

-- name: GetInsightsVelocity :many
SELECT
    DATE_FORMAT(DATE_SUB(completed_at, INTERVAL WEEKDAY(completed_at) DAY), '%Y-%m-%d') AS week_start,
    COUNT(*) AS completed
FROM flow_tasks
WHERE workspace_id = ?
  AND status = 'done'
  AND completed_at >= DATE_SUB(NOW(), INTERVAL 8 WEEK)
GROUP BY YEARWEEK(completed_at, 1)
ORDER BY week_start ASC;

-- name: GetInsightsOpenTaskCount :one
SELECT COUNT(*) AS open_tasks
FROM flow_tasks
WHERE workspace_id = ? AND status NOT IN ('done', 'cancelled') AND parent_task_id IS NULL;

-- cross-schema read from users and hr_time_entries — insights time-by-member
-- aggregation over BOTH stores of hours.
--
-- Flow's hours live in two tables. Work-day lines are hr_time_entries rows
-- carrying a task_id; the task detail time panel writes flow_time_logs. This
-- summed the second alone, so a workspace whose members use the work-day screen
-- reported zero hours for everyone (tryrtylr/rtylr#903).
--
-- flow_time_logs rows with time_entry_id set are excluded: that column names the
-- hr_time_entries row they duplicate, and counting both sides would double the
-- same hour. Nothing writes it yet, so today this is a plain union of the two.
--
-- Restricted to source='flow' rows, which is a correctness requirement and not
-- a convenience. hr_time_entries.employee_id is not one id space: Flow's
-- CreateWorkDayLine stores the signed-in users.id there (worktime_service.go
-- scope() returns authUser.UserID as the employee), while hr_time writes entries
-- whose employee_id is an hr_employees.id, and hr_employees.user_id is nullable.
-- Joining users on both would silently drop HR-authored rows at best and
-- attribute them to the wrong person at worst.
--
-- The cost is that a time entry created in HR against a Flow task is not counted
-- here. Fixing that means resolving hr_employees.user_id, which is a join this
-- query cannot make without hr_employees, so it is left for #903.
-- name: GetInsightsTimeByMember :many
SELECT
    CAST(u.id AS CHAR) AS member_id,
    COALESCE(u.full_name, u.email) AS name,
    SUM(e.duration_minutes) / 60.0 AS hours
FROM (
    SELECT te.employee_id AS user_id, te.task_id, te.duration_minutes
    FROM hr_time_entries te
    WHERE te.task_id IS NOT NULL AND te.source = 'flow'
    UNION ALL
    SELECT tl.user_id, tl.task_id, tl.duration_minutes
    FROM flow_time_logs tl
    WHERE tl.time_entry_id IS NULL
) AS e
JOIN users u ON u.id = e.user_id
JOIN flow_tasks t ON t.id = e.task_id
WHERE t.workspace_id = ?
GROUP BY u.id, u.full_name, u.email
ORDER BY SUM(e.duration_minutes) DESC
LIMIT 10;

-- name: CountWorkspaceMembersForWorkload :one
SELECT COUNT(*) FROM flow_workspace_members WHERE workspace_id = ?;

-- cross-schema read from users — team workload member summary
-- name: GetTeamWorkload :many
SELECT
    CAST(u.id AS CHAR) AS member_id,
    COALESCE(u.full_name, u.email) AS name,
    COUNT(t.id) AS task_count
FROM flow_workspace_members wm
JOIN users u ON u.id = wm.user_id
LEFT JOIN flow_tasks t ON t.assignee_id = u.id
    AND t.workspace_id = ?
    AND t.status NOT IN ('done', 'cancelled')
    AND t.parent_task_id IS NULL
WHERE wm.workspace_id = ?
GROUP BY u.id, u.full_name, u.email
ORDER BY task_count DESC
LIMIT ? OFFSET ?;

-- Per-member fan-out from GetTeamWorkload, one call per member on the current
-- page. Intentionally capped at 10 - a workload card shows a preview of a
-- member's open work, not their full backlog. The outer member list is what's
-- paginated; this inner cap is fixed regardless of page size.
-- name: GetTeamWorkloadMemberTasks :many
SELECT
    id, project_id, workspace_id, owner_id, parent_task_id, section_id,
    title, description, status, priority,
    assignee_id, reporter_id, due_date, start_date, completed_at,
    estimated_hours, actual_hours, sort_order, labels, archived,
    created_at, updated_at
FROM flow_tasks
WHERE workspace_id = ? AND assignee_id = ? AND status NOT IN ('done', 'cancelled') AND parent_task_id IS NULL
ORDER BY priority DESC, due_date ASC
LIMIT 10;

-- name: CountRoadmapProjects :one
SELECT COUNT(*) FROM flow_projects WHERE workspace_id = ? AND status != 'archived';

-- name: GetRoadmap :many
SELECT
    p.id, p.workspace_id, p.owner_id, p.name, p.description, p.icon, p.color,
    p.status, p.start_date, p.due_date, p.created_by, p.created_at, p.updated_at
FROM flow_projects p
WHERE p.workspace_id = ? AND p.status != 'archived'
ORDER BY p.start_date ASC, p.created_at ASC
LIMIT ? OFFSET ?;

-- name: CountNotifications :one
SELECT COUNT(*) FROM flow_notifications WHERE user_id = ?;

-- name: ListNotifications :many
SELECT
    id, user_id, workspace_id, type, title, COALESCE(description, '') AS description,
    task_id, COALESCE(task_title, '') AS task_title, actor_id, COALESCE(actor_initials, '') AS actor_initials,
    COALESCE(project, '') AS project, COALESCE(project_color, '') AS project_color,
    is_read, created_at
FROM flow_notifications
WHERE user_id = ?
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: CreateNotification :exec
INSERT INTO flow_notifications
    (user_id, workspace_id, type, title, description, task_id, task_title, actor_id, actor_initials, project, project_color)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: MarkNotificationRead :exec
UPDATE flow_notifications
SET is_read = TRUE
WHERE id = ? AND user_id = ?;

-- name: MarkAllNotificationsRead :exec
UPDATE flow_notifications
SET is_read = TRUE
WHERE user_id = ?;

-- name: GetUserSettings :one
SELECT task_assigned, task_comments, mentions, status_changes, weekly_digest, email_notifications, sidebar_preferences_json
FROM flow_user_settings
WHERE user_id = ?;

-- name: UpsertUserSettings :exec
INSERT INTO flow_user_settings (user_id, task_assigned, task_comments, mentions, status_changes, weekly_digest, email_notifications, sidebar_preferences_json)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
    task_assigned = VALUES(task_assigned),
    task_comments = VALUES(task_comments),
    mentions = VALUES(mentions),
    status_changes = VALUES(status_changes),
    weekly_digest = VALUES(weekly_digest),
    email_notifications = VALUES(email_notifications),
    sidebar_preferences_json = VALUES(sidebar_preferences_json);

-- name: GetProjectTaskCounts :one
SELECT
    COUNT(*) AS total,
    COALESCE(SUM(status = 'done'), 0) AS done
FROM flow_tasks
WHERE project_id = ? AND parent_task_id IS NULL;

-- ── v2: business/project-scoped replacements ─────────────────────────────

-- name: ListCyclesV2 :many
SELECT c.id, c.project_id, c.name, c.description, c.status,
       c.start_date, c.end_date, c.created_by, c.created_at,
       (SELECT COUNT(*) FROM flow_cycle_tasks ct WHERE ct.cycle_id = c.id) AS scope,
       (SELECT COUNT(*) FROM flow_cycle_tasks ct JOIN flow_tasks t ON t.id = ct.task_id
        WHERE ct.cycle_id = c.id AND t.status = 'done') AS completed
FROM flow_cycles c
WHERE c.project_id = ? AND c.business_id = ?
  AND (? = 1 OR EXISTS (SELECT 1 FROM flow_projects p
        WHERE p.id = c.project_id
          AND (p.created_by = ? OR p.owner_id = ?
               OR EXISTS (SELECT 1 FROM flow_project_members pm
                          WHERE pm.project_id = p.id AND pm.user_id = ?))))
ORDER BY c.start_date DESC
LIMIT ? OFFSET ?;

-- name: CountCyclesV2 :one
SELECT COUNT(*) FROM flow_cycles c
WHERE c.project_id = ? AND c.business_id = ?
  AND (? = 1 OR EXISTS (SELECT 1 FROM flow_projects p
        WHERE p.id = c.project_id
          AND (p.created_by = ? OR p.owner_id = ?
               OR EXISTS (SELECT 1 FROM flow_project_members pm
                          WHERE pm.project_id = p.id AND pm.user_id = ?))));

-- name: GetInsightsTaskCountsV2 :one
SELECT
    COUNT(*) AS total_tasks,
    COALESCE(SUM(status = 'done'), 0) AS done_tasks
FROM flow_tasks WHERE business_id = ?;

-- name: GetInsightsCycleTimeV2 :one
SELECT AVG(DATEDIFF(completed_at, created_at)) AS avg_days
FROM flow_tasks
WHERE business_id = ? AND status = 'done' AND completed_at IS NOT NULL;

-- name: GetInsightsVelocityV2 :many
SELECT
    DATE_FORMAT(DATE_SUB(completed_at, INTERVAL WEEKDAY(completed_at) DAY), '%Y-%m-%d') AS week_start,
    COUNT(*) AS completed
FROM flow_tasks
WHERE business_id = ? AND status = 'done' AND completed_at IS NOT NULL
GROUP BY week_start
ORDER BY week_start DESC;

-- name: GetInsightsOpenTaskCountV2 :one
SELECT COUNT(*) AS open_tasks
FROM flow_tasks
WHERE business_id = ? AND status NOT IN ('done', 'cancelled') AND parent_task_id IS NULL;

-- name: GetRoadmapV2 :many
SELECT p.id, p.owner_id, p.name, p.description, p.icon, p.color,
       p.status, p.start_date, p.due_date, p.created_by, p.created_at, p.updated_at
FROM flow_projects p
WHERE p.business_id = ? AND p.is_personal = FALSE AND p.archived_at IS NULL AND p.status != 'archived'
ORDER BY p.due_date ASC, p.created_at DESC
LIMIT ? OFFSET ?;

-- name: CountRoadmapV2 :one
SELECT COUNT(*) FROM flow_projects
WHERE business_id = ? AND is_personal = FALSE AND archived_at IS NULL AND status != 'archived';
