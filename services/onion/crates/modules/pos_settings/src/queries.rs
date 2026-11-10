//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `GetPOSSettings` (Go `GetPOSSettingsParams`).
pub struct GetPOSSettingsParams {
    pub business_id: i64,
}

/// `GetPOSSettings`: `one`.
pub async fn get_pos_settings<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetPOSSettingsParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, business_id, business_profile, theme, features, receipt_footer, created_at, updated_at\nFROM pos_business_settings\nWHERE business_id = ?\nLIMIT 1")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `UpsertPOSSettings` (Go `UpsertPOSSettingsParams`).
pub struct UpsertPOSSettingsParams {
    pub business_id: i64,
    pub business_profile: String,
    pub theme: String,
    pub features: String,
    pub receipt_footer: Option<String>,
}

/// `UpsertPOSSettings`: `exec`.
pub async fn upsert_pos_settings<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpsertPOSSettingsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO pos_business_settings (business_id, business_profile, theme, features, receipt_footer)\nVALUES (?, ?, ?, ?, ?)\nON DUPLICATE KEY UPDATE\n  business_profile = VALUES(business_profile),\n  theme = VALUES(theme),\n  features = VALUES(features),\n  receipt_footer = VALUES(receipt_footer),\n  updated_at = CURRENT_TIMESTAMP")
        .bind(params.business_id)
        .bind(&params.business_profile)
        .bind(&params.theme)
        .bind(&params.features)
        .bind(&params.receipt_footer)
        .execute(exec)
        .await
}
