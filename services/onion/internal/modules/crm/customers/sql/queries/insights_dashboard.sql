-- name: CountCustomersCreatedBetween :one
SELECT COUNT(*)
FROM crm_customers
WHERE business_id = sqlc.arg(business_id)
  AND created_at >= sqlc.arg(date_from)
  AND created_at <= sqlc.arg(date_to);
