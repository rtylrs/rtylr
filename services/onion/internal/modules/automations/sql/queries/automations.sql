-- name: ListAutomations :many
SELECT id, owner_id, business_id, name, trigger_type, trigger_config, actions, is_active,
       last_run_at, run_count, created_at, updated_at
FROM crm_automations WHERE business_id = ? ORDER BY created_at DESC LIMIT 100;

-- name: CreateAutomation :execresult
INSERT INTO crm_automations
  (owner_id, business_id, name, trigger_type, trigger_config, actions, is_active)
VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: UpdateAutomation :execresult
UPDATE crm_automations
SET name = ?, trigger_type = ?, trigger_config = ?, actions = ?, is_active = ?
WHERE id = ? AND business_id = ?;

-- name: DeleteAutomation :execresult
DELETE FROM crm_automations WHERE id = ? AND business_id = ?;
