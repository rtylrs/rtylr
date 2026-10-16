-- name: ListIntegrations :many
SELECT id, owner_id, business_id, provider, status, config, connected_at, created_at
FROM hub_integrations
WHERE business_id = ?
ORDER BY created_at DESC
LIMIT 200;

-- name: ConnectIntegration :execresult
INSERT INTO hub_integrations (owner_id, business_id, provider, status, config, connected_at)
VALUES (?, ?, ?, 'connected', ?, NOW());

-- name: FindIntegrationByProvider :one
SELECT id FROM hub_integrations WHERE business_id = ? AND provider = ? ORDER BY id LIMIT 1;

-- name: ReconnectIntegration :exec
UPDATE hub_integrations SET status = 'connected', config = ?, connected_at = NOW()
WHERE id = ? AND business_id = ?;

-- name: DisconnectIntegration :execresult
DELETE FROM hub_integrations WHERE id = ? AND business_id = ?;
