-- name: SearchPOSCustomers :many
SELECT id, business_id, first_name, last_name, email, phone
FROM crm_customers
WHERE business_id = ?
  AND (
    phone = ?
    OR email = ?
    OR (? <> '' AND MATCH(first_name, last_name, email, phone) AGAINST (? IN BOOLEAN MODE))
  )
ORDER BY first_name LIMIT 20;
