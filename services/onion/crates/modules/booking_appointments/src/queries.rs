//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `CreateAppointment` (Go `CreateAppointmentParams`).
pub struct CreateAppointmentParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub service_id: Option<i64>,
    pub resource_id: Option<i64>,
    pub staff_id: Option<i64>,
    pub customer_name: Option<String>,
    pub customer_email: Option<String>,
    pub customer_phone: Option<String>,
    pub start_time: chrono::NaiveDate,
    pub end_time: chrono::NaiveDate,
    pub status: String,
    pub notes: Option<String>,
}

/// `CreateAppointment`: `execresult`.
pub async fn create_appointment<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateAppointmentParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO booking_appointments (owner_id, business_id, service_id, resource_id, staff_id, customer_name, customer_email, customer_phone, start_time, end_time, status, notes)\nVALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(params.service_id)
        .bind(params.resource_id)
        .bind(params.staff_id)
        .bind(&params.customer_name)
        .bind(&params.customer_email)
        .bind(&params.customer_phone)
        .bind(params.start_time)
        .bind(params.end_time)
        .bind(&params.status)
        .bind(&params.notes)
        .execute(exec)
        .await
}

/// Params for `DeleteAppointment` (Go `DeleteAppointmentParams`).
pub struct DeleteAppointmentParams {
    pub business_id: i64,
    pub id: i64,
}

/// `DeleteAppointment`: `execresult`.
pub async fn delete_appointment<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DeleteAppointmentParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("DELETE FROM booking_appointments WHERE business_id = ? AND id = ?")
        .bind(params.business_id)
        .bind(params.id)
        .execute(exec)
        .await
}

/// Params for `EnsureScheduleLock` (Go `EnsureScheduleLockParams`).
pub struct EnsureScheduleLockParams {
    pub business_id: i64,
}

/// `EnsureScheduleLock`: `exec`.
pub async fn ensure_schedule_lock<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &EnsureScheduleLockParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO booking_schedule_locks (business_id) VALUES (?) ON DUPLICATE KEY UPDATE business_id = business_id")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `GetAppointment` (Go `GetAppointmentParams`).
pub struct GetAppointmentParams {
    pub business_id: i64,
    pub id: i64,
}

/// `GetAppointment`: `one`.
pub async fn get_appointment<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetAppointmentParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, service_id, resource_id, staff_id,\n       customer_name, customer_email, customer_phone, start_time, end_time,\n       status, notes, created_at, updated_at\nFROM booking_appointments\nWHERE business_id = ? AND id = ?")
        .bind(params.business_id)
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetAppointmentStatusForUpdate` (Go `GetAppointmentStatusForUpdateParams`).
pub struct GetAppointmentStatusForUpdateParams {
    pub business_id: i64,
    pub id: i64,
}

/// `GetAppointmentStatusForUpdate`: `one`.
pub async fn get_appointment_status_for_update<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetAppointmentStatusForUpdateParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query(
        "SELECT status FROM booking_appointments WHERE business_id = ? AND id = ? FOR UPDATE",
    )
    .bind(params.business_id)
    .bind(params.id)
    .fetch_optional(exec)
    .await
}

/// Params for `GetBookingResourceForAppointment` (Go `GetBookingResourceForAppointmentParams`).
pub struct GetBookingResourceForAppointmentParams {
    pub id: i64,
    pub business_id: i64,
}

/// `GetBookingResourceForAppointment`: `one`.
pub async fn get_booking_resource_for_appointment<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetBookingResourceForAppointmentParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query(
        "SELECT id FROM booking_resources WHERE id = ? AND business_id = ? AND is_active = TRUE",
    )
    .bind(params.id)
    .bind(params.business_id)
    .fetch_optional(exec)
    .await
}

/// Params for `GetBookingServiceForAppointment` (Go `GetBookingServiceForAppointmentParams`).
pub struct GetBookingServiceForAppointmentParams {
    pub id: i64,
    pub business_id: i64,
}

/// `GetBookingServiceForAppointment`: `one`.
pub async fn get_booking_service_for_appointment<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetBookingServiceForAppointmentParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query(
        "SELECT id FROM booking_services WHERE id = ? AND business_id = ? AND is_active = TRUE",
    )
    .bind(params.id)
    .bind(params.business_id)
    .fetch_optional(exec)
    .await
}

/// Params for `ListAppointments` (Go `ListAppointmentsParams`).
pub struct ListAppointmentsParams {
    pub business_id: i64,
}

/// `ListAppointments`: `many`.
pub async fn list_appointments<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListAppointmentsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, service_id, resource_id, staff_id, customer_name, customer_email, customer_phone, start_time, end_time, status, notes, created_at, updated_at\nFROM booking_appointments WHERE business_id = ? ORDER BY start_time DESC, id DESC LIMIT 500")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `ListAppointmentsByDateRange` (Go `ListAppointmentsByDateRangeParams`).
pub struct ListAppointmentsByDateRangeParams {
    pub business_id: i64,
    pub start_time: chrono::NaiveDate,
    pub start_time_2: chrono::NaiveDate,
}

/// `ListAppointmentsByDateRange`: `many`.
pub async fn list_appointments_by_date_range<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListAppointmentsByDateRangeParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, service_id, resource_id, staff_id, customer_name, customer_email, customer_phone, start_time, end_time, status, notes, created_at, updated_at\nFROM booking_appointments WHERE business_id = ? AND start_time >= ? AND start_time < ? ORDER BY start_time ASC, id ASC LIMIT 500")
        .bind(params.business_id)
        .bind(params.start_time)
        .bind(params.start_time_2)
        .execute(exec)
        .await
}

/// Params for `LockAppointmentOverlaps` (Go `LockAppointmentOverlapsParams`).
pub struct LockAppointmentOverlapsParams {
    pub business_id: i64,
    pub start_time: chrono::NaiveDate,
    pub end_time: chrono::NaiveDate,
    pub column4: String,
    pub id: i64,
    pub column6: String,
    pub resource_id: Option<i64>,
    pub column8: String,
    pub staff_id: Option<i64>,
}

/// `LockAppointmentOverlaps`: `many`.
pub async fn lock_appointment_overlaps<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &LockAppointmentOverlapsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id\nFROM booking_appointments\nWHERE business_id = ?\n  AND status NOT IN ('cancelled', 'no_show')\n  AND start_time < ?\n  AND end_time > ?\n  AND (? = 0 OR id <> ?)\n  AND ((? > 0 AND resource_id = ?) OR (? > 0 AND staff_id = ?))\nLIMIT 1\nFOR UPDATE")
        .bind(params.business_id)
        .bind(params.start_time)
        .bind(params.end_time)
        .bind(&params.column4)
        .bind(params.id)
        .bind(&params.column6)
        .bind(params.resource_id)
        .bind(&params.column8)
        .bind(params.staff_id)
        .execute(exec)
        .await
}

/// Params for `LockBusinessForBooking` (Go `LockBusinessForBookingParams`).
pub struct LockBusinessForBookingParams {
    pub business_id: i64,
}

/// `LockBusinessForBooking`: `one`.
pub async fn lock_business_for_booking<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &LockBusinessForBookingParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT business_id FROM booking_schedule_locks WHERE business_id = ? FOR UPDATE")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `UpdateAppointment` (Go `UpdateAppointmentParams`).
pub struct UpdateAppointmentParams {
    pub service_id: Option<i64>,
    pub resource_id: Option<i64>,
    pub staff_id: Option<i64>,
    pub customer_name: Option<String>,
    pub customer_email: Option<String>,
    pub customer_phone: Option<String>,
    pub start_time: chrono::NaiveDate,
    pub end_time: chrono::NaiveDate,
    pub status: String,
    pub notes: Option<String>,
    pub business_id: i64,
    pub id: i64,
}

/// `UpdateAppointment`: `execresult`.
pub async fn update_appointment<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateAppointmentParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE booking_appointments SET service_id = ?, resource_id = ?, staff_id = ?, customer_name = ?, customer_email = ?, customer_phone = ?, start_time = ?, end_time = ?, status = ?, notes = ?\nWHERE business_id = ? AND id = ?")
        .bind(params.service_id)
        .bind(params.resource_id)
        .bind(params.staff_id)
        .bind(&params.customer_name)
        .bind(&params.customer_email)
        .bind(&params.customer_phone)
        .bind(params.start_time)
        .bind(params.end_time)
        .bind(&params.status)
        .bind(&params.notes)
        .bind(params.business_id)
        .bind(params.id)
        .execute(exec)
        .await
}
