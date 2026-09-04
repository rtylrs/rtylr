-- name: CountWorkspacesByOwner :one
SELECT COUNT(*) AS cnt FROM flow_workspaces WHERE business_id = ?;

-- name: CreateWorkspace :execresult
INSERT INTO flow_workspaces (owner_id, business_id, name, description, icon, color, is_default, created_by)
VALUES (?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetWorkspacesByOwner :many
SELECT id, owner_id, business_id, name, description, icon, color, is_default, created_by, created_at, updated_at
FROM flow_workspaces
WHERE business_id = ?
ORDER BY is_default DESC, created_at ASC, id ASC
LIMIT ? OFFSET ?;

-- name: UpdateWorkspace :exec
UPDATE flow_workspaces SET name = ?, description = ?, icon = ?, color = ?
WHERE id = ? AND business_id = ?;

-- name: DeleteWorkspace :exec
DELETE FROM flow_workspaces WHERE id = ? AND business_id = ?;

-- name: AddWorkspaceMember :exec
INSERT INTO flow_workspace_members (workspace_id, user_id, role) VALUES (?, ?, ?)
ON DUPLICATE KEY UPDATE role = VALUES(role);

-- name: RemoveWorkspaceMember :exec
DELETE FROM flow_workspace_members WHERE workspace_id = ? AND user_id = ?;

-- name: CreateProject :execresult
INSERT INTO flow_projects (workspace_id, owner_id, business_id, name, description, icon, color, visibility, is_personal, start_date, due_date, created_by)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: CountProjectsByWorkspace :one
SELECT COUNT(*) FROM flow_projects WHERE workspace_id = ? AND business_id = ?;

-- name: GetProjectsByWorkspace :many
WITH project_page AS (
    SELECT p.id, p.workspace_id, p.owner_id, p.business_id, p.name, p.description, p.icon, p.color, p.status,
           p.start_date, p.due_date, p.created_by, p.created_at, p.updated_at
    FROM flow_projects p
    WHERE p.workspace_id = ? AND p.business_id = ?
    ORDER BY p.created_at DESC, p.id DESC
    LIMIT ? OFFSET ?
),
task_counts AS (
    SELECT t.project_id,
           COUNT(*) AS task_count,
           SUM(t.status = 'done') AS completed_count
    FROM flow_tasks t
    JOIN project_page p ON p.id = t.project_id
    GROUP BY t.project_id
),
member_counts AS (
    SELECT pm.project_id, COUNT(*) AS member_count
    FROM flow_project_members pm
    JOIN project_page p ON p.id = pm.project_id
    GROUP BY pm.project_id
)
SELECT
    p.id, p.workspace_id, p.owner_id, p.business_id, p.name,
    COALESCE(p.description, '') AS description,
    COALESCE(p.icon, '') AS icon,
    COALESCE(p.color, '') AS color,
    COALESCE(p.status, 'active') AS status,
    p.start_date, p.due_date,
    p.created_by, p.created_at, p.updated_at,
    COALESCE(tc.task_count, 0) AS task_count,
    COALESCE(tc.completed_count, 0) AS completed_count,
    COALESCE(mc.member_count, 0) AS member_count
FROM project_page p
LEFT JOIN task_counts tc ON tc.project_id = p.id
LEFT JOIN member_counts mc ON mc.project_id = p.id
ORDER BY p.created_at DESC, p.id DESC;

-- name: UpdateProject :exec
UPDATE flow_projects SET name = ?, description = ?, icon = ?, color = ?, status = ?, start_date = ?, due_date = ?
WHERE id = ? AND business_id = ?;

-- name: ArchiveProject :exec
-- Stamps the recycle-bin timestamps: purge_at arrives from the service, computed
-- from the business's retention setting at archive time.
UPDATE flow_projects SET status = 'archived', archived_at = NOW(), purge_at = ?
WHERE id = ? AND business_id = ?;

-- name: CreateTask :execresult
INSERT INTO flow_tasks
(project_id, workspace_id, owner_id, business_id, parent_task_id, title, description, priority,
 assignee_id, reporter_id, due_date, start_date, estimated_hours, labels)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetTaskByID :one
SELECT id, project_id, workspace_id, owner_id, parent_task_id, title, description,
       status, priority, assignee_id, reporter_id, due_date, start_date, completed_at,
       estimated_hours, actual_hours, sort_order, labels, created_at, updated_at
FROM flow_tasks WHERE id = ?;

-- name: UpdateTask :exec
UPDATE flow_tasks
SET title = ?, description = ?, status = ?, priority = ?, assignee_id = ?,
    due_date = ?, start_date = ?, labels = ?,
    completed_at = CASE WHEN ? = 'done' AND completed_at IS NULL THEN NOW() WHEN ? != 'done' THEN NULL ELSE completed_at END
WHERE id = ? AND business_id = ?;

-- name: UpdateTaskStatus :exec
UPDATE flow_tasks
SET status = ?,
    completed_at = CASE WHEN ? = 'done' AND completed_at IS NULL THEN NOW() WHEN ? != 'done' THEN NULL ELSE completed_at END
WHERE id = ? AND business_id = ?;

-- name: DeleteTask :exec
DELETE FROM flow_tasks WHERE id = ? AND business_id = ?;

-- name: CreateComment :execresult
INSERT INTO flow_task_comments (task_id, user_id, content) VALUES (?, ?, ?);

-- name: GetCommentByID :one
-- cross-schema read from users: resolve the comment author's display identity
SELECT c.id, c.task_id, c.user_id, COALESCE(u.full_name, u.email, '') AS actor_name,
       c.content, c.is_edited, c.created_at, c.updated_at
FROM flow_task_comments c
LEFT JOIN users u ON u.id = c.user_id
WHERE c.id = ?
LIMIT 1;

-- cross-schema read from users: resolve the comment author's display identity
-- name: GetCommentsByTask :many
SELECT c.id, c.task_id, c.user_id, COALESCE(u.full_name, u.email, '') AS actor_name,
       c.content, c.is_edited, c.created_at, c.updated_at
FROM flow_task_comments c
LEFT JOIN users u ON u.id = c.user_id
WHERE c.task_id = ? ORDER BY c.created_at ASC, c.id ASC LIMIT 500;

-- name: DeleteComment :exec
DELETE FROM flow_task_comments WHERE id = ? AND user_id = ?;

-- name: LogActivity :exec
INSERT INTO flow_task_activity (task_id, user_id, action, old_value, new_value) VALUES (?, ?, ?, ?, ?);

-- name: GetActivityByTask :many
-- cross-schema read from users: resolve the activity author's display identity
SELECT a.id, a.task_id, a.user_id, COALESCE(u.full_name, u.email, '') AS actor_name,
       a.action, a.old_value, a.new_value, a.created_at
FROM flow_task_activity a
LEFT JOIN users u ON u.id = a.user_id
WHERE a.task_id = ? ORDER BY a.created_at DESC LIMIT ?;

-- name: CreateAttachment :execresult
INSERT INTO flow_task_attachments (task_id, owner_id, business_id, name, url, size, uploaded_by)
VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: GetAttachmentsByTask :many
SELECT id, task_id, owner_id, business_id, name, url, size, uploaded_by, created_at
FROM flow_task_attachments WHERE task_id = ? ORDER BY created_at DESC, id DESC LIMIT 500;

-- name: DeleteAttachment :exec
DELETE FROM flow_task_attachments WHERE id = ? AND business_id = ?;

-- name: CreateTimeLog :execresult
INSERT INTO flow_time_logs (owner_id, business_id, task_id, user_id, duration_minutes, description)
VALUES (?, ?, ?, ?, ?, ?);

-- name: GetTimeLogsByTask :many
SELECT id, owner_id, business_id, task_id, user_id, duration_minutes, description, logged_at, created_at
FROM flow_time_logs WHERE task_id = ? ORDER BY logged_at DESC, id DESC LIMIT 500;

-- name: CreateLabel :execresult
INSERT INTO flow_labels (workspace_id, name, color) VALUES (?, ?, ?);

-- name: CountLabelsByWorkspace :one
SELECT COUNT(*) FROM flow_labels WHERE workspace_id = ?;

-- name: GetLabelsByWorkspace :many
SELECT id, workspace_id, name, color, created_at
FROM flow_labels WHERE workspace_id = ? ORDER BY name ASC
LIMIT ? OFFSET ?;

-- name: GetDashboardStats :one
SELECT
    COUNT(CASE WHEN status = 'todo' THEN 1 END) AS open_tasks,
    COUNT(CASE WHEN status = 'in_progress' THEN 1 END) AS in_progress_tasks,
    COUNT(CASE WHEN status = 'done' AND completed_at >= ? AND completed_at < ? THEN 1 END) AS completed_this_week,
    COUNT(CASE WHEN status NOT IN ('done','cancelled') AND due_date < NOW() THEN 1 END) AS overdue_tasks
FROM flow_tasks WHERE workspace_id = ? AND business_id = ?;

-- name: CountProjectColumns :one
SELECT COUNT(*) FROM flow_project_columns WHERE project_id = ?;
