-- name: ListIntegrations :many
SELECT id, owner_id, provider, status, config, connected_at, created_at
FROM hub_integrations
WHERE owner_id = ?
ORDER BY created_at DESC;

-- name: ConnectIntegration :execresult
INSERT INTO hub_integrations (owner_id, provider, status, config, connected_at)
VALUES (?, ?, 'connected', ?, NOW());

-- name: DisconnectIntegration :exec
DELETE FROM hub_integrations WHERE id = ? AND owner_id = ?;
