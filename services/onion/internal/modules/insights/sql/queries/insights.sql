-- name: GetInsightsByOwner :one
SELECT
    COUNT(*) AS total_tasks,
    COUNT(CASE WHEN status = 'done' THEN 1 END) AS completed_tasks,
    COUNT(CASE WHEN status = 'in_progress' THEN 1 END) AS in_progress_tasks,
    COUNT(CASE WHEN status NOT IN ('done','cancelled') AND due_date < NOW() THEN 1 END) AS overdue_tasks,
    COALESCE(AVG(CASE WHEN status = 'done' AND completed_at IS NOT NULL THEN TIMESTAMPDIFF(HOUR, created_at, completed_at) / 24.0 END), 0) AS avg_cycle_time_days
FROM flow_tasks WHERE owner_id = ?;

-- name: GetInsightsByWorkspace :one
SELECT
    COUNT(*) AS total_tasks,
    COUNT(CASE WHEN status = 'done' THEN 1 END) AS completed_tasks,
    COUNT(CASE WHEN status = 'in_progress' THEN 1 END) AS in_progress_tasks,
    COUNT(CASE WHEN status NOT IN ('done','cancelled') AND due_date < NOW() THEN 1 END) AS overdue_tasks,
    COALESCE(AVG(CASE WHEN status = 'done' AND completed_at IS NOT NULL THEN TIMESTAMPDIFF(HOUR, created_at, completed_at) / 24.0 END), 0) AS avg_cycle_time_days
FROM flow_tasks WHERE owner_id = ? AND workspace_id = ?;

-- name: GetThroughputByOwner :one
SELECT
    COUNT(CASE WHEN completed_at >= ? THEN 1 END) AS tasks_this_week,
    COUNT(CASE WHEN completed_at >= ? AND completed_at < ? THEN 1 END) AS tasks_last_week
FROM flow_tasks WHERE owner_id = ? AND status = 'done';

-- name: GetThroughputByWorkspace :one
SELECT
    COUNT(CASE WHEN completed_at >= ? THEN 1 END) AS tasks_this_week,
    COUNT(CASE WHEN completed_at >= ? AND completed_at < ? THEN 1 END) AS tasks_last_week
FROM flow_tasks WHERE owner_id = ? AND status = 'done' AND workspace_id = ?;

-- name: GetTeamWorkloadByOwner :many
SELECT
    assignee_id,
    COUNT(CASE WHEN status = 'todo' THEN 1 END) AS open_tasks,
    COUNT(CASE WHEN status = 'in_progress' THEN 1 END) AS in_progress_tasks,
    COUNT(CASE WHEN status = 'done' THEN 1 END) AS completed_tasks,
    COUNT(CASE WHEN status NOT IN ('done','cancelled') AND due_date < NOW() THEN 1 END) AS overdue_tasks
FROM flow_tasks
WHERE owner_id = ? AND assignee_id IS NOT NULL
GROUP BY assignee_id;

-- name: GetTeamWorkloadByWorkspace :many
SELECT
    assignee_id,
    COUNT(CASE WHEN status = 'todo' THEN 1 END) AS open_tasks,
    COUNT(CASE WHEN status = 'in_progress' THEN 1 END) AS in_progress_tasks,
    COUNT(CASE WHEN status = 'done' THEN 1 END) AS completed_tasks,
    COUNT(CASE WHEN status NOT IN ('done','cancelled') AND due_date < NOW() THEN 1 END) AS overdue_tasks
FROM flow_tasks
WHERE owner_id = ? AND assignee_id IS NOT NULL AND workspace_id = ?
GROUP BY assignee_id;

-- name: GetRoadmapByOwner :many
SELECT
    p.id, p.name, p.status, p.start_date, p.due_date,
    COUNT(t.id) AS total_tasks,
    COUNT(CASE WHEN t.status = 'done' THEN 1 END) AS completed_tasks
FROM flow_projects p
LEFT JOIN flow_tasks t ON t.project_id = p.id
WHERE p.owner_id = ? AND p.status != 'archived'
GROUP BY p.id, p.name, p.status, p.start_date, p.due_date
ORDER BY p.start_date ASC, p.created_at ASC;

-- name: GetRoadmapByWorkspace :many
SELECT
    p.id, p.name, p.status, p.start_date, p.due_date,
    COUNT(t.id) AS total_tasks,
    COUNT(CASE WHEN t.status = 'done' THEN 1 END) AS completed_tasks
FROM flow_projects p
LEFT JOIN flow_tasks t ON t.project_id = p.id
WHERE p.owner_id = ? AND p.status != 'archived' AND p.workspace_id = ?
GROUP BY p.id, p.name, p.status, p.start_date, p.due_date
ORDER BY p.start_date ASC, p.created_at ASC;
