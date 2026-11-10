//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `CountInboxTickets` (Go `CountInboxTicketsParams`).
pub struct CountInboxTicketsParams {
    pub business_id: i64,
}

/// `CountInboxTickets`: `one`.
pub async fn count_inbox_tickets<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountInboxTicketsParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) FROM support_tickets WHERE business_id = ? AND status IN ('open','pending')")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `ListInbox` (Go `ListInboxParams`).
pub struct ListInboxParams {
    pub business_id: i64,
    pub limit: i32,
    pub offset: i32,
    pub business_id_2: i64,
    pub limit_2: i32,
    pub offset_2: i32,
}

/// `ListInbox`: `many`.
pub async fn list_inbox<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListInboxParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT t.id, t.customer_name, t.customer_email, t.subject, t.status, t.priority, t.channel,\n       last_message.body AS last_message,\n       last_message.created_at AS last_message_at,\n       COALESCE(message_stats.message_count, 0) AS message_count\nFROM (\n    SELECT id, customer_name, customer_email, subject, status, priority, channel, updated_at\n    FROM support_tickets tickets\n    WHERE tickets.business_id = ? AND tickets.status IN ('open','pending')\n    ORDER BY updated_at DESC, id DESC\n    LIMIT ? OFFSET ?\n) AS t\nLEFT JOIN (\n    SELECT m.ticket_id, MAX(m.created_at) AS last_message_at, COUNT(*) AS message_count\n    FROM support_messages m\n    INNER JOIN (\n        SELECT id\n        FROM support_tickets tickets\n        WHERE tickets.business_id = ? AND tickets.status IN ('open','pending')\n        ORDER BY updated_at DESC, id DESC\n        LIMIT ? OFFSET ?\n    ) AS page_tickets ON page_tickets.id = m.ticket_id\n    GROUP BY m.ticket_id\n) AS message_stats ON message_stats.ticket_id = t.id\nLEFT JOIN support_messages last_message\n    ON last_message.ticket_id = t.id\n   AND last_message.created_at = message_stats.last_message_at\nLEFT JOIN support_messages newer\n    ON newer.ticket_id = last_message.ticket_id\n   AND newer.created_at = last_message.created_at\n   AND newer.id > last_message.id\nWHERE newer.id IS NULL\nGROUP BY t.id, t.customer_name, t.customer_email, t.subject, t.status, t.priority, t.channel, t.updated_at,\n         last_message.id, last_message.body, last_message.created_at, message_stats.message_count\nORDER BY t.updated_at DESC, t.id DESC")
        .bind(params.business_id)
        .bind(params.limit)
        .bind(params.offset)
        .bind(params.business_id_2)
        .bind(params.limit_2)
        .bind(params.offset_2)
        .execute(exec)
        .await
}

/// Params for `UnreadCount` (Go `UnreadCountParams`).
pub struct UnreadCountParams {
    pub business_id: i64,
}

/// `UnreadCount`: `one`.
pub async fn unread_count<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UnreadCountParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) FROM support_tickets WHERE business_id = ? AND status = 'open'")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}
