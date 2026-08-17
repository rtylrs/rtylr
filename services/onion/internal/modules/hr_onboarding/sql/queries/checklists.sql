-- name: CreateEmployeeChecklist :execresult
INSERT INTO hr_onboarding_employee_checklists (owner_id, business_id, employee_id, template_id, status, triggered_at)
VALUES (?, ?, ?, ?, ?, ?);

-- name: GetEmployeeChecklist :one
SELECT id, owner_id, business_id, employee_id, template_id, status, triggered_at, created_at, updated_at
FROM hr_onboarding_employee_checklists
WHERE business_id = ? AND employee_id = ? AND id = ?;

-- name: ListEmployeeChecklists :many
SELECT id, owner_id, business_id, employee_id, template_id, status, triggered_at, created_at, updated_at
FROM hr_onboarding_employee_checklists
WHERE business_id = ? AND employee_id = ?
ORDER BY triggered_at DESC
LIMIT ? OFFSET ?;

-- name: CountEmployeeChecklists :one
SELECT COUNT(*) AS total
FROM hr_onboarding_employee_checklists
WHERE business_id = ? AND employee_id = ?;

-- name: UpdateEmployeeChecklistStatus :exec
UPDATE hr_onboarding_employee_checklists
SET status = ?
WHERE business_id = ? AND id = ?;

-- name: CreateEmployeeChecklistItem :exec
INSERT INTO hr_onboarding_employee_checklist_items (owner_id, business_id, employee_checklist_id, checklist_item_id, assigned_to_employee_id, due_date, status)
VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: GetEmployeeChecklistItem :one
SELECT id, owner_id, business_id, employee_checklist_id, checklist_item_id, assigned_to_employee_id, due_date, status, completed_at, completed_by, attachment_url, created_at, updated_at
FROM hr_onboarding_employee_checklist_items
WHERE business_id = ? AND id = ?;

-- name: CompleteEmployeeChecklistItem :exec
UPDATE hr_onboarding_employee_checklist_items
SET status = 'complete', completed_at = NOW(), completed_by = ?
WHERE business_id = ? AND id = ?;

-- name: CountChecklistItemsTotal :one
SELECT COUNT(*) AS total
FROM hr_onboarding_employee_checklist_items
WHERE business_id = ? AND employee_checklist_id = ?;

-- name: CountChecklistItemsCompleted :one
SELECT COUNT(*) AS total
FROM hr_onboarding_employee_checklist_items
WHERE business_id = ? AND employee_checklist_id = ? AND status = 'complete';

-- name: ListEmployeeChecklistItems :many
SELECT id, owner_id, business_id, employee_checklist_id, checklist_item_id, assigned_to_employee_id, due_date, status, completed_at, completed_by, attachment_url, created_at, updated_at
FROM hr_onboarding_employee_checklist_items
WHERE business_id = ? AND employee_checklist_id = ?
ORDER BY due_date ASC
LIMIT 500;

-- name: ListBusinessEmployeeChecklistSummaries :many
-- Business-wide progress summary, one row per employee checklist. Overdue is
-- computed against CURDATE() in the database so no caller has to re-derive it
-- from raw due dates.
SELECT
    c.id AS checklist_id,
    c.employee_id,
    c.template_id,
    c.status,
    c.triggered_at,
    c.updated_at,
    COUNT(i.id) AS total_items,
    COUNT(CASE WHEN i.status = 'complete' THEN 1 END) AS completed_items,
    COUNT(CASE WHEN i.status <> 'complete' AND i.due_date < CURDATE() THEN 1 END) AS overdue_items
FROM hr_onboarding_employee_checklists c
LEFT JOIN hr_onboarding_employee_checklist_items i
    ON i.employee_checklist_id = c.id AND i.business_id = c.business_id
WHERE c.business_id = ?
GROUP BY c.id, c.employee_id, c.template_id, c.status, c.triggered_at, c.updated_at
ORDER BY c.employee_id ASC, c.triggered_at DESC
LIMIT ? OFFSET ?;

-- name: CountBusinessEmployeeChecklists :one
SELECT COUNT(*) AS total
FROM hr_onboarding_employee_checklists
WHERE business_id = ?;
