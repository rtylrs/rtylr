-- name: ListUndispatchedEvents :many
SELECT id, business_id, event_type, subject_type, subject_id, payload, occurred_at
FROM platform_events
WHERE dispatched_at IS NULL
ORDER BY id ASC
LIMIT ?;

-- name: MarkEventDispatched :execresult
UPDATE platform_events
SET dispatched_at = NOW()
WHERE id = ? AND dispatched_at IS NULL;

-- name: InsertEvent :exec
INSERT INTO platform_events
    (business_id, event_type, subject_type, subject_id, payload, dedupe_key, occurred_at)
VALUES (?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE id = id;
