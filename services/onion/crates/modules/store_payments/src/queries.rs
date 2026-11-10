//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `CreatePaymentMethod` (Go `CreatePaymentMethodParams`).
pub struct CreatePaymentMethodParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub provider: String,
    pub config: String,
    pub is_active: Option<bool>,
}

/// `CreatePaymentMethod`: `execresult`.
pub async fn create_payment_method<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreatePaymentMethodParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO store_payment_methods (owner_id, business_id, provider, config, is_active) VALUES (?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(&params.provider)
        .bind(&params.config)
        .bind(params.is_active)
        .execute(exec)
        .await
}

/// Params for `ListPaymentMethods` (Go `ListPaymentMethodsParams`).
pub struct ListPaymentMethodsParams {
    pub business_id: i64,
}

/// `ListPaymentMethods`: `many`.
pub async fn list_payment_methods<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListPaymentMethodsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, provider, COALESCE(config, '{}') AS config, is_active, created_at\nFROM store_payment_methods WHERE business_id = ? ORDER BY provider ASC\nLIMIT 200")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `UpdatePaymentMethod` (Go `UpdatePaymentMethodParams`).
pub struct UpdatePaymentMethodParams {
    pub provider: String,
    pub config: String,
    pub is_active: Option<bool>,
    pub business_id: i64,
    pub id: i64,
}

/// `UpdatePaymentMethod`: `exec`.
pub async fn update_payment_method<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdatePaymentMethodParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE store_payment_methods SET provider = ?, config = ?, is_active = ? WHERE business_id = ? AND id = ?")
        .bind(&params.provider)
        .bind(&params.config)
        .bind(params.is_active)
        .bind(params.business_id)
        .bind(params.id)
        .execute(exec)
        .await
}
