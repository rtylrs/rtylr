-- name: CountOrdersCreatedBetween :one
SELECT COUNT(*)
FROM pos_orders
WHERE business_id = sqlc.arg(business_id)
  AND created_at >= sqlc.arg(date_from)
  AND created_at <= sqlc.arg(date_to);
