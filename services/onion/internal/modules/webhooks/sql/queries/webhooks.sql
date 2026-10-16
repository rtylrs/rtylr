-- name: ListWebhooks :many
SELECT id, owner_id, business_id, url, events, secret, is_active, last_triggered_at, created_at, updated_at
FROM hub_webhooks
WHERE business_id = ?
ORDER BY created_at DESC;

-- name: CreateWebhook :execresult
INSERT INTO hub_webhooks (owner_id, business_id, url, events, secret)
VALUES (?, ?, ?, ?, ?);

-- name: DeleteWebhook :exec
DELETE FROM hub_webhooks WHERE id = ? AND business_id = ?;

-- name: GetWebhook :one
SELECT id, owner_id, business_id, url, events, secret, is_active, last_triggered_at, created_at, updated_at
FROM hub_webhooks
WHERE id = ? AND business_id = ?;

-- name: UpdateWebhook :exec
UPDATE hub_webhooks
SET url = ?, events = ?, is_active = ?
WHERE id = ? AND business_id = ?;

-- name: RotateWebhookSecret :execresult
UPDATE hub_webhooks SET secret = ? WHERE id = ? AND business_id = ?;
