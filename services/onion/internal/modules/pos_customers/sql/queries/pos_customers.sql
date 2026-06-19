-- name: SearchPOSCustomers :many
SELECT id, business_id, first_name, last_name, email, phone
FROM crm_customers
WHERE business_id = ? AND (phone LIKE ? OR email LIKE ? OR first_name LIKE ? OR last_name LIKE ?)
ORDER BY first_name LIMIT 20;
