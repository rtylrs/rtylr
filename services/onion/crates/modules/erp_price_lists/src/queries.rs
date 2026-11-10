//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `CreatePriceList` (Go `CreatePriceListParams`).
pub struct CreatePriceListParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub name: String,
    pub currency: String,
    pub is_default: bool,
}

/// `CreatePriceList`: `execresult`.
pub async fn create_price_list<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreatePriceListParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO erp_price_lists (owner_id, business_id, name, currency, is_default)\nVALUES (?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(&params.name)
        .bind(&params.currency)
        .bind(params.is_default)
        .execute(exec)
        .await
}

/// Params for `ListPriceLists` (Go `ListPriceListsParams`).
pub struct ListPriceListsParams {
    pub business_id: i64,
}

/// `ListPriceLists`: `many`.
pub async fn list_price_lists<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListPriceListsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, name, currency, is_default, created_at, updated_at\nFROM erp_price_lists\nWHERE business_id = ?\nORDER BY name ASC")
        .bind(params.business_id)
        .execute(exec)
        .await
}
