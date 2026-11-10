//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `CreateBookingService` (Go `CreateBookingServiceParams`).
pub struct CreateBookingServiceParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub name: String,
    pub description: Option<String>,
    pub duration_minutes: i32,
    pub price: String,
    pub color: Option<String>,
    pub is_active: Option<bool>,
}

/// `CreateBookingService`: `execresult`.
pub async fn create_booking_service<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateBookingServiceParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO booking_services (owner_id, business_id, name, description, duration_minutes, price, color, is_active) VALUES (?, ?, ?, ?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(&params.name)
        .bind(&params.description)
        .bind(params.duration_minutes)
        .bind(&params.price)
        .bind(&params.color)
        .bind(params.is_active)
        .execute(exec)
        .await
}

/// Params for `DeleteBookingService` (Go `DeleteBookingServiceParams`).
pub struct DeleteBookingServiceParams {
    pub business_id: i64,
    pub id: i64,
}

/// `DeleteBookingService`: `exec`.
pub async fn delete_booking_service<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DeleteBookingServiceParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("DELETE FROM booking_services WHERE business_id = ? AND id = ?")
        .bind(params.business_id)
        .bind(params.id)
        .execute(exec)
        .await
}

/// Params for `ListBookingServices` (Go `ListBookingServicesParams`).
pub struct ListBookingServicesParams {
    pub business_id: i64,
}

/// `ListBookingServices`: `many`.
pub async fn list_booking_services<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListBookingServicesParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, name, description, duration_minutes, price, color, is_active, created_at, updated_at\nFROM booking_services WHERE business_id = ? ORDER BY name ASC LIMIT 100")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `UpdateBookingService` (Go `UpdateBookingServiceParams`).
pub struct UpdateBookingServiceParams {
    pub name: String,
    pub description: Option<String>,
    pub duration_minutes: i32,
    pub price: String,
    pub color: Option<String>,
    pub is_active: Option<bool>,
    pub business_id: i64,
    pub id: i64,
}

/// `UpdateBookingService`: `exec`.
pub async fn update_booking_service<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateBookingServiceParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE booking_services SET name = ?, description = ?, duration_minutes = ?, price = ?, color = ?, is_active = ? WHERE business_id = ? AND id = ?")
        .bind(&params.name)
        .bind(&params.description)
        .bind(params.duration_minutes)
        .bind(&params.price)
        .bind(&params.color)
        .bind(params.is_active)
        .bind(params.business_id)
        .bind(params.id)
        .execute(exec)
        .await
}
