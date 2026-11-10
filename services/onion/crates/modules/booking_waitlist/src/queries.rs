//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `CreateWaitlistEntry` (Go `CreateWaitlistEntryParams`).
pub struct CreateWaitlistEntryParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub service_id: Option<i64>,
    pub customer_name: Option<String>,
    pub customer_email: Option<String>,
    pub customer_phone: Option<String>,
    pub preferred_date: Option<chrono::DateTime<chrono::Utc>>,
    pub preferred_time: Option<String>,
    pub status: String,
}

/// `CreateWaitlistEntry`: `execresult`.
pub async fn create_waitlist_entry<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateWaitlistEntryParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO booking_waitlist (owner_id, business_id, service_id, customer_name, customer_email, customer_phone, preferred_date, preferred_time, status)\nVALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(params.service_id)
        .bind(&params.customer_name)
        .bind(&params.customer_email)
        .bind(&params.customer_phone)
        .bind(params.preferred_date)
        .bind(&params.preferred_time)
        .bind(&params.status)
        .execute(exec)
        .await
}

/// Params for `GetBookingServiceForOwner` (Go `GetBookingServiceForOwnerParams`).
pub struct GetBookingServiceForOwnerParams {
    pub id: i64,
    pub business_id: i64,
}

/// `GetBookingServiceForOwner`: `one`.
pub async fn get_booking_service_for_owner<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetBookingServiceForOwnerParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query(
        "SELECT id, owner_id, business_id FROM booking_services WHERE id = ? AND business_id = ?",
    )
    .bind(params.id)
    .bind(params.business_id)
    .fetch_optional(exec)
    .await
}

/// Params for `GetWaitlistEntry` (Go `GetWaitlistEntryParams`).
pub struct GetWaitlistEntryParams {
    pub id: i64,
    pub business_id: i64,
}

/// `GetWaitlistEntry`: `one`.
pub async fn get_waitlist_entry<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetWaitlistEntryParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id FROM booking_waitlist WHERE id = ? AND business_id = ?")
        .bind(params.id)
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `ListWaitlist` (Go `ListWaitlistParams`).
pub struct ListWaitlistParams {
    pub business_id: i64,
}

/// `ListWaitlist`: `many`.
pub async fn list_waitlist<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListWaitlistParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, service_id, customer_name, customer_email, customer_phone, preferred_date, preferred_time, status, created_at\nFROM booking_waitlist WHERE business_id = ? ORDER BY created_at DESC LIMIT 100")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `UpdateWaitlistEntry` (Go `UpdateWaitlistEntryParams`).
pub struct UpdateWaitlistEntryParams {
    pub service_id: Option<i64>,
    pub customer_name: Option<String>,
    pub customer_email: Option<String>,
    pub customer_phone: Option<String>,
    pub preferred_date: Option<chrono::DateTime<chrono::Utc>>,
    pub preferred_time: Option<String>,
    pub status: String,
    pub business_id: i64,
    pub id: i64,
}

/// `UpdateWaitlistEntry`: `execresult`.
pub async fn update_waitlist_entry<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateWaitlistEntryParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE booking_waitlist SET service_id = ?, customer_name = ?, customer_email = ?, customer_phone = ?, preferred_date = ?, preferred_time = ?, status = ?\nWHERE business_id = ? AND id = ?")
        .bind(params.service_id)
        .bind(&params.customer_name)
        .bind(&params.customer_email)
        .bind(&params.customer_phone)
        .bind(params.preferred_date)
        .bind(&params.preferred_time)
        .bind(&params.status)
        .bind(params.business_id)
        .bind(params.id)
        .execute(exec)
        .await
}
