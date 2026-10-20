-- assigned_to = 0 selects unassigned tasks.

-- name: CreateTask :execresult
INSERT INTO crm_tasks (business_id, owner_id, lead_id, deal_id, title, type, due_date, assigned_to, created_by)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetTask :one
SELECT * FROM crm_tasks WHERE id = ? AND business_id = ?;

-- name: UpdateTask :exec
UPDATE crm_tasks SET title = ?, type = ?, due_date = ?, assigned_to = ? WHERE id = ? AND business_id = ?;

-- name: SetTaskCompletedAt :exec
UPDATE crm_tasks SET completed_at = ? WHERE id = ? AND business_id = ?;

-- name: DeleteTask :execresult
DELETE FROM crm_tasks WHERE id = ? AND business_id = ?;

-- name: ListOpenTasks :many
-- Earliest due first. The joined names let a worklist say who to call.
SELECT t.id, t.lead_id, t.deal_id, t.title, t.type, t.due_date, t.assigned_to, t.completed_at,
       t.created_by, t.created_at,
       COALESCE(l.first_name, '') AS lead_first_name, COALESCE(l.last_name, '') AS lead_last_name,
       COALESCE(l.company_name, '') AS lead_company_name, COALESCE(l.phone, '') AS lead_phone,
       COALESCE(d.title, '') AS deal_title
FROM crm_tasks t
LEFT JOIN crm_leads l ON l.id = t.lead_id AND l.business_id = t.business_id
LEFT JOIN crm_deals d ON d.id = t.deal_id AND d.business_id = t.business_id
WHERE t.business_id = sqlc.arg('business_id') AND t.completed_at IS NULL
  AND (sqlc.narg('assigned_to') IS NULL OR COALESCE(t.assigned_to, 0) = sqlc.narg('assigned_to'))
  AND (sqlc.narg('due_from') IS NULL OR t.due_date >= sqlc.narg('due_from'))
  AND (sqlc.narg('due_to') IS NULL OR t.due_date <= sqlc.narg('due_to'))
  AND (sqlc.narg('lead_id') IS NULL OR t.lead_id = sqlc.narg('lead_id'))
  AND (sqlc.narg('deal_id') IS NULL OR t.deal_id = sqlc.narg('deal_id'))
  AND (sqlc.narg('cursor_id') IS NULL OR t.due_date > sqlc.narg('cursor_date')
       OR (t.due_date = sqlc.narg('cursor_date') AND t.id > sqlc.narg('cursor_id')))
ORDER BY t.due_date, t.id
LIMIT ?;

-- name: ListCompletedTasks :many
SELECT t.id, t.lead_id, t.deal_id, t.title, t.type, t.due_date, t.assigned_to, t.completed_at,
       t.created_by, t.created_at,
       COALESCE(l.first_name, '') AS lead_first_name, COALESCE(l.last_name, '') AS lead_last_name,
       COALESCE(l.company_name, '') AS lead_company_name, COALESCE(l.phone, '') AS lead_phone,
       COALESCE(d.title, '') AS deal_title
FROM crm_tasks t
LEFT JOIN crm_leads l ON l.id = t.lead_id AND l.business_id = t.business_id
LEFT JOIN crm_deals d ON d.id = t.deal_id AND d.business_id = t.business_id
WHERE t.business_id = sqlc.arg('business_id') AND t.completed_at IS NOT NULL
  AND (sqlc.narg('assigned_to') IS NULL OR COALESCE(t.assigned_to, 0) = sqlc.narg('assigned_to'))
  AND (sqlc.narg('lead_id') IS NULL OR t.lead_id = sqlc.narg('lead_id'))
  AND (sqlc.narg('deal_id') IS NULL OR t.deal_id = sqlc.narg('deal_id'))
  AND (sqlc.narg('cursor_id') IS NULL OR t.completed_at < sqlc.narg('cursor_time')
       OR (t.completed_at = sqlc.narg('cursor_time') AND t.id < sqlc.narg('cursor_id')))
ORDER BY t.completed_at DESC, t.id DESC
LIMIT ?;

-- name: CountOpenTasks :one
SELECT CAST(COALESCE(SUM(due_date < sqlc.arg('today')), 0) AS SIGNED) AS overdue,
       CAST(COALESCE(SUM(due_date = sqlc.arg('today')), 0) AS SIGNED) AS due_today,
       CAST(COALESCE(SUM(due_date > sqlc.arg('today')), 0) AS SIGNED) AS upcoming
FROM crm_tasks
WHERE business_id = sqlc.arg('business_id') AND completed_at IS NULL
  AND (sqlc.narg('assigned_to') IS NULL OR COALESCE(assigned_to, 0) = sqlc.narg('assigned_to'));

-- name: AttachLeadTasksToDeal :exec
UPDATE crm_tasks SET deal_id = ? WHERE business_id = ? AND lead_id = ? AND deal_id IS NULL;

-- name: DeleteLeadTasks :exec
DELETE FROM crm_tasks WHERE business_id = sqlc.arg('business_id') AND lead_id IN (sqlc.slice('lead_ids'));

-- name: DeleteDealTasks :exec
DELETE FROM crm_tasks WHERE business_id = ? AND deal_id = ?;

-- name: ListDeletableLeadIDs :many
SELECT id FROM crm_leads
WHERE business_id = sqlc.arg('business_id') AND id IN (sqlc.slice('ids')) AND status <> 'qualified';
