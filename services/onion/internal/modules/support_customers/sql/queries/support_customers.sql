-- name: CountCustomers :one
SELECT COUNT(*) FROM support_customers WHERE business_id = ?;

-- name: ListCustomers :many
SELECT id, business_id, name, email, phone, customer_id, total_tickets, last_ticket_at, created_at
FROM support_customers WHERE business_id = ?
ORDER BY created_at DESC LIMIT ? OFFSET ?;

-- name: GetCustomerByID :one
SELECT id, business_id, name, email, phone, customer_id, total_tickets, last_ticket_at, created_at
FROM support_customers WHERE business_id = ? AND id = ?;
