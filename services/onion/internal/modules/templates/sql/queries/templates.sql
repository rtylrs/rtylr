-- name: ListTemplates :many
SELECT id, owner_id, business_id, name, type, subject, body, variables, is_active, created_at, updated_at
FROM crm_templates WHERE business_id = ? ORDER BY created_at DESC, id DESC;

-- name: CreateTemplate :execresult
INSERT INTO crm_templates
  (owner_id, business_id, name, type, subject, body, variables, is_active)
VALUES (?, ?, ?, ?, ?, ?, ?, ?);

-- name: UpdateTemplate :execresult
UPDATE crm_templates
SET name = ?, type = ?, subject = ?, body = ?, variables = ?, is_active = ?
WHERE id = ? AND business_id = ?;

-- name: DeleteTemplate :execresult
DELETE FROM crm_templates WHERE id = ? AND business_id = ?;
