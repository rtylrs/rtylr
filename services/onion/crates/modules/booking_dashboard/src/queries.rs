//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `CountAppointmentsThisMonth` (Go `CountAppointmentsThisMonthParams`).
pub struct CountAppointmentsThisMonthParams {
    pub business_id: i64,
    pub start_time: chrono::NaiveDate,
}

/// `CountAppointmentsThisMonth`: `one`.
pub async fn count_appointments_this_month<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountAppointmentsThisMonthParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) AS cnt\nFROM booking_appointments\nWHERE business_id = ? AND start_time >= ?")
        .bind(params.business_id)
        .bind(params.start_time)
        .fetch_optional(exec)
        .await
}

/// Params for `CountAppointmentsToday` (Go `CountAppointmentsTodayParams`).
pub struct CountAppointmentsTodayParams {
    pub business_id: i64,
    pub start_time: chrono::NaiveDate,
    pub start_time_2: chrono::NaiveDate,
}

/// `CountAppointmentsToday`: `one`.
pub async fn count_appointments_today<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountAppointmentsTodayParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) AS cnt\nFROM booking_appointments\nWHERE business_id = ? AND start_time >= ? AND start_time < ?")
        .bind(params.business_id)
        .bind(params.start_time)
        .bind(params.start_time_2)
        .fetch_optional(exec)
        .await
}

/// Params for `CountCancelledAppointmentsThisMonth` (Go `CountCancelledAppointmentsThisMonthParams`).
pub struct CountCancelledAppointmentsThisMonthParams {
    pub business_id: i64,
    pub start_time: chrono::NaiveDate,
}

/// `CountCancelledAppointmentsThisMonth`: `one`.
pub async fn count_cancelled_appointments_this_month<
    'e,
    E: sqlx::Executor<'e, Database = MySql>,
>(
    exec: E,
    params: &CountCancelledAppointmentsThisMonthParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) AS cnt\nFROM booking_appointments\nWHERE business_id = ? AND start_time >= ? AND status = 'cancelled'")
        .bind(params.business_id)
        .bind(params.start_time)
        .fetch_optional(exec)
        .await
}

/// Params for `CountUpcomingAppointments` (Go `CountUpcomingAppointmentsParams`).
pub struct CountUpcomingAppointmentsParams {
    pub business_id: i64,
    pub start_time: chrono::NaiveDate,
}

/// `CountUpcomingAppointments`: `one`.
pub async fn count_upcoming_appointments<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountUpcomingAppointmentsParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) AS cnt\nFROM booking_appointments\nWHERE business_id = ? AND start_time >= ? AND status IN ('pending', 'confirmed')")
        .bind(params.business_id)
        .bind(params.start_time)
        .fetch_optional(exec)
        .await
}

/// Params for `SumMonthRevenue` (Go `SumMonthRevenueParams`).
pub struct SumMonthRevenueParams {
    pub business_id: i64,
    pub start_time: chrono::NaiveDate,
}

/// `SumMonthRevenue`: `one`.
pub async fn sum_month_revenue<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &SumMonthRevenueParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COALESCE(SUM(bs.price), 0) AS revenue\nFROM booking_appointments ba\nJOIN booking_services bs ON ba.service_id = bs.id\nWHERE ba.business_id = ? AND ba.start_time >= ?\n  AND ba.status IN ('completed', 'confirmed')")
        .bind(params.business_id)
        .bind(params.start_time)
        .fetch_optional(exec)
        .await
}
