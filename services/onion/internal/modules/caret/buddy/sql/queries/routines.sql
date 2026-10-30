-- name: EnsureBuddy :exec
INSERT IGNORE INTO caret_buddies (business_id, run_as_user_id, enabled, instructions)
VALUES (?, ?, 0, '');

-- name: InsertRoutine :execresult
INSERT INTO caret_buddy_routines (buddy_id, business_id, user_id, name, instructions, every_minutes, enabled, next_run_at)
VALUES (?, ?, ?, ?, ?, ?, 1, ?);

-- name: ListRoutinesByOwner :many
SELECT * FROM caret_buddy_routines
WHERE business_id = ? AND user_id = ?
ORDER BY id;

-- name: GetRoutine :one
SELECT * FROM caret_buddy_routines
WHERE id = ? AND business_id = ? AND user_id = ?;

-- name: GetRoutineByID :one
SELECT * FROM caret_buddy_routines WHERE id = ?;

-- name: UpdateRoutine :execrows
UPDATE caret_buddy_routines
SET name = ?, instructions = ?, every_minutes = ?, enabled = ?, next_run_at = ?
WHERE id = ? AND business_id = ? AND user_id = ?;

-- name: DeleteRoutine :execrows
DELETE FROM caret_buddy_routines
WHERE id = ? AND business_id = ? AND user_id = ?;

-- name: ListDueRoutinesForUpdate :many
SELECT * FROM caret_buddy_routines
WHERE enabled = 1 AND next_run_at IS NOT NULL AND next_run_at <= NOW()
ORDER BY next_run_at
LIMIT ?
FOR UPDATE;

-- name: AdvanceRoutine :exec
UPDATE caret_buddy_routines
SET last_run_at = NOW(),
    next_run_at = DATE_ADD(
      next_run_at,
      INTERVAL (FLOOR(TIMESTAMPDIFF(MINUTE, next_run_at, NOW()) / every_minutes) + 1) * every_minutes MINUTE
    )
WHERE id = ?;

-- name: InsertRoutineRun :execresult
INSERT INTO caret_buddy_runs (buddy_id, business_id, routine_id, trigger_kind) VALUES (?, ?, ?, 'routine');
