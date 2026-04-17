-- name: CreateReviewCycle :execresult
INSERT INTO hr_performance_review_cycles (owner_id, name, cycle_type, period_start, period_end, status)
VALUES (?, ?, ?, ?, ?, 'draft');

-- name: GetReviewCycle :one
SELECT id, owner_id, name, cycle_type, period_start, period_end, status, created_at, updated_at
FROM hr_performance_review_cycles
WHERE owner_id = ? AND id = ?;

-- name: ListReviewCycles :many
SELECT id, owner_id, name, cycle_type, period_start, period_end, status, created_at, updated_at
FROM hr_performance_review_cycles
WHERE owner_id = ?
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: CountReviewCycles :one
SELECT COUNT(*) AS total
FROM hr_performance_review_cycles
WHERE owner_id = ?;

-- name: UpdateReviewCycleStatus :exec
UPDATE hr_performance_review_cycles
SET status = ?
WHERE owner_id = ? AND id = ?;
