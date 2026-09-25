-- name: CreateReviewGoal :execresult
INSERT INTO hr_performance_review_goals (owner_id, business_id, employee_id, review_cycle_id, title, description, weight, success_criteria, status)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, 'active');

-- name: GetReviewGoal :one
SELECT id, owner_id, business_id, employee_id, review_cycle_id, title, description, weight, success_criteria, status, created_at, updated_at
FROM hr_performance_review_goals
WHERE business_id = ? AND id = ?;

-- name: UpdateReviewGoal :exec
UPDATE hr_performance_review_goals
SET title = ?, description = ?, weight = ?, success_criteria = ?, status = ?
WHERE business_id = ? AND id = ?;

-- name: ListReviewGoals :many
SELECT id, owner_id, business_id, employee_id, review_cycle_id, title, description, weight, success_criteria, status, created_at, updated_at
FROM hr_performance_review_goals
WHERE business_id = sqlc.arg(business_id)
  AND (sqlc.arg(employee_id) = 0 OR employee_id = sqlc.arg(employee_id))
  AND (sqlc.arg(review_cycle_id) = 0 OR review_cycle_id = sqlc.arg(review_cycle_id))
ORDER BY created_at ASC
LIMIT ? OFFSET ?;

-- name: CountReviewGoals :one
SELECT COUNT(*) AS total
FROM hr_performance_review_goals
WHERE business_id = sqlc.arg(business_id)
  AND (sqlc.arg(employee_id) = 0 OR employee_id = sqlc.arg(employee_id))
  AND (sqlc.arg(review_cycle_id) = 0 OR review_cycle_id = sqlc.arg(review_cycle_id));
