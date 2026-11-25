-- name: GetTotalAndCompletedTasks :one
SELECT
    COUNT(*) AS total_tasks,
    COUNT(CASE WHEN status = 'done' THEN 1 END) AS completed_tasks,
    COUNT(CASE WHEN status NOT IN ('done','cancelled') AND due_date < CURDATE() THEN 1 END) AS overdue_count
FROM flow_tasks WHERE business_id = ?;

-- name: GetAvgCompletionDays :one
SELECT COALESCE(AVG(DATEDIFF(completed_at, created_at)), 0) AS avg_days
FROM flow_tasks WHERE business_id = ? AND status = 'done' AND completed_at IS NOT NULL;

-- name: GetVelocityThisWeek :one
SELECT COUNT(*) AS cnt
FROM flow_tasks
WHERE business_id = ? AND status = 'done'
  AND completed_at >= DATE_SUB(CURDATE(), INTERVAL WEEKDAY(CURDATE()) DAY)
  AND completed_at < DATE_ADD(DATE_SUB(CURDATE(), INTERVAL WEEKDAY(CURDATE()) DAY), INTERVAL 7 DAY);

-- name: GetVelocityLastWeek :one
SELECT COUNT(*) AS cnt
FROM flow_tasks
WHERE business_id = ? AND status = 'done'
  AND completed_at >= DATE_SUB(DATE_SUB(CURDATE(), INTERVAL WEEKDAY(CURDATE()) DAY), INTERVAL 7 DAY)
  AND completed_at < DATE_SUB(CURDATE(), INTERVAL WEEKDAY(CURDATE()) DAY);

-- name: GetTasksByStatus :many
SELECT status, COUNT(*) AS cnt
FROM flow_tasks WHERE business_id = ?
  AND created_at >= DATE_SUB(UTC_TIMESTAMP(), INTERVAL 12 MONTH)
GROUP BY status;

-- name: GetTasksByPriority :many
SELECT priority, COUNT(*) AS cnt
FROM flow_tasks
WHERE business_id = ?
  AND created_at >= DATE_SUB(UTC_TIMESTAMP(), INTERVAL 12 MONTH)
GROUP BY priority;

-- name: GetTeamWorkload :many
SELECT
    ft.assignee_id AS user_id,
    COUNT(*) AS total_tasks,
    COUNT(CASE WHEN ft.status = 'done' THEN 1 END) AS completed_tasks,
    COUNT(CASE WHEN ft.status = 'in_progress' THEN 1 END) AS in_progress_tasks,
    COUNT(CASE WHEN ft.status NOT IN ('done','cancelled') AND ft.due_date < CURDATE() THEN 1 END) AS overdue_tasks
FROM flow_tasks ft
WHERE ft.business_id = ?
  AND ft.created_at >= DATE_SUB(UTC_TIMESTAMP(), INTERVAL 12 MONTH)
  AND ft.assignee_id IS NOT NULL
GROUP BY ft.assignee_id;

-- Progress as in GetProjectTaskProgress, over tasks created in the last year.
-- name: GetRoadmapProjects :many
SELECT
    fp.id,
    fp.name,
    fp.status,
    fp.visibility,
    COALESCE(DATE_FORMAT(fp.start_date, '%Y-%m-%d'), '') AS start_date,
    COALESCE(DATE_FORMAT(fp.due_date, '%Y-%m-%d'), '') AS due_date,
    COUNT(CASE WHEN ft.id IS NOT NULL AND COALESCE(fs.bucket, ft.status, 'todo') <> 'cancelled' THEN 1 END) AS total_tasks,
    COUNT(CASE WHEN COALESCE(fs.bucket, ft.status) = 'done' THEN 1 END) AS done_tasks
FROM flow_projects fp
LEFT JOIN flow_tasks ft ON ft.project_id = fp.id
 AND ft.parent_task_id IS NULL
 AND COALESCE(ft.archived, 0) = 0
 AND ft.created_at >= DATE_SUB(UTC_TIMESTAMP(), INTERVAL 12 MONTH)
LEFT JOIN flow_statuses fs ON fs.id = ft.status_id AND fs.project_id = ft.project_id
WHERE fp.business_id = ? AND fp.status != 'archived'
GROUP BY fp.id, fp.name, fp.status, fp.visibility, fp.start_date, fp.due_date
ORDER BY fp.created_at ASC;

-- name: GetMilestonesByProjectIDs :many
SELECT
    fm.project_id,
    fm.id,
    fm.name,
    COALESCE(DATE_FORMAT(fm.target_date, '%Y-%m-%d'), '') AS due_date,
    COUNT(ft.id) AS task_count,
    COUNT(CASE WHEN ft.status = 'done' THEN 1 END) AS done_count
FROM flow_milestones fm
LEFT JOIN flow_milestone_tasks fmt ON fmt.milestone_id = fm.id
LEFT JOIN flow_tasks ft ON ft.id = fmt.task_id AND ft.parent_task_id IS NULL
WHERE fm.business_id = ? AND fm.project_id IN (sqlc.slice('project_ids'))
GROUP BY fm.project_id, fm.id, fm.name, fm.target_date
ORDER BY fm.target_date ASC;

-- Flow's own time logs only: rows linked to an HR time entry are HR's time.
-- Tenancy comes from the task and project, since old log rows can carry
-- business_id 0. Visibility and status let the caller name only public projects.
-- name: GetLoggedMinutesByProject :many
SELECT
    fp.id AS project_id,
    fp.name,
    fp.visibility,
    fp.status,
    CAST(COALESCE(SUM(tl.duration_minutes), 0) AS SIGNED) AS minutes
FROM flow_time_logs tl
JOIN flow_tasks ft ON ft.id = tl.task_id
JOIN flow_projects fp ON fp.id = ft.project_id
WHERE ft.business_id = sqlc.arg('business_id') AND fp.business_id = sqlc.arg('business_id')
  AND tl.time_entry_id IS NULL
  AND tl.logged_at >= sqlc.arg('date_from') AND tl.logged_at <= sqlc.arg('date_to')
GROUP BY fp.id, fp.name, fp.visibility, fp.status
ORDER BY minutes DESC, fp.id ASC;

-- The projects a user may open a project dashboard for: every public project
-- in the business, and the private ones the user owns, created or holds a
-- flow_project_members row on (admin, member or viewer). Personal My Tasks
-- projects and archived projects are left out.
-- name: ListDashboardProjectsForUser :many
SELECT
    fp.id,
    fp.name,
    COALESCE(fp.status, 'active') AS status,
    fp.visibility
FROM flow_projects fp
WHERE fp.business_id = sqlc.arg('business_id')
  AND fp.is_personal = FALSE
  AND fp.archived_at IS NULL
  AND COALESCE(fp.status, 'active') <> 'archived'
  AND (fp.visibility = 'public'
       OR fp.owner_id = sqlc.arg('user_id')
       OR fp.created_by = sqlc.arg('user_id')
       OR EXISTS (SELECT 1 FROM flow_project_members pm
                  WHERE pm.project_id = fp.id AND pm.user_id = sqlc.arg('user_id')))
ORDER BY fp.name ASC, fp.id ASC
LIMIT 500;

-- One project under the same rule as ListDashboardProjectsForUser, so a
-- private project the user does not belong to and a project in another
-- business both come back as no row. Archived projects stay reachable by id.
-- name: GetDashboardProjectForUser :one
SELECT
    fp.id,
    fp.name,
    COALESCE(fp.status, 'active') AS status,
    fp.visibility
FROM flow_projects fp
WHERE fp.id = sqlc.arg('project_id')
  AND fp.business_id = sqlc.arg('business_id')
  AND fp.is_personal = FALSE
  AND (fp.visibility = 'public'
       OR fp.owner_id = sqlc.arg('user_id')
       OR fp.created_by = sqlc.arg('user_id')
       OR EXISTS (SELECT 1 FROM flow_project_members pm
                  WHERE pm.project_id = fp.id AND pm.user_id = sqlc.arg('user_id')));

-- One project by id with no viewer rule, for a project Flow's own access
-- check has already granted. Personal projects still never come back.
-- name: GetDashboardProjectByID :one
SELECT
    fp.id,
    fp.name,
    COALESCE(fp.status, 'active') AS status,
    fp.visibility
FROM flow_projects fp
WHERE fp.id = sqlc.arg('project_id')
  AND fp.business_id = sqlc.arg('business_id')
  AND fp.is_personal = FALSE;

-- The projects among project_ids, for the ones Flow's sidebar lists that the
-- viewer rule above does not cover. Same exclusions as
-- ListDashboardProjectsForUser.
-- name: ListDashboardProjectsByIDs :many
SELECT
    fp.id,
    fp.name,
    COALESCE(fp.status, 'active') AS status,
    fp.visibility
FROM flow_projects fp
WHERE fp.business_id = sqlc.arg('business_id')
  AND fp.is_personal = FALSE
  AND fp.archived_at IS NULL
  AND COALESCE(fp.status, 'active') <> 'archived'
  AND fp.id IN (sqlc.slice('project_ids'))
ORDER BY fp.name ASC, fp.id ASC;

-- Progress is Flow's project progress (projects_v2.sql in workspaces): top-level
-- tasks, cancelled ones out of the total, done decided by the status bucket
-- (the custom status in the project, else the legacy status, whose 'done' and
-- 'cancelled' are their own bucket). Overdue counts every open task past its
-- due date, subtasks included, like the business-wide overdue count. Archived
-- tasks are hidden in Flow, so neither counts them.
-- name: GetProjectTaskProgress :one
SELECT
    CAST(COUNT(CASE WHEN ft.parent_task_id IS NULL AND COALESCE(fs.bucket, ft.status, 'todo') <> 'cancelled' THEN 1 END) AS SIGNED) AS total_tasks,
    CAST(COUNT(CASE WHEN ft.parent_task_id IS NULL AND COALESCE(fs.bucket, ft.status) = 'done' THEN 1 END) AS SIGNED) AS completed_tasks,
    CAST(COUNT(CASE WHEN ft.status NOT IN ('done','cancelled') AND ft.due_date < CURDATE() THEN 1 END) AS SIGNED) AS overdue_tasks
FROM flow_tasks ft
LEFT JOIN flow_statuses fs ON fs.id = ft.status_id AND fs.project_id = ft.project_id
WHERE ft.business_id = sqlc.arg('business_id')
  AND ft.project_id = sqlc.arg('project_id')
  AND COALESCE(ft.archived, 0) = 0;

-- The five open tasks longest past their due date.
-- name: ListProjectOverdueTasks :many
SELECT
    ft.id,
    ft.title,
    DATE_FORMAT(ft.due_date, '%Y-%m-%d') AS due_date,
    ft.assignee_id
FROM flow_tasks ft
WHERE ft.business_id = sqlc.arg('business_id')
  AND ft.project_id = sqlc.arg('project_id')
  AND COALESCE(ft.archived, 0) = 0
  AND ft.status NOT IN ('done','cancelled')
  AND ft.due_date < CURDATE()
ORDER BY ft.due_date ASC, ft.id ASC
LIMIT 5;

-- Flow's own time logs on one project, under the same rules as
-- GetLoggedMinutesByProject: rows linked to an HR time entry are HR's time.
-- name: GetProjectLoggedMinutes :one
SELECT
    CAST(COALESCE(SUM(tl.duration_minutes), 0) AS SIGNED) AS all_time_minutes,
    CAST(COALESCE(SUM(CASE WHEN tl.logged_at >= sqlc.arg('this_from') AND tl.logged_at <= sqlc.arg('this_to') THEN tl.duration_minutes END), 0) AS SIGNED) AS this_minutes,
    CAST(COALESCE(SUM(CASE WHEN tl.logged_at >= sqlc.arg('prior_from') AND tl.logged_at <= sqlc.arg('prior_to') THEN tl.duration_minutes END), 0) AS SIGNED) AS prior_minutes
FROM flow_time_logs tl
JOIN flow_tasks ft ON ft.id = tl.task_id
WHERE ft.business_id = sqlc.arg('business_id')
  AND ft.project_id = sqlc.arg('project_id')
  AND tl.time_entry_id IS NULL;
