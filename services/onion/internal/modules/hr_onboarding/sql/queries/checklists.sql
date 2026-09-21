-- name: CreateEmployeeChecklist :execresult
INSERT INTO hr_onboarding_employee_checklists (owner_id, business_id, employee_id, template_id, status, triggered_at)
VALUES (?, ?, ?, ?, ?, ?);

-- name: GetEmployeeChecklist :one
SELECT id, owner_id, business_id, employee_id, template_id, status, triggered_at, created_at, updated_at
FROM hr_onboarding_employee_checklists
WHERE business_id = ? AND employee_id = ? AND id = ?;

-- name: GetLatestEmployeeChecklist :one
SELECT id, owner_id, business_id, employee_id, template_id, status, triggered_at, created_at, updated_at
FROM hr_onboarding_employee_checklists
WHERE business_id = ? AND employee_id = ?
ORDER BY triggered_at DESC, id DESC
LIMIT 1;

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
