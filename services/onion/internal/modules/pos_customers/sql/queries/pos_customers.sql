-- name: SearchPOSCustomersExact :many
SELECT id, business_id, first_name, last_name, email, phone
FROM crm_customers
WHERE business_id = ?
  AND (phone = ? OR email = ?)
ORDER BY first_name LIMIT 20;
