-- name: ListReviews :many
SELECT id, owner_id, business_id, staff_id, reviewer_id, period,
       rating, goals_rating, values_rating, skills_rating,
       notes, status, created_at, updated_at
FROM hr_reviews
WHERE business_id = ?
ORDER BY created_at DESC, id DESC
LIMIT 500;

-- name: CreateReview :execresult
INSERT INTO hr_reviews
(owner_id, business_id, staff_id, reviewer_id, period, rating, goals_rating, values_rating, skills_rating, notes, status)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);
