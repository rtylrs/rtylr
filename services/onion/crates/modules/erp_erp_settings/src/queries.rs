//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `GetErpSettings` (Go `GetErpSettingsParams`).
pub struct GetErpSettingsParams {
    pub business_id: i64,
}

/// `GetErpSettings`: `one`.
pub async fn get_erp_settings<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetErpSettingsParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, business_id, currency, tax_rate, low_stock_threshold, default_warehouse_id, notify_on_low_stock, notify_on_expiry, created_at, updated_at\nFROM erp_business_settings\nWHERE business_id = ?\nLIMIT 1")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `UpsertErpSettings` (Go `UpsertErpSettingsParams`).
pub struct UpsertErpSettingsParams {
    pub business_id: i64,
    pub currency: String,
    pub tax_rate: f64,
    pub low_stock_threshold: f64,
    pub default_warehouse_id: i64,
    pub notify_on_low_stock: bool,
    pub notify_on_expiry: bool,
}

/// `UpsertErpSettings`: `exec`.
pub async fn upsert_erp_settings<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpsertErpSettingsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO erp_business_settings\n  (business_id, currency, tax_rate, low_stock_threshold, default_warehouse_id, notify_on_low_stock, notify_on_expiry)\nVALUES (?, ?, ?, ?, ?, ?, ?)\nON DUPLICATE KEY UPDATE\n  currency = VALUES(currency),\n  tax_rate = VALUES(tax_rate),\n  low_stock_threshold = VALUES(low_stock_threshold),\n  default_warehouse_id = VALUES(default_warehouse_id),\n  notify_on_low_stock = VALUES(notify_on_low_stock),\n  notify_on_expiry = VALUES(notify_on_expiry),\n  updated_at = CURRENT_TIMESTAMP")
        .bind(params.business_id)
        .bind(&params.currency)
        .bind(params.tax_rate)
        .bind(params.low_stock_threshold)
        .bind(params.default_warehouse_id)
        .bind(params.notify_on_low_stock)
        .bind(params.notify_on_expiry)
        .execute(exec)
        .await
}
