-- name: InsertRun :execresult
INSERT INTO caret_buddy_runs (buddy_id, business_id, trigger_kind) VALUES (?, ?, ?);

-- name: ClaimNextRunForUpdate :one
SELECT * FROM caret_buddy_runs
WHERE status = 'queued'
ORDER BY id
LIMIT 1
FOR UPDATE;

-- name: MarkRunRunning :exec
UPDATE caret_buddy_runs
SET status = 'running', attempts = attempts + 1, started_at = NOW(),
    lease_until = DATE_ADD(NOW(), INTERVAL CAST(sqlc.arg(lease_seconds) AS SIGNED) SECOND)
WHERE id = ?;

-- name: FinishRun :exec
UPDATE caret_buddy_runs
SET status = ?, report = ?, error = ?, tool_calls = ?, tokens = ?,
    finished_at = NOW(), lease_until = NULL
WHERE id = ?;

-- name: RequeueExpiredRuns :execrows
UPDATE caret_buddy_runs
SET status = 'queued', lease_until = NULL
WHERE status = 'running' AND lease_until < NOW() AND attempts < ?;

-- name: FailExhaustedRuns :execrows
UPDATE caret_buddy_runs
SET status = 'failed', error = 'run was interrupted too many times', finished_at = NOW(), lease_until = NULL
WHERE status = 'running' AND lease_until < NOW() AND attempts >= ?;

-- name: ListRunsByBusiness :many
SELECT * FROM caret_buddy_runs
WHERE business_id = ?
ORDER BY id DESC
LIMIT ?;
