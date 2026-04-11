-- name: CountInboxTickets :one
SELECT COUNT(*) FROM support_tickets WHERE business_id = ? AND status IN ('open','pending');

-- name: ListInbox :many
SELECT t.id, t.customer_name, t.customer_email, t.subject, t.status, t.priority, t.channel,
       (SELECT body FROM support_messages WHERE ticket_id = t.id ORDER BY created_at DESC LIMIT 1) AS last_message,
       (SELECT created_at FROM support_messages WHERE ticket_id = t.id ORDER BY created_at DESC LIMIT 1) AS last_message_at,
       (SELECT COUNT(*) FROM support_messages WHERE ticket_id = t.id) AS message_count
FROM support_tickets t
WHERE t.business_id = ? AND t.status IN ('open','pending')
ORDER BY t.updated_at DESC
LIMIT ? OFFSET ?;

-- name: UnreadCount :one
SELECT COUNT(*) FROM support_tickets WHERE business_id = ? AND status = 'open';
