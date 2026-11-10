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
    sqlx::query("SELECT id, owner_id, business_id, default_consent, from_email, from_name, sms_sender_id, timezone, created_at, updated_at\nFROM crm_settings WHERE business_id = ?")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `UpsertSettings` (Go `UpsertSettingsParams`).
pub struct UpsertSettingsParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub default_consent: Option<bool>,
    pub from_email: Option<String>,
    pub from_name: Option<String>,
    pub sms_sender_id: Option<String>,
    pub timezone: Option<String>,
}

/// `UpsertSettings`: `exec`.
pub async fn upsert_settings<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpsertSettingsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO crm_settings\n  (owner_id, business_id, default_consent, from_email, from_name, sms_sender_id, timezone)\nVALUES (?, ?, ?, ?, ?, ?, ?)\nON DUPLICATE KEY UPDATE\n  default_consent = VALUES(default_consent),\n  from_email = VALUES(from_email),\n  from_name = VALUES(from_name),\n  sms_sender_id = VALUES(sms_sender_id),\n  timezone = VALUES(timezone)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(params.default_consent)
        .bind(&params.from_email)
        .bind(&params.from_name)
        .bind(&params.sms_sender_id)
        .bind(&params.timezone)
        .execute(exec)
        .await
}
