//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `CreateNotification` (Go `CreateNotificationParams`).
pub struct CreateNotificationParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub user_id: i64,
    pub title: String,
    pub body: Option<String>,
    pub r#type: String,
    pub link: Option<String>,
}

/// `CreateNotification`: `exec`.
pub async fn create_notification<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateNotificationParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO hub_notifications (owner_id, business_id, user_id, title, body, type, link)\nVALUES (?, ?, ?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(params.user_id)
        .bind(&params.title)
        .bind(&params.body)
        .bind(&params.r#type)
        .bind(&params.link)
        .execute(exec)
        .await
}

/// Params for `ListNotifications` (Go `ListNotificationsParams`).
pub struct ListNotificationsParams {
    pub business_id: i64,
    pub user_id: i64,
    pub cursor_created_at: Option<chrono::DateTime<chrono::Utc>>,
    pub cursor_id: Option<i64>,
    pub limit: i32,
}

/// `ListNotifications`: `many`.
pub async fn list_notifications<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListNotificationsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, user_id, title, COALESCE(body,'') AS body, type, is_read, COALESCE(link,'') AS link, created_at\nFROM hub_notifications\nWHERE business_id = ? AND user_id = ?\n  AND (\n    ? IS NULL\n    OR created_at < ?\n    OR (created_at = ? AND id < ?)\n  )\nORDER BY created_at DESC, id DESC\nLIMIT ?")
        .bind(params.business_id)
        .bind(params.user_id)
        .bind(params.cursor_created_at)
        .bind(params.cursor_created_at)
        .bind(params.cursor_created_at)
        .bind(params.cursor_id)
        .bind(params.limit)
        .execute(exec)
        .await
}

/// Params for `MarkAllNotificationsRead` (Go `MarkAllNotificationsReadParams`).
pub struct MarkAllNotificationsReadParams {
    pub business_id: i64,
    pub user_id: i64,
}

/// `MarkAllNotificationsRead`: `exec`.
pub async fn mark_all_notifications_read<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &MarkAllNotificationsReadParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE hub_notifications\nSET is_read = TRUE\nWHERE business_id = ? AND user_id = ? AND is_read = FALSE")
        .bind(params.business_id)
        .bind(params.user_id)
        .execute(exec)
        .await
}

/// Params for `MarkNotificationRead` (Go `MarkNotificationReadParams`).
pub struct MarkNotificationReadParams {
    pub id: i64,
    pub business_id: i64,
    pub user_id: i64,
}

/// `MarkNotificationRead`: `exec`.
pub async fn mark_notification_read<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &MarkNotificationReadParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE hub_notifications\nSET is_read = TRUE\nWHERE id = ? AND business_id = ? AND user_id = ?")
        .bind(params.id)
        .bind(params.business_id)
        .bind(params.user_id)
        .execute(exec)
        .await
}
