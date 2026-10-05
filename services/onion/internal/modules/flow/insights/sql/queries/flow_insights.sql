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

-- name: GetRoadmapProjects :many
SELECT
    fp.id,
    fp.name,
    fp.status,
    COALESCE(DATE_FORMAT(fp.start_date, '%Y-%m-%d'), '') AS start_date,
    COALESCE(DATE_FORMAT(fp.due_date, '%Y-%m-%d'), '') AS due_date,
    COUNT(ft.id) AS total_tasks,
    COUNT(CASE WHEN ft.status = 'done' THEN 1 END) AS done_tasks
FROM flow_projects fp
LEFT JOIN flow_tasks ft ON ft.project_id = fp.id
 AND ft.parent_task_id IS NULL
 AND ft.created_at >= DATE_SUB(UTC_TIMESTAMP(), INTERVAL 12 MONTH)
WHERE fp.business_id = ? AND fp.status != 'archived'
GROUP BY fp.id, fp.name, fp.status, fp.start_date, fp.due_date
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
