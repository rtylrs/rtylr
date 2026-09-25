-- name: ListFeedback :many
-- cross-schema read from crm_customers — resolves a display name for each feedback row
SELECT f.id, f.owner_id, f.business_id, f.customer_id,
       COALESCE(TRIM(CONCAT_WS(' ', c.first_name, c.last_name)), '') AS customer_name,
       f.type, f.score, f.comment, f.source, f.created_at
FROM crm_feedback f LEFT JOIN crm_customers c ON c.id = f.customer_id
WHERE f.business_id = ?
ORDER BY f.created_at DESC LIMIT ? OFFSET ?;

-- name: CreateFeedback :execresult
INSERT INTO crm_feedback (owner_id, business_id, customer_id, type, score, comment, source)
VALUES (?, ?, ?, ?, ?, ?, ?);
