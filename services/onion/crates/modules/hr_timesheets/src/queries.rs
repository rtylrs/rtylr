//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `GetTimesheets` (Go `GetTimesheetsParams`).
pub struct GetTimesheetsParams {
    pub business_id: i64,
}

/// `GetTimesheets`: `many`.
pub async fn get_timesheets<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetTimesheetsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT s.staff_id,\n       CONCAT(st.first_name, ' ', st.last_name) AS staff_name,\n       DATE(s.date - INTERVAL WEEKDAY(s.date) DAY) AS week_start,\n       COALESCE(SUM(s.hours_worked), 0) AS total_hours,\n       COALESCE(SUM(s.hours_worked) - SUM(s.overtime_hours), 0) AS regular_hours,\n       COALESCE(SUM(s.overtime_hours), 0) AS overtime_hours\nFROM hr_shifts s\nJOIN hr_staff st ON st.id = s.staff_id AND st.business_id = s.business_id\nWHERE s.business_id = ? AND s.status = 'completed'\nGROUP BY s.staff_id, staff_name, week_start\nORDER BY week_start DESC, staff_name ASC")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `GetTimesheetsForWeek` (Go `GetTimesheetsForWeekParams`).
pub struct GetTimesheetsForWeekParams {
    pub business_id: i64,
    pub date: Option<chrono::DateTime<chrono::Utc>>,
    pub date_2: Option<chrono::DateTime<chrono::Utc>>,
}

/// `GetTimesheetsForWeek`: `many`.
pub async fn get_timesheets_for_week<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetTimesheetsForWeekParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT s.staff_id,\n       CONCAT(st.first_name, ' ', st.last_name) AS staff_name,\n       DATE(s.date - INTERVAL WEEKDAY(s.date) DAY) AS week_start,\n       COALESCE(SUM(s.hours_worked), 0) AS total_hours,\n       COALESCE(SUM(s.hours_worked) - SUM(s.overtime_hours), 0) AS regular_hours,\n       COALESCE(SUM(s.overtime_hours), 0) AS overtime_hours\nFROM hr_shifts s\nJOIN hr_staff st ON st.id = s.staff_id AND st.business_id = s.business_id\nWHERE s.business_id = ?\n  AND s.status = 'completed'\n  AND s.date >= ?\n  AND s.date < ?\nGROUP BY s.staff_id, staff_name, week_start\nORDER BY week_start DESC, staff_name ASC")
        .bind(params.business_id)
        .bind(params.date)
        .bind(params.date_2)
        .execute(exec)
        .await
}
