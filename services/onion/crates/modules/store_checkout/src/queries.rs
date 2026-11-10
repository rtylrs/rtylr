//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `CompleteCheckout` (Go `CompleteCheckoutParams`).
pub struct CompleteCheckoutParams {
    pub business_id: i64,
    pub id: i64,
}

/// `CompleteCheckout`: `execresult`.
pub async fn complete_checkout<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CompleteCheckoutParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE store_checkouts SET status = 'completed'\nWHERE business_id = ? AND id = ? AND status <> 'completed'")
        .bind(params.business_id)
        .bind(params.id)
        .execute(exec)
        .await
}

/// Params for `CreateCheckout` (Go `CreateCheckoutParams`).
pub struct CreateCheckoutParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub session_id: Option<String>,
    pub customer_email: Option<String>,
    pub items: String,
    pub subtotal: String,
    pub tax: String,
    pub total: String,
}

/// `CreateCheckout`: `execresult`.
pub async fn create_checkout<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateCheckoutParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO store_checkouts (owner_id, business_id, session_id, customer_email, items, subtotal, tax, total)\nVALUES (?, ?, ?, ?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(&params.session_id)
        .bind(&params.customer_email)
        .bind(&params.items)
        .bind(&params.subtotal)
        .bind(&params.tax)
        .bind(&params.total)
        .execute(exec)
        .await
}

/// Params for `GetCheckoutForComplete` (Go `GetCheckoutForCompleteParams`).
pub struct GetCheckoutForCompleteParams {
    pub business_id: i64,
    pub id: i64,
}

/// `GetCheckoutForComplete`: `one`.
pub async fn get_checkout_for_complete<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetCheckoutForCompleteParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT customer_email, total, status FROM store_checkouts WHERE business_id = ? AND id = ? ORDER BY id ASC LIMIT 1")
        .bind(params.business_id)
        .bind(params.id)
        .fetch_optional(exec)
        .await
}
