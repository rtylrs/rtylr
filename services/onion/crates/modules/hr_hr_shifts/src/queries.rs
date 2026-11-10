//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `CountShifts` (Go `CountShiftsParams`).
pub struct CountShiftsParams {
    pub business_id: i64,
    pub staff_id: Option<i64>,
    pub date: Option<chrono::DateTime<chrono::Utc>>,
    pub location_id: Option<i64>,
    pub kind: String,
}

/// `CountShifts`: `one`.
pub async fn count_shifts<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountShiftsParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) AS total\nFROM hr_shifts\nWHERE business_id = ?\n  AND (? IS NULL OR staff_id = ?)\n  AND (? IS NULL OR date = ?)\n  AND (? IS NULL OR location_id = ?)\n  AND (? IS NULL\n       OR (? = 'template' AND staff_id IS NULL)\n       OR (? = 'attendance' AND staff_id IS NOT NULL))")
        .bind(params.business_id)
        .bind(params.staff_id)
        .bind(params.staff_id)
        .bind(params.date)
        .bind(params.date)
        .bind(params.location_id)
        .bind(params.location_id)
        .bind(&params.kind)
        .bind(&params.kind)
        .bind(&params.kind)
        .fetch_optional(exec)
        .await
}

/// Params for `CountStaffAvailability` (Go `CountStaffAvailabilityParams`).
pub struct CountStaffAvailabilityParams {
    pub business_id: i64,
    pub staff_id: Option<i64>,
    pub day_of_week: Option<i16>,
    pub date_from: Option<chrono::DateTime<chrono::Utc>>,
    pub date_to: Option<chrono::DateTime<chrono::Utc>>,
}

/// `CountStaffAvailability`: `one`.
pub async fn count_staff_availability<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountStaffAvailabilityParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*)\nFROM hr_staff_availability\nWHERE business_id = ?\n  AND (? IS NULL OR staff_id = ?)\n  AND (? IS NULL OR day_of_week = ?)\n  AND (? IS NULL OR specific_date IS NULL OR specific_date >= ?)\n  AND (? IS NULL OR specific_date IS NULL OR specific_date <= ?)")
        .bind(params.business_id)
        .bind(params.staff_id)
        .bind(params.staff_id)
        .bind(params.day_of_week)
        .bind(params.day_of_week)
        .bind(params.date_from)
        .bind(params.date_from)
        .bind(params.date_to)
        .bind(params.date_to)
        .fetch_optional(exec)
        .await
}

/// Params for `CreateShift` (Go `CreateShiftParams`).
pub struct CreateShiftParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub staff_id: Option<i64>,
    pub department_id: Option<i64>,
    pub date: Option<chrono::DateTime<chrono::Utc>>,
    pub start_time: String,
    pub end_time: String,
    pub break_duration_minutes: Option<i32>,
    pub status: String,
    pub notes: Option<String>,
    pub location_id: Option<i64>,
    pub name: Option<String>,
    pub color: Option<String>,
}

/// `CreateShift`: `execresult`.
pub async fn create_shift<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateShiftParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO hr_shifts (\n    owner_id, business_id, staff_id, department_id, date, start_time, end_time,\n    break_duration_minutes, status, notes, location_id, name, color\n) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(params.staff_id)
        .bind(params.department_id)
        .bind(params.date)
        .bind(&params.start_time)
        .bind(&params.end_time)
        .bind(params.break_duration_minutes)
        .bind(&params.status)
        .bind(&params.notes)
        .bind(params.location_id)
        .bind(&params.name)
        .bind(&params.color)
        .execute(exec)
        .await
}

/// Params for `CreateShiftAssignment` (Go `CreateShiftAssignmentParams`).
pub struct CreateShiftAssignmentParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub employee_id: i64,
    pub shift_id: i64,
    pub date: chrono::NaiveDate,
}

/// `CreateShiftAssignment`: `execresult`.
pub async fn create_shift_assignment<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateShiftAssignmentParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO hr_shift_assignments (owner_id, business_id, employee_id, shift_id, date, status)\nVALUES (?, ?, ?, ?, ?, 'scheduled')")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(params.employee_id)
        .bind(params.shift_id)
        .bind(params.date)
        .execute(exec)
        .await
}

/// Params for `CreateStaffAvailability` (Go `CreateStaffAvailabilityParams`).
pub struct CreateStaffAvailabilityParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub staff_id: i64,
    pub day_of_week: Option<i16>,
    pub specific_date: Option<chrono::DateTime<chrono::Utc>>,
    pub start_time: String,
    pub end_time: String,
    pub is_available: bool,
}

/// `CreateStaffAvailability`: `execresult`.
pub async fn create_staff_availability<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateStaffAvailabilityParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO hr_staff_availability (owner_id, business_id, staff_id, day_of_week, specific_date, start_time, end_time, is_available)\nVALUES (?, ?, ?, ?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(params.staff_id)
        .bind(params.day_of_week)
        .bind(params.specific_date)
        .bind(&params.start_time)
        .bind(&params.end_time)
        .bind(params.is_available)
        .execute(exec)
        .await
}

/// Params for `DeleteShift` (Go `DeleteShiftParams`).
pub struct DeleteShiftParams {
    pub id: i64,
    pub business_id: i64,
}

/// `DeleteShift`: `exec`.
pub async fn delete_shift<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DeleteShiftParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("DELETE FROM hr_shifts WHERE id = ? AND business_id = ?")
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `DeleteStaffAvailability` (Go `DeleteStaffAvailabilityParams`).
pub struct DeleteStaffAvailabilityParams {
    pub business_id: i64,
    pub id: i64,
}

/// `DeleteStaffAvailability`: `execresult`.
pub async fn delete_staff_availability<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DeleteStaffAvailabilityParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("DELETE FROM hr_staff_availability\nWHERE business_id = ? AND id = ?")
        .bind(params.business_id)
        .bind(params.id)
        .execute(exec)
        .await
}

/// Params for `GetShift` (Go `GetShiftParams`).
pub struct GetShiftParams {
    pub business_id: i64,
    pub id: i64,
}

/// `GetShift`: `one`.
pub async fn get_shift<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetShiftParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, staff_id, department_id, date, start_time, end_time,\n       break_duration_minutes, status, clock_in_at, clock_out_at,\n       hours_worked, overtime_hours, notes, location_id, name, color,\n       created_at, updated_at\nFROM hr_shifts\nWHERE business_id = ? AND id = ?")
        .bind(params.business_id)
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetShiftAssignment` (Go `GetShiftAssignmentParams`).
pub struct GetShiftAssignmentParams {
    pub business_id: i64,
    pub id: i64,
}

/// `GetShiftAssignment`: `one`.
pub async fn get_shift_assignment<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetShiftAssignmentParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, employee_id, shift_id, date, status, published, created_at, updated_at\nFROM hr_shift_assignments\nWHERE business_id = ? AND id = ?")
        .bind(params.business_id)
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetShiftForUpdate` (Go `GetShiftForUpdateParams`).
pub struct GetShiftForUpdateParams {
    pub business_id: i64,
    pub id: i64,
}

/// `GetShiftForUpdate`: `one`.
pub async fn get_shift_for_update<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetShiftForUpdateParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, staff_id, department_id, date, start_time, end_time,\n       break_duration_minutes, status, clock_in_at, clock_out_at,\n       hours_worked, overtime_hours, notes, location_id, name, color,\n       created_at, updated_at\nFROM hr_shifts\nWHERE business_id = ? AND id = ?\nFOR UPDATE")
        .bind(params.business_id)
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetStaffAvailability` (Go `GetStaffAvailabilityParams`).
pub struct GetStaffAvailabilityParams {
    pub business_id: i64,
    pub id: i64,
}

/// `GetStaffAvailability`: `one`.
pub async fn get_staff_availability<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetStaffAvailabilityParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, staff_id, day_of_week, specific_date, start_time, end_time, is_available, created_at, updated_at, day_key, date_key\nFROM hr_staff_availability\nWHERE business_id = ? AND id = ?")
        .bind(params.business_id)
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetStaffAvailabilityByNaturalKey` (Go `GetStaffAvailabilityByNaturalKeyParams`).
pub struct GetStaffAvailabilityByNaturalKeyParams {
    pub business_id: i64,
    pub staff_id: i64,
    pub day_of_week: String,
    pub specific_date: String,
    pub start_time: String,
}

/// `GetStaffAvailabilityByNaturalKey`: `one`.
pub async fn get_staff_availability_by_natural_key<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetStaffAvailabilityByNaturalKeyParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, staff_id, day_of_week, specific_date, start_time, end_time, is_available, created_at, updated_at, day_key, date_key\nFROM hr_staff_availability\nWHERE business_id = ?\n  AND staff_id = ?\n  AND IFNULL(day_of_week, -1) = IFNULL(?, -1)\n  AND IFNULL(specific_date, '1000-01-01') = IFNULL(?, '1000-01-01')\n  AND start_time = ?")
        .bind(params.business_id)
        .bind(params.staff_id)
        .bind(&params.day_of_week)
        .bind(&params.specific_date)
        .bind(&params.start_time)
        .fetch_optional(exec)
        .await
}

/// Params for `ListEmployeeAssignmentsForDate` (Go `ListEmployeeAssignmentsForDateParams`).
pub struct ListEmployeeAssignmentsForDateParams {
    pub business_id: i64,
    pub employee_id: i64,
    pub date: chrono::NaiveDate,
    pub limit: i32,
    pub offset: i32,
}

/// `ListEmployeeAssignmentsForDate`: `many`.
pub async fn list_employee_assignments_for_date<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListEmployeeAssignmentsForDateParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, employee_id, shift_id, date, status, published, created_at, updated_at\nFROM hr_shift_assignments\nWHERE business_id = ? AND employee_id = ? AND date = ?\nLIMIT ? OFFSET ?")
        .bind(params.business_id)
        .bind(params.employee_id)
        .bind(params.date)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `ListRosterAssignments` (Go `ListRosterAssignmentsParams`).
pub struct ListRosterAssignmentsParams {
    pub business_id: i64,
    pub date: chrono::NaiveDate,
    pub date_2: chrono::NaiveDate,
    pub limit: i32,
    pub offset: i32,
}

/// `ListRosterAssignments`: `many`.
pub async fn list_roster_assignments<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListRosterAssignmentsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, employee_id, shift_id, date, status, published, created_at, updated_at\nFROM hr_shift_assignments\nWHERE business_id = ? AND date >= ? AND date < ?\nORDER BY date ASC, employee_id ASC\nLIMIT ? OFFSET ?")
        .bind(params.business_id)
        .bind(params.date)
        .bind(params.date_2)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `ListShifts` (Go `ListShiftsParams`).
pub struct ListShiftsParams {
    pub business_id: i64,
    pub staff_id: Option<i64>,
    pub date: Option<chrono::DateTime<chrono::Utc>>,
    pub location_id: Option<i64>,
    pub kind: String,
    pub limit: i32,
    pub offset: i32,
}

/// `ListShifts`: `many`.
pub async fn list_shifts<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListShiftsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, staff_id, department_id, date, start_time, end_time,\n       break_duration_minutes, status, clock_in_at, clock_out_at,\n       hours_worked, overtime_hours, notes, location_id, name, color,\n       created_at, updated_at\nFROM hr_shifts\nWHERE business_id = ?\n  AND (? IS NULL OR staff_id = ?)\n  AND (? IS NULL OR date = ?)\n  AND (? IS NULL OR location_id = ?)\n  AND (? IS NULL\n       OR (? = 'template' AND staff_id IS NULL)\n       OR (? = 'attendance' AND staff_id IS NOT NULL))\nORDER BY date DESC, start_time ASC, name ASC\nLIMIT ? OFFSET ?")
        .bind(params.business_id)
        .bind(params.staff_id)
        .bind(params.staff_id)
        .bind(params.date)
        .bind(params.date)
        .bind(params.location_id)
        .bind(params.location_id)
        .bind(&params.kind)
        .bind(&params.kind)
        .bind(&params.kind)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `ListShiftsByLocation` (Go `ListShiftsByLocationParams`).
pub struct ListShiftsByLocationParams {
    pub business_id: i64,
    pub location_id: Option<i64>,
    pub limit: i32,
    pub offset: i32,
}

/// `ListShiftsByLocation`: `many`.
pub async fn list_shifts_by_location<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListShiftsByLocationParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, staff_id, department_id, date, start_time, end_time,\n       break_duration_minutes, status, clock_in_at, clock_out_at,\n       hours_worked, overtime_hours, notes, location_id, name, color,\n       created_at, updated_at\nFROM hr_shifts\nWHERE business_id = ? AND location_id = ?\nORDER BY start_time ASC\nLIMIT ? OFFSET ?")
        .bind(params.business_id)
        .bind(params.location_id)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `ListStaffAvailability` (Go `ListStaffAvailabilityParams`).
pub struct ListStaffAvailabilityParams {
    pub business_id: i64,
    pub staff_id: Option<i64>,
    pub day_of_week: Option<i16>,
    pub date_from: Option<chrono::DateTime<chrono::Utc>>,
    pub date_to: Option<chrono::DateTime<chrono::Utc>>,
    pub limit: i32,
    pub offset: i32,
}

/// `ListStaffAvailability`: `many`.
pub async fn list_staff_availability<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListStaffAvailabilityParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, staff_id, day_of_week, specific_date, start_time, end_time, is_available, created_at, updated_at, day_key, date_key\nFROM hr_staff_availability\nWHERE business_id = ?\n  AND (? IS NULL OR staff_id = ?)\n  AND (? IS NULL OR day_of_week = ?)\n  AND (? IS NULL OR specific_date IS NULL OR specific_date >= ?)\n  AND (? IS NULL OR specific_date IS NULL OR specific_date <= ?)\nORDER BY staff_id ASC, day_of_week ASC, specific_date ASC, start_time ASC\nLIMIT ? OFFSET ?")
        .bind(params.business_id)
        .bind(params.staff_id)
        .bind(params.staff_id)
        .bind(params.day_of_week)
        .bind(params.day_of_week)
        .bind(params.date_from)
        .bind(params.date_from)
        .bind(params.date_to)
        .bind(params.date_to)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `ListStaffAvailabilityForDate` (Go `ListStaffAvailabilityForDateParams`).
pub struct ListStaffAvailabilityForDateParams {
    pub business_id: i64,
    pub staff_id: i64,
    pub on_date: Option<chrono::DateTime<chrono::Utc>>,
    pub day_of_week: Option<i16>,
    pub limit: i32,
    pub offset: i32,
}

/// `ListStaffAvailabilityForDate`: `many`.
pub async fn list_staff_availability_for_date<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListStaffAvailabilityForDateParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, staff_id, day_of_week, specific_date, start_time, end_time, is_available, created_at, updated_at, day_key, date_key\nFROM hr_staff_availability\nWHERE business_id = ?\n  AND staff_id = ?\n  AND (specific_date = ? OR (specific_date IS NULL AND day_of_week = ?))\nORDER BY specific_date IS NULL ASC, start_time ASC\nLIMIT ? OFFSET ?")
        .bind(params.business_id)
        .bind(params.staff_id)
        .bind(params.on_date)
        .bind(params.day_of_week)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `PublishRosterAssignments` (Go `PublishRosterAssignmentsParams`).
pub struct PublishRosterAssignmentsParams {
    pub business_id: i64,
    pub date: chrono::NaiveDate,
    pub date_2: chrono::NaiveDate,
}

/// `PublishRosterAssignments`: `exec`.
pub async fn publish_roster_assignments<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &PublishRosterAssignmentsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE hr_shift_assignments\nSET published = TRUE, updated_at = NOW()\nWHERE business_id = ? AND date >= ? AND date < ?")
        .bind(params.business_id)
        .bind(params.date)
        .bind(params.date_2)
        .execute(exec)
        .await
}

/// Params for `UpdateShift` (Go `UpdateShiftParams`).
pub struct UpdateShiftParams {
    pub staff_id: Option<i64>,
    pub department_id: Option<i64>,
    pub date: Option<chrono::DateTime<chrono::Utc>>,
    pub start_time: String,
    pub end_time: String,
    pub break_duration_minutes: Option<i32>,
    pub status: String,
    pub notes: Option<String>,
    pub location_id: Option<i64>,
    pub name: Option<String>,
    pub color: Option<String>,
    pub id: i64,
    pub business_id: i64,
}

/// `UpdateShift`: `exec`.
pub async fn update_shift<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateShiftParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE hr_shifts\nSET staff_id = ?, department_id = ?, date = ?, start_time = ?, end_time = ?,\n    break_duration_minutes = ?, status = ?, notes = ?,\n    location_id = ?, name = ?, color = ?, updated_at = NOW()\nWHERE id = ? AND business_id = ?")
        .bind(params.staff_id)
        .bind(params.department_id)
        .bind(params.date)
        .bind(&params.start_time)
        .bind(&params.end_time)
        .bind(params.break_duration_minutes)
        .bind(&params.status)
        .bind(&params.notes)
        .bind(params.location_id)
        .bind(&params.name)
        .bind(&params.color)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `UpdateStaffAvailability` (Go `UpdateStaffAvailabilityParams`).
pub struct UpdateStaffAvailabilityParams {
    pub day_of_week: Option<i16>,
    pub specific_date: Option<chrono::DateTime<chrono::Utc>>,
    pub start_time: String,
    pub end_time: String,
    pub is_available: bool,
    pub business_id: i64,
    pub id: i64,
}

/// `UpdateStaffAvailability`: `exec`.
pub async fn update_staff_availability<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateStaffAvailabilityParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE hr_staff_availability\nSET day_of_week = ?, specific_date = ?, start_time = ?, end_time = ?, is_available = ?, updated_at = NOW()\nWHERE business_id = ? AND id = ?")
        .bind(params.day_of_week)
        .bind(params.specific_date)
        .bind(&params.start_time)
        .bind(&params.end_time)
        .bind(params.is_available)
        .bind(params.business_id)
        .bind(params.id)
        .execute(exec)
        .await
}
