-- name: UpsertBuddy :exec
INSERT INTO caret_buddies (business_id, run_as_user_id, enabled, instructions, interval_minutes, daily_run_cap, next_run_at)
VALUES (?, ?, ?, ?, ?, ?, CASE WHEN CAST(sqlc.arg(enabled_flag) AS SIGNED) = 1 THEN NOW() ELSE NULL END)
ON DUPLICATE KEY UPDATE
  run_as_user_id = VALUES(run_as_user_id),
  enabled = VALUES(enabled),
  instructions = VALUES(instructions),
  interval_minutes = VALUES(interval_minutes),
  daily_run_cap = VALUES(daily_run_cap),
  next_run_at = CASE
    WHEN VALUES(enabled) = 0 THEN NULL
    WHEN next_run_at IS NULL THEN NOW()
    ELSE next_run_at
  END;

-- name: GetBuddyByBusiness :one
SELECT * FROM caret_buddies WHERE business_id = ?;

-- name: GetBuddy :one
SELECT * FROM caret_buddies WHERE id = ?;

-- name: ListDueBuddiesForUpdate :many
SELECT * FROM caret_buddies
WHERE enabled = 1 AND next_run_at IS NOT NULL AND next_run_at <= NOW()
ORDER BY next_run_at
LIMIT ?
FOR UPDATE;

-- name: AdvanceBuddy :exec
UPDATE caret_buddies
SET last_run_at = NOW(),
    next_run_at = DATE_ADD(NOW(), INTERVAL interval_minutes MINUTE)
WHERE id = ?;

-- name: CountRunsSince :one
SELECT COUNT(*) FROM caret_buddy_runs
WHERE buddy_id = ? AND created_at >= DATE_SUB(NOW(), INTERVAL 24 HOUR);

-- name: GetBuddyByBusinessForUpdate :one
SELECT * FROM caret_buddies WHERE business_id = ? FOR UPDATE;

-- name: RecordBuddyEvent :execresult
INSERT IGNORE INTO caret_buddy_events (event_id, buddy_id) VALUES (?, ?);

-- name: CountQueuedEventRuns :one
SELECT COUNT(*) FROM caret_buddy_runs
WHERE buddy_id = ? AND trigger_kind = 'event' AND status = 'queued';
