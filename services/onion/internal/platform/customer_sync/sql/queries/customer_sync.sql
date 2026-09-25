-- name: FindCRMCustomerByContact :one
SELECT id FROM crm_customers
WHERE business_id = sqlc.arg(business_id)
  AND ((sqlc.arg(email) <> '' AND LOWER(email) = LOWER(sqlc.arg(email)))
    OR (sqlc.arg(phone_input) <> '' AND phone = sqlc.arg(phone_input)))
ORDER BY CASE WHEN sqlc.arg(email) <> '' AND LOWER(email) = LOWER(sqlc.arg(email)) THEN 0 ELSE 1 END
LIMIT 1;

-- name: GetActiveBusinessOwnerID :one
SELECT owner_id FROM businesses WHERE id = ? AND is_active = TRUE LIMIT 1;

-- name: InsertCRMCustomerFromSync :execresult
INSERT INTO crm_customers
  (owner_id, business_id, first_name, last_name, email, phone, total_spent, visit_count, last_visit_at, loyalty_tier, tags, is_active)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 'bronze', JSON_ARRAY(), TRUE);

-- name: UpdateCRMCustomerFromSync :exec
UPDATE crm_customers
SET first_name = CASE WHEN sqlc.arg(first_name) <> '' THEN sqlc.arg(first_name) ELSE first_name END,
    last_name = CASE WHEN sqlc.arg(last_name) <> '' THEN sqlc.arg(last_name) ELSE last_name END,
    email = CASE WHEN sqlc.arg(email) <> '' THEN sqlc.arg(email) ELSE email END,
    phone = CASE WHEN sqlc.arg(phone) <> '' THEN sqlc.arg(phone) ELSE phone END,
    total_spent = total_spent + sqlc.arg(total_spent_delta),
    visit_count = visit_count + sqlc.arg(visit_count_delta),
    last_visit_at = COALESCE(sqlc.narg(last_visit_at), last_visit_at),
    is_active = TRUE
WHERE business_id = sqlc.arg(business_id) AND id = sqlc.arg(customer_id);
