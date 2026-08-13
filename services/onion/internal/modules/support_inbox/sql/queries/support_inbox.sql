-- name: CountInboxTickets :one
SELECT COUNT(*) FROM support_tickets WHERE business_id = ? AND status IN ('open','pending');

-- name: ListInbox :many
SELECT t.id, t.customer_name, t.customer_email, t.subject, t.status, t.priority, t.channel,
       last_message.body AS last_message,
       last_message.created_at AS last_message_at,
       COALESCE(message_stats.message_count, 0) AS message_count
FROM (
    SELECT id, customer_name, customer_email, subject, status, priority, channel, updated_at
    FROM support_tickets tickets
    WHERE tickets.business_id = ? AND tickets.status IN ('open','pending')
    ORDER BY updated_at DESC, id DESC
    LIMIT ? OFFSET ?
) AS t
LEFT JOIN (
    SELECT m.ticket_id, MAX(m.created_at) AS last_message_at, COUNT(*) AS message_count
    FROM support_messages m
    INNER JOIN (
        SELECT id
        FROM support_tickets tickets
        WHERE tickets.business_id = ? AND tickets.status IN ('open','pending')
        ORDER BY updated_at DESC, id DESC
        LIMIT ? OFFSET ?
    ) AS page_tickets ON page_tickets.id = m.ticket_id
    GROUP BY m.ticket_id
) AS message_stats ON message_stats.ticket_id = t.id
LEFT JOIN support_messages last_message
    ON last_message.ticket_id = t.id
   AND last_message.created_at = message_stats.last_message_at
LEFT JOIN support_messages newer
    ON newer.ticket_id = last_message.ticket_id
   AND newer.created_at = last_message.created_at
   AND newer.id > last_message.id
WHERE newer.id IS NULL
GROUP BY t.id, t.customer_name, t.customer_email, t.subject, t.status, t.priority, t.channel, t.updated_at,
         last_message.id, last_message.body, last_message.created_at, message_stats.message_count
ORDER BY t.updated_at DESC, t.id DESC;

-- name: UnreadCount :one
SELECT COUNT(*) FROM support_tickets WHERE business_id = ? AND status = 'open';
