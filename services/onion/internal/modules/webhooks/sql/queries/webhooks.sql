-- name: ListWebhooks :many
SELECT id, owner_id, url, events, is_active, last_triggered_at, created_at, updated_at
FROM hub_webhooks
WHERE owner_id = ?
ORDER BY created_at DESC;

-- name: CreateWebhook :execresult
INSERT INTO hub_webhooks (owner_id, url, events, secret)
VALUES (?, ?, ?, ?);

-- name: DeleteWebhook :exec
DELETE FROM hub_webhooks WHERE id = ? AND owner_id = ?;
