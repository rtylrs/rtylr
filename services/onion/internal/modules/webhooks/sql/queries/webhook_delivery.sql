-- Delivery side of the event fan-out. The dispatcher
-- (internal/platform/eventbus) writes one webhook_deliveries row per
-- registration that subscribes to an event; these statements drain that queue.

-- cross-schema read from shared — webhook_deliveries is the platform fan-out queue.
-- Ordered by due time so a delivery backed off to later does not overtake one
-- ready now. 'failed' means retryable; 'dead' has spent its attempts and is
-- deliberately never picked up again.
-- name: ListDueWebhookDeliveries :many
SELECT id, business_id, webhook_id, event_id, attempts, max_attempts
FROM webhook_deliveries
WHERE status IN ('pending', 'failed')
  AND next_attempt_at <= NOW()
ORDER BY next_attempt_at ASC, id ASC
LIMIT ?;

-- Optimistic claim. The status predicate is what makes several workers reading
-- the same due list safe: exactly one UPDATE affects the row.
--
-- attempts increments here rather than on failure, so a delivery that crashes
-- the worker still counts against max_attempts instead of retrying forever.
-- name: ClaimWebhookDelivery :execresult
UPDATE webhook_deliveries
SET status = 'delivering', attempts = attempts + 1, updated_at = NOW()
WHERE id = ? AND status IN ('pending', 'failed');

-- name: MarkWebhookDelivered :exec
UPDATE webhook_deliveries
SET status = 'delivered',
    delivered_at = NOW(),
    response_status = sqlc.arg('response_status'),
    last_error = NULL,
    updated_at = NOW()
WHERE id = sqlc.arg('id');

-- 'dead' once attempts reach the cap, so an endpoint that never answers stops
-- consuming the queue and stays visible rather than retrying forever.
-- name: MarkWebhookDeliveryFailed :exec
UPDATE webhook_deliveries
SET status = IF(attempts >= max_attempts, 'dead', 'failed'),
    next_attempt_at = DATE_ADD(NOW(), INTERVAL sqlc.arg('retry_after_seconds') SECOND),
    response_status = sqlc.arg('response_status'),
    last_error = sqlc.arg('last_error'),
    updated_at = NOW()
WHERE id = sqlc.arg('id');

-- A destination that can never be delivered to, such as one resolving to a
-- private address, skips the retries entirely.
-- name: MarkWebhookDeliveryDead :exec
UPDATE webhook_deliveries
SET status = 'dead', last_error = ?, updated_at = NOW()
WHERE id = ?;

-- Scoped by business_id as well as id: the delivery carries both, and matching
-- on id alone would let a corrupted row send another tenant's payload to a
-- registration, or sign it with the wrong secret.
-- name: GetWebhookForDelivery :one
SELECT id, url, secret, is_active
FROM hub_webhooks
WHERE id = ? AND business_id = ?;

-- cross-schema read from shared — the event being delivered.
-- name: GetEventForDelivery :one
SELECT id, event_type, subject_type, subject_id, payload, occurred_at
FROM platform_events
WHERE id = ? AND business_id = ?;

-- last_triggered_at exists on hub_webhooks and nothing has ever written it.
-- name: RecordWebhookTriggered :exec
UPDATE hub_webhooks
SET last_triggered_at = NOW()
WHERE id = ? AND business_id = ?;
