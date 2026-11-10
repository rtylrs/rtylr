//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `CountAllAttendance` (Go `CountAllAttendanceParams`).
pub struct CountAllAttendanceParams {
    pub business_id: i64,
    pub date: chrono::NaiveDate,
    pub date_2: chrono::NaiveDate,
}

/// `CountAllAttendance`: `one`.
pub async fn count_all_attendance<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountAllAttendanceParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) AS total FROM hr_attendance_records\nWHERE business_id = ? AND date >= ? AND date <= ?")
        .bind(params.business_id)
        .bind(params.date)
        .bind(params.date_2)
        .fetch_optional(exec)
        .await
}

/// Params for `CountAttendance` (Go `CountAttendanceParams`).
pub struct CountAttendanceParams {
    pub business_id: i64,
    pub employee_id: i64,
    pub date: chrono::NaiveDate,
    pub date_2: chrono::NaiveDate,
}

/// `CountAttendance`: `one`.
pub async fn count_attendance<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountAttendanceParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) AS total FROM hr_attendance_records\nWHERE business_id = ? AND employee_id = ? AND date >= ? AND date <= ?")
        .bind(params.business_id)
        .bind(params.employee_id)
        .bind(params.date)
        .bind(params.date_2)
        .fetch_optional(exec)
        .await
}

/// Params for `CreateAttendanceRecord` (Go `CreateAttendanceRecordParams`).
pub struct CreateAttendanceRecordParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub employee_id: i64,
    pub location_id: i64,
    pub date: chrono::NaiveDate,
    pub clock_in_at: chrono::NaiveDate,
    pub clock_in_method: String,
    pub clock_in_lat: Option<f64>,
    pub clock_in_lng: Option<f64>,
    pub is_late: bool,
    pub late_minutes: i32,
    pub status: String,
}

/// `CreateAttendanceRecord`: `execresult`.
pub async fn create_attendance_record<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateAttendanceRecordParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO hr_attendance_records (\n    owner_id, business_id, employee_id, location_id, date, clock_in_at,\n    clock_in_method, clock_in_lat, clock_in_lng, is_late, late_minutes, status\n) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(params.employee_id)
        .bind(params.location_id)
        .bind(params.date)
        .bind(params.clock_in_at)
        .bind(&params.clock_in_method)
        .bind(params.clock_in_lat)
        .bind(params.clock_in_lng)
        .bind(params.is_late)
        .bind(params.late_minutes)
        .bind(&params.status)
        .execute(exec)
        .await
}

/// Params for `GetActiveClockIn` (Go `GetActiveClockInParams`).
pub struct GetActiveClockInParams {
    pub business_id: i64,
    pub employee_id: i64,
}

/// `GetActiveClockIn`: `one`.
pub async fn get_active_clock_in<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetActiveClockInParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, employee_id, location_id, date, clock_in_at, clock_out_at,\n       duration_minutes, clock_in_method, clock_in_lat, clock_in_lng,\n       is_late, late_minutes, status, created_at, updated_at\nFROM hr_attendance_records\nWHERE business_id = ? AND employee_id = ? AND clock_out_at IS NULL\nORDER BY clock_in_at DESC\nLIMIT 1")
        .bind(params.business_id)
        .bind(params.employee_id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetAttendanceConfig` (Go `GetAttendanceConfigParams`).
pub struct GetAttendanceConfigParams {
    pub business_id: i64,
}

/// `GetAttendanceConfig`: `one`.
pub async fn get_attendance_config<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetAttendanceConfigParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, app_enabled, qr_enabled, gps_enabled, pin_enabled,\n       facial_enabled, geofence_radius_m, default_location_lat, default_location_lng\nFROM hr_attendance_config\nWHERE business_id = ?")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetDailyAttendance` (Go `GetDailyAttendanceParams`).
pub struct GetDailyAttendanceParams {
    pub business_id: i64,
    pub location_id: i64,
    pub date: chrono::NaiveDate,
    pub limit: i32,
    pub offset: i32,
}

/// `GetDailyAttendance`: `many`.
pub async fn get_daily_attendance<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetDailyAttendanceParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, employee_id, location_id, date, clock_in_at, clock_out_at,\n       duration_minutes, clock_in_method, clock_in_lat, clock_in_lng,\n       is_late, late_minutes, status, created_at, updated_at\nFROM hr_attendance_records\nWHERE business_id = ? AND location_id = ? AND date = ?\nORDER BY clock_in_at ASC\nLIMIT ? OFFSET ?")
        .bind(params.business_id)
        .bind(params.location_id)
        .bind(params.date)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `GetEmployeePINHash` (Go `GetEmployeePINHashParams`).
pub struct GetEmployeePINHashParams {
    pub id: i64,
    pub business_id: i64,
}

/// `GetEmployeePINHash`: `one`.
pub async fn get_employee_pin_hash<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetEmployeePINHashParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT clock_in_pin_hash\nFROM hr_employees\nWHERE id = ? AND business_id = ?")
        .bind(params.id)
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetLocationCoordinates` (Go `GetLocationCoordinatesParams`).
pub struct GetLocationCoordinatesParams {
    pub id: i64,
    pub business_id: i64,
}

/// `GetLocationCoordinates`: `one`.
pub async fn get_location_coordinates<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetLocationCoordinatesParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT latitude, longitude\nFROM hr_locations\nWHERE id = ? AND business_id = ?")
        .bind(params.id)
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `ListAllAttendance` (Go `ListAllAttendanceParams`).
pub struct ListAllAttendanceParams {
    pub business_id: i64,
    pub date: chrono::NaiveDate,
    pub date_2: chrono::NaiveDate,
    pub limit: i32,
    pub offset: i32,
}

/// `ListAllAttendance`: `many`.
pub async fn list_all_attendance<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListAllAttendanceParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, employee_id, location_id, date, clock_in_at, clock_out_at,\n       duration_minutes, clock_in_method, clock_in_lat, clock_in_lng,\n       is_late, late_minutes, status, created_at, updated_at\nFROM hr_attendance_records\nWHERE business_id = ? AND date >= ? AND date <= ?\nORDER BY date DESC, clock_in_at DESC\nLIMIT ? OFFSET ?")
        .bind(params.business_id)
        .bind(params.date)
        .bind(params.date_2)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `ListAttendance` (Go `ListAttendanceParams`).
pub struct ListAttendanceParams {
    pub business_id: i64,
    pub employee_id: i64,
    pub date: chrono::NaiveDate,
    pub date_2: chrono::NaiveDate,
    pub limit: i32,
    pub offset: i32,
}

/// `ListAttendance`: `many`.
pub async fn list_attendance<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListAttendanceParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, employee_id, location_id, date, clock_in_at, clock_out_at,\n       duration_minutes, clock_in_method, clock_in_lat, clock_in_lng,\n       is_late, late_minutes, status, created_at, updated_at\nFROM hr_attendance_records\nWHERE business_id = ? AND employee_id = ? AND date >= ? AND date <= ?\nORDER BY date DESC, clock_in_at DESC\nLIMIT ? OFFSET ?")
        .bind(params.business_id)
        .bind(params.employee_id)
        .bind(params.date)
        .bind(params.date_2)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `SetEmployeePINHash` (Go `SetEmployeePINHashParams`).
pub struct SetEmployeePINHashParams {
    pub clock_in_pin_hash: Option<String>,
    pub id: i64,
    pub business_id: i64,
}

/// `SetEmployeePINHash`: `execresult`.
pub async fn set_employee_pin_hash<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &SetEmployeePINHashParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE hr_employees\nSET clock_in_pin_hash = ?, updated_at = NOW()\nWHERE id = ? AND business_id = ?")
        .bind(&params.clock_in_pin_hash)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `UpdateAttendanceRecord` (Go `UpdateAttendanceRecordParams`).
pub struct UpdateAttendanceRecordParams {
    pub clock_out_at: Option<chrono::DateTime<chrono::Utc>>,
    pub duration_minutes: Option<i32>,
    pub status: String,
    pub id: i64,
    pub business_id: i64,
}

/// `UpdateAttendanceRecord`: `exec`.
pub async fn update_attendance_record<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateAttendanceRecordParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE hr_attendance_records\nSET clock_out_at = ?, duration_minutes = ?, status = ?, updated_at = NOW()\nWHERE id = ? AND business_id = ?")
        .bind(params.clock_out_at)
        .bind(params.duration_minutes)
        .bind(&params.status)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}
