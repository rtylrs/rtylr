-- name: CreateFeedback :execresult
INSERT INTO crm_feedback (owner_id, business_id, customer_id, type, score, comment, source)
VALUES (?, ?, ?, ?, ?, ?, ?);
