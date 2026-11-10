//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `CountCustomers` (Go `CountCustomersParams`).
pub struct CountCustomersParams {
    pub business_id: i64,
}

/// `CountCustomers`: `one`.
pub async fn count_customers<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountCustomersParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) FROM support_customers WHERE business_id = ?")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetCustomerByID` (Go `GetCustomerByIDParams`).
pub struct GetCustomerByIDParams {
    pub business_id: i64,
    pub id: i64,
}

/// `GetCustomerByID`: `one`.
pub async fn get_customer_by_id<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetCustomerByIDParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, business_id, name, email, phone, customer_id, total_tickets, last_ticket_at, created_at\nFROM support_customers WHERE business_id = ? AND id = ?")
        .bind(params.business_id)
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `ListCustomers` (Go `ListCustomersParams`).
pub struct ListCustomersParams {
    pub business_id: i64,
    pub limit: i32,
    pub offset: i32,
}

/// `ListCustomers`: `many`.
pub async fn list_customers<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListCustomersParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, business_id, name, email, phone, customer_id, total_tickets, last_ticket_at, created_at\nFROM support_customers WHERE business_id = ?\nORDER BY created_at DESC LIMIT ? OFFSET ?")
        .bind(params.business_id)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}
