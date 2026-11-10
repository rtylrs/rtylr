//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `FetchCalendarAppointments` (Go `FetchCalendarAppointmentsParams`).
pub struct FetchCalendarAppointmentsParams {
    pub business_id: i64,
    pub start_time: chrono::NaiveDate,
    pub start_time_2: chrono::NaiveDate,
}

/// `FetchCalendarAppointments`: `many`.
pub async fn fetch_calendar_appointments<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &FetchCalendarAppointmentsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, service_id, customer_name, start_time, end_time, status, business_id\nFROM booking_appointments\nWHERE business_id = ?\n  AND start_time >= ?\n  AND start_time < ?\nORDER BY start_time ASC\nLIMIT 100")
        .bind(params.business_id)
        .bind(params.start_time)
        .bind(params.start_time_2)
        .execute(exec)
        .await
}
