-- Fan-out for the platform event outbox. The dispatcher reads events nobody has
-- expanded yet, works out who wants each one, and writes a queue row per
-- interested target. It never delivers anything itself.

-- name: ListUndispatchedEvents :many
SELECT id, business_id, event_type, subject_type, subject_id
FROM platform_events
WHERE dispatched_at IS NULL
ORDER BY id ASC
LIMIT ?;

-- cross-schema read from hub — webhook registrations decide who receives an event.
-- Scoped by business_id so an event can only ever reach registrations belonging
-- to the tenant that produced it. events is a JSON array of event type strings;
-- JSON_CONTAINS with JSON_QUOTE matches one element exactly rather than by
-- substring, so "order_created" does not match "order_created_v2".
-- name: ListWebhooksForEvent :many
SELECT id
FROM hub_webhooks
WHERE business_id = ?
  AND is_active = TRUE
  AND JSON_CONTAINS(events, JSON_QUOTE(sqlc.arg('event_type')))
ORDER BY id ASC;

-- cross-schema read from crm — automations subscribe by trigger_type.
-- Same tenant scoping as the webhook lookup above.
-- name: ListAutomationsForEvent :many
SELECT id
FROM crm_automations
WHERE business_id = ?
  AND is_active = TRUE
  AND trigger_type = ?
ORDER BY id ASC;

-- ON DUPLICATE KEY UPDATE id = id rather than INSERT IGNORE: this must be a
-- no-op when the delivery already exists, without also swallowing genuine
-- errors the way IGNORE does. The uniqueness is what makes a re-run of the
-- dispatcher safe.
-- name: EnqueueWebhookDelivery :exec
INSERT INTO webhook_deliveries (business_id, webhook_id, event_id)
VALUES (?, ?, ?)
ON DUPLICATE KEY UPDATE id = id;

-- name: EnqueueAutomationRun :exec
INSERT INTO automation_runs (business_id, automation_id, event_id)
VALUES (?, ?, ?)
ON DUPLICATE KEY UPDATE id = id;

-- The dispatched_at IS NULL guard makes the mark itself idempotent, so two
-- dispatchers racing on the same event cannot both claim to have expanded it.
-- name: MarkEventDispatched :execresult
UPDATE platform_events
SET dispatched_at = NOW()
WHERE id = ? AND dispatched_at IS NULL;
