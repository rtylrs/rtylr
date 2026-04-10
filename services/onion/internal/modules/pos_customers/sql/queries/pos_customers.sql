-- name: SearchPOSCustomers :many
SELECT id, first_name, last_name, email, phone
FROM crm_customers
WHERE owner_id = ? AND (phone LIKE ? OR email LIKE ? OR first_name LIKE ? OR last_name LIKE ?)
ORDER BY first_name LIMIT 20;
