-- name: CountTickets :one
SELECT COUNT(*) FROM support_tickets WHERE business_id = ?;

-- name: CountTicketsByStatus :one
SELECT COUNT(*) FROM support_tickets WHERE business_id = ? AND status = ?;

-- name: ListTickets :many
SELECT id, business_id, customer_name, customer_email, subject, status, priority, assigned_to, channel, created_at, updated_at
FROM support_tickets WHERE business_id = ?
ORDER BY created_at DESC LIMIT ? OFFSET ?;

-- name: ListTicketsByStatus :many
SELECT id, business_id, customer_name, customer_email, subject, status, priority, assigned_to, channel, created_at, updated_at
FROM support_tickets WHERE business_id = ? AND status = ?
ORDER BY created_at DESC LIMIT ? OFFSET ?;

-- name: GetTicketByID :one
SELECT id, business_id, customer_name, customer_email, subject, status, priority, assigned_to, channel, created_at, updated_at
FROM support_tickets WHERE business_id = ? AND id = ?;

-- name: CreateTicket :execresult
INSERT INTO support_tickets (business_id, customer_name, customer_email, subject, status, priority, assigned_to, channel)
VALUES (?, ?, ?, ?, ?, ?, ?, ?);

-- name: UpdateTicket :exec
UPDATE support_tickets SET subject = ?, status = ?, priority = ?, assigned_to = ?, channel = ?
WHERE business_id = ? AND id = ?;

-- name: ListMessages :many
SELECT m.id, m.ticket_id, m.sender_type, m.sender_name, m.body, m.is_internal, m.created_at
FROM support_messages m
JOIN support_tickets t ON t.id = m.ticket_id
WHERE m.ticket_id = ? AND t.business_id = ?
ORDER BY m.created_at ASC;

-- name: CreateMessage :execresult
INSERT INTO support_messages (ticket_id, sender_type, sender_name, body, is_internal)
VALUES (?, ?, ?, ?, ?);
