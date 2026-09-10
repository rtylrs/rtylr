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

-- Requeues a delivery abandoned mid-flight, which happens when the worker dies
-- between claiming a row and recording its outcome. Without this the row sits
-- in 'delivering' forever and the subscriber never hears about that event.
--
-- Requeuing means a delivery already sent may be sent again, so this makes
-- webhooks at-least-once. That is the documented contract for webhooks and the
-- reason every delivery carries X-Rtylr-Delivery-Id and X-Rtylr-Event-Id: a
-- subscriber is expected to dedupe on the event id. The automation queue is
-- deliberately handled the other way round; see the note on
-- RecoverAbandonedAutomationRuns.
--
-- Ages off updated_at, which the claim sets explicitly. The threshold must be
-- comfortably longer than any legitimate send, which is bounded by the
-- per-request timeout.
-- name: RecoverAbandonedWebhookDeliveries :execresult
UPDATE webhook_deliveries
SET status = IF(attempts >= max_attempts, 'dead', 'failed'),
    next_attempt_at = NOW(),
    last_error = 'delivery was abandoned in flight and has been requeued',
    updated_at = NOW()
WHERE status = 'delivering'
  AND updated_at < DATE_SUB(NOW(), INTERVAL sqlc.arg('stale_after_seconds') SECOND);

-- The tenant-facing delivery listing. Scoped by business so a caller only ever
-- sees their own registrations' deliveries, and joined to platform_events for
-- the event type, which is what makes a row meaningful without a second call.
--
-- The two filters are optional and expressed the sqlc way: a zero webhook_id
-- and an empty status mean "no filter" rather than "match nothing".
-- name: ListWebhookDeliveriesForBusiness :many
SELECT d.id, d.webhook_id, d.event_id, e.event_type, d.status,
       d.attempts, d.max_attempts, d.response_status, d.last_error,
       d.next_attempt_at, d.delivered_at, d.created_at
FROM webhook_deliveries d
JOIN platform_events e ON e.id = d.event_id
WHERE d.business_id = sqlc.arg('business_id')
  AND (sqlc.arg('webhook_id') = 0 OR d.webhook_id = sqlc.arg('webhook_id'))
  AND (sqlc.arg('status') = '' OR d.status = sqlc.arg('status'))
ORDER BY d.id DESC
LIMIT ? OFFSET ?;

-- name: CountWebhookDeliveriesForBusiness :one
SELECT COUNT(*)
FROM webhook_deliveries d
WHERE d.business_id = sqlc.arg('business_id')
  AND (sqlc.arg('webhook_id') = 0 OR d.webhook_id = sqlc.arg('webhook_id'))
  AND (sqlc.arg('status') = '' OR d.status = sqlc.arg('status'));

-- Registrations that have dead-lettered something recently, one row each.
--
-- Grouped rather than one row per dead delivery: an endpoint that stops
-- answering produces many, and the owner needs to hear that their endpoint is
-- failing, not receive one message per lost event.
--
-- Bounded by how recently a delivery died, so a registration that failed
-- months ago and was never fixed does not generate a notification forever.
-- updated_at is when the row reached its final state.
-- name: ListRegistrationsWithRecentDeadDeliveries :many
SELECT d.business_id, d.webhook_id, w.url,
       COUNT(*) AS dead_count,
       MAX(d.updated_at) AS last_dead_at
FROM webhook_deliveries d
JOIN hub_webhooks w ON w.id = d.webhook_id AND w.business_id = d.business_id
WHERE d.status = 'dead'
  AND d.updated_at >= DATE_SUB(NOW(), INTERVAL sqlc.arg('within_hours') HOUR)
GROUP BY d.business_id, d.webhook_id, w.url
ORDER BY d.webhook_id ASC
LIMIT ?;

-- cross-schema read from users and shared — who to tell about a failing
-- registration.
--
-- hub_webhooks.owner_id is the user who created the registration, so they are
-- the person who set it up and the one who should hear it broke. Gated on an
-- active authorization_business_members row for the registration's own
-- business, so somebody who has since left is not sent a current tenant's
-- failures.
-- name: GetWebhookNotificationRecipient :one
SELECT u.id, COALESCE(u.email, '') AS email, COALESCE(u.full_name, '') AS full_name
FROM hub_webhooks w
JOIN users u ON u.id = w.owner_id
JOIN authorization_business_members m
  ON m.user_id = u.id AND m.business_id = w.business_id AND m.status = 'active'
WHERE w.id = ? AND w.business_id = ?
LIMIT 1;
