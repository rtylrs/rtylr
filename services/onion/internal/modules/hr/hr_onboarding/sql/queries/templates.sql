-- name: ListChecklistTemplates :many
SELECT id, owner_id, business_id, name, trigger_type, role_id, created_at, updated_at
FROM hr_onboarding_checklist_templates
WHERE business_id = ?
  AND (sqlc.narg('trigger_type') IS NULL OR trigger_type = sqlc.narg('trigger_type'))
ORDER BY name ASC
LIMIT ? OFFSET ?;

-- name: CountChecklistTemplates :one
SELECT COUNT(*) AS total
FROM hr_onboarding_checklist_templates
WHERE business_id = ?
  AND (sqlc.narg('trigger_type') IS NULL OR trigger_type = sqlc.narg('trigger_type'));

-- name: GetChecklistTemplate :one
SELECT id, owner_id, business_id, name, trigger_type, role_id, created_at, updated_at
FROM hr_onboarding_checklist_templates
WHERE business_id = ? AND id = ?;

-- name: CreateChecklistTemplate :execresult
INSERT INTO hr_onboarding_checklist_templates (owner_id, business_id, name, trigger_type, role_id)
VALUES (?, ?, ?, ?, ?);

-- name: UpdateChecklistTemplate :execresult
UPDATE hr_onboarding_checklist_templates
SET name = ?,
    trigger_type = COALESCE(sqlc.narg('trigger_type'), trigger_type),
    role_id = COALESCE(sqlc.narg('role_id'), role_id)
WHERE business_id = ? AND id = ?;

-- name: CreateChecklistItem :execresult
INSERT INTO hr_onboarding_checklist_items (owner_id, business_id, template_id, title, description, assignee_role, due_offset_days, requires_attachment)
VALUES (?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetChecklistItem :one
SELECT id, owner_id, business_id, template_id, title, description, assignee_role, due_offset_days, requires_attachment, created_at, updated_at
FROM hr_onboarding_checklist_items
WHERE business_id = ? AND id = ?;
