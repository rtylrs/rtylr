-- name: ListAutomations :many
SELECT id, owner_id, name, trigger_type, trigger_config, actions, is_active,
       last_run_at, run_count, created_at, updated_at
FROM crm_automations WHERE owner_id = ? ORDER BY created_at DESC;

-- name: CreateAutomation :execresult
INSERT INTO crm_automations
  (owner_id, name, trigger_type, trigger_config, actions, is_active)
VALUES (?, ?, ?, ?, ?, ?);

-- name: UpdateAutomation :execresult
UPDATE crm_automations
SET name = ?, trigger_type = ?, trigger_config = ?, actions = ?, is_active = ?
WHERE id = ? AND owner_id = ?;

-- name: DeleteAutomation :execresult
DELETE FROM crm_automations WHERE id = ? AND owner_id = ?;
