//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `GetSettings` (Go `GetSettingsParams`).
pub struct GetSettingsParams {
    pub business_id: i64,
}

/// `GetSettings`: `one`.
pub async fn get_settings<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetSettingsParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, business_id, auto_assign, sla_hours, business_hours, notification_email, created_at, updated_at\nFROM support_settings WHERE business_id = ?")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `UpsertSettings` (Go `UpsertSettingsParams`).
pub struct UpsertSettingsParams {
    pub business_id: i64,
    pub auto_assign: Option<bool>,
    pub sla_hours: Option<i32>,
    pub business_hours: String,
    pub notification_email: Option<String>,
}

/// `UpsertSettings`: `exec`.
pub async fn upsert_settings<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpsertSettingsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO support_settings (business_id, auto_assign, sla_hours, business_hours, notification_email)\nVALUES (?, ?, ?, ?, ?)\nON DUPLICATE KEY UPDATE auto_assign = VALUES(auto_assign), sla_hours = VALUES(sla_hours),\nbusiness_hours = VALUES(business_hours), notification_email = VALUES(notification_email)")
        .bind(params.business_id)
        .bind(params.auto_assign)
        .bind(params.sla_hours)
        .bind(&params.business_hours)
        .bind(&params.notification_email)
        .execute(exec)
        .await
}
