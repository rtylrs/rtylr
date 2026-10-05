-- name: InsertFact :exec
INSERT INTO crm_outbox (business_id, event_type, subject_type, subject_id, payload, dedupe_key, occurred_at)
VALUES (?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE id = id;

-- name: PendingFacts :many
SELECT id, business_id, event_type, subject_type, subject_id, payload, dedupe_key, occurred_at
FROM crm_outbox WHERE relayed_at IS NULL ORDER BY id LIMIT ?;

-- name: MarkRelayed :exec
UPDATE crm_outbox SET relayed_at = CURRENT_TIMESTAMP WHERE id = ? AND relayed_at IS NULL;
