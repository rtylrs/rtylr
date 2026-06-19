-- name: CountFeedback :one
SELECT COUNT(*) FROM crm_feedback WHERE business_id = ?;

-- name: CountFeedbackByType :one
SELECT COUNT(*) FROM crm_feedback WHERE business_id = ? AND type = ?;

-- name: ListFeedback :many
SELECT id, owner_id, business_id, customer_id, type, score, comment, source, created_at
FROM crm_feedback WHERE business_id = ?
ORDER BY created_at DESC LIMIT ? OFFSET ?;

-- name: ListFeedbackByType :many
SELECT id, owner_id, business_id, customer_id, type, score, comment, source, created_at
FROM crm_feedback WHERE business_id = ? AND type = ?
ORDER BY created_at DESC LIMIT ? OFFSET ?;

-- name: CreateFeedback :execresult
INSERT INTO crm_feedback (owner_id, business_id, customer_id, type, score, comment, source)
VALUES (?, ?, ?, ?, ?, ?, ?);
