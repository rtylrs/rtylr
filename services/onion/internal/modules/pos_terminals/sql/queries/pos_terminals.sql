-- name: ListTerminalsByBusiness :many
SELECT id, business_id, branch_id, name, device_id, is_active, last_seen_at, created_at
FROM pos_terminals WHERE business_id = ? ORDER BY name;

-- name: CreateTerminal :execresult
INSERT INTO pos_terminals (business_id, branch_id, name, device_id, is_active)
VALUES (?, ?, ?, ?, TRUE);

-- name: UpdateTerminal :exec
UPDATE pos_terminals SET branch_id = ?, name = ?, device_id = ?, is_active = ?
WHERE id = ? AND business_id = ?;

-- name: UpdateTerminalHeartbeat :execresult
UPDATE pos_terminals SET last_seen_at = NOW()
WHERE id = ? AND business_id = ?;
