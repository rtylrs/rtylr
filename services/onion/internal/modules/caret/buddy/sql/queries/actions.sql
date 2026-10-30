-- name: InsertAction :execresult
INSERT IGNORE INTO caret_buddy_actions (run_id, business_id, tool_name, input, fingerprint, expires_at)
VALUES (?, ?, ?, ?, ?, DATE_ADD(NOW(), INTERVAL CAST(sqlc.arg(ttl_hours) AS SIGNED) HOUR));

-- name: ListActionsByBusiness :many
SELECT * FROM caret_buddy_actions
WHERE business_id = ? AND status = ?
ORDER BY id DESC
LIMIT ?;

-- name: GetAction :one
SELECT * FROM caret_buddy_actions WHERE id = ? AND business_id = ?;

-- name: ExpireActions :execrows
UPDATE caret_buddy_actions
SET status = 'expired'
WHERE status = 'pending' AND expires_at < NOW();

-- name: ListAllActionsByBusiness :many
SELECT * FROM caret_buddy_actions
WHERE business_id = ?
ORDER BY id DESC
LIMIT ?;

-- name: DecideAction :execrows
UPDATE caret_buddy_actions
SET status = ?, decided_by = ?, decided_at = NOW()
WHERE id = ? AND business_id = ? AND status = 'pending' AND expires_at > NOW();

-- name: FinishAction :execrows
UPDATE caret_buddy_actions
SET status = ?, result = ?
WHERE id = ? AND business_id = ? AND status = 'approved' AND decided_by = ?;
