-- name: CreatePerformanceReview :exec
INSERT INTO hr_performance_reviews (owner_id, business_id, employee_id, review_cycle_id, reviewer_employee_id, review_type, status)
VALUES (?, ?, ?, ?, ?, ?, 'draft');

-- name: GetPerformanceReview :one
SELECT id, owner_id, business_id, employee_id, review_cycle_id, reviewer_employee_id, review_type, overall_rating, narrative, status, shared_with_employee_at, created_at, updated_at
FROM hr_performance_reviews
WHERE business_id = ? AND id = ?;

-- name: ListPerformanceReviewsByCycle :many
SELECT id, owner_id, business_id, employee_id, review_cycle_id, reviewer_employee_id, review_type, overall_rating, narrative, status, shared_with_employee_at, created_at, updated_at
FROM hr_performance_reviews
WHERE business_id = ? AND review_cycle_id = ?
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: CountPerformanceReviewsByCycle :one
SELECT COUNT(*) AS total
FROM hr_performance_reviews
WHERE business_id = ? AND review_cycle_id = ?;

-- name: ListPerformanceReviewsByCycleAndEmployee :many
SELECT id, owner_id, business_id, employee_id, review_cycle_id, reviewer_employee_id, review_type, overall_rating, narrative, status, shared_with_employee_at, created_at, updated_at
FROM hr_performance_reviews
WHERE business_id = ? AND review_cycle_id = ? AND employee_id = ?
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: CountPerformanceReviewsByCycleAndEmployee :one
SELECT COUNT(*) AS total
FROM hr_performance_reviews
WHERE business_id = ? AND review_cycle_id = ? AND employee_id = ?;

-- name: UpdateReviewSubmission :exec
UPDATE hr_performance_reviews
SET overall_rating = ?, narrative = ?, status = 'submitted'
WHERE business_id = ? AND id = ?;

-- name: UpdateReviewStatus :exec
UPDATE hr_performance_reviews
SET status = ?
WHERE business_id = ? AND id = ?;

-- name: ShareReview :exec
UPDATE hr_performance_reviews
SET status = 'shared', shared_with_employee_at = NOW()
WHERE business_id = ? AND id = ?;
