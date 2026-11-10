//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `CreateShippingMethod` (Go `CreateShippingMethodParams`).
pub struct CreateShippingMethodParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub name: String,
    pub price: String,
    pub estimated_days: Option<i32>,
    pub is_active: Option<bool>,
}

/// `CreateShippingMethod`: `execresult`.
pub async fn create_shipping_method<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateShippingMethodParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO store_shipping_methods (owner_id, business_id, name, price, estimated_days, is_active) VALUES (?, ?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(&params.name)
        .bind(&params.price)
        .bind(params.estimated_days)
        .bind(params.is_active)
        .execute(exec)
        .await
}

/// Params for `DeleteShippingMethod` (Go `DeleteShippingMethodParams`).
pub struct DeleteShippingMethodParams {
    pub business_id: i64,
    pub id: i64,
}

/// `DeleteShippingMethod`: `exec`.
pub async fn delete_shipping_method<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DeleteShippingMethodParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("DELETE FROM store_shipping_methods WHERE business_id = ? AND id = ?")
        .bind(params.business_id)
        .bind(params.id)
        .execute(exec)
        .await
}

/// Params for `ListShippingMethods` (Go `ListShippingMethodsParams`).
pub struct ListShippingMethodsParams {
    pub business_id: i64,
}

/// `ListShippingMethods`: `many`.
pub async fn list_shipping_methods<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListShippingMethodsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, name, price, estimated_days, is_active, created_at\nFROM store_shipping_methods WHERE business_id = ? ORDER BY name ASC\nLIMIT 200")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `UpdateShippingMethod` (Go `UpdateShippingMethodParams`).
pub struct UpdateShippingMethodParams {
    pub name: String,
    pub price: String,
    pub estimated_days: Option<i32>,
    pub is_active: Option<bool>,
    pub business_id: i64,
    pub id: i64,
}

/// `UpdateShippingMethod`: `exec`.
pub async fn update_shipping_method<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateShippingMethodParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE store_shipping_methods SET name = ?, price = ?, estimated_days = ?, is_active = ? WHERE business_id = ? AND id = ?")
        .bind(&params.name)
        .bind(&params.price)
        .bind(params.estimated_days)
        .bind(params.is_active)
        .bind(params.business_id)
        .bind(params.id)
        .execute(exec)
        .await
}
