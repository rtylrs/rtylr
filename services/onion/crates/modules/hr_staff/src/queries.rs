//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `ClockIn` (Go `ClockInParams`).
pub struct ClockInParams {
    pub business_id: i64,
    pub staff_id: Option<i64>,
}

/// `ClockIn`: `exec`.
pub async fn clock_in<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ClockInParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE hr_shifts SET clock_in_at = NOW(), status = 'in_progress', updated_at = NOW()\nWHERE business_id = ? AND staff_id = ? AND date = CURDATE() AND clock_in_at IS NULL\nORDER BY start_time ASC LIMIT 1")
        .bind(params.business_id)
        .bind(params.staff_id)
        .execute(exec)
        .await
}

/// Params for `ClockOut` (Go `ClockOutParams`).
pub struct ClockOutParams {
    pub business_id: i64,
    pub staff_id: Option<i64>,
}

/// `ClockOut`: `exec`.
pub async fn clock_out<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ClockOutParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE hr_shifts SET clock_out_at = NOW(), status = 'completed',\n       hours_worked = GREATEST(0, (TIMESTAMPDIFF(MINUTE, clock_in_at, NOW()) - COALESCE(break_duration_minutes, 0)) / 60.0), updated_at = NOW()\nWHERE business_id = ? AND staff_id = ? AND clock_in_at IS NOT NULL AND clock_out_at IS NULL\nORDER BY clock_in_at DESC LIMIT 1")
        .bind(params.business_id)
        .bind(params.staff_id)
        .execute(exec)
        .await
}

/// Params for `CountActiveStaff` (Go `CountActiveStaffParams`).
pub struct CountActiveStaffParams {
    pub business_id: i64,
}

/// `CountActiveStaff`: `one`.
pub async fn count_active_staff<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountActiveStaffParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) FROM hr_staff WHERE business_id = ? AND status = 'active'")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `CountDraftPayroll` (Go `CountDraftPayrollParams`).
pub struct CountDraftPayrollParams {
    pub business_id: i64,
}

/// `CountDraftPayroll`: `one`.
pub async fn count_draft_payroll<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountDraftPayrollParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query(
        "SELECT COUNT(*) FROM hr_payroll_records WHERE business_id = ? AND status = 'draft'",
    )
    .bind(params.business_id)
    .fetch_optional(exec)
    .await
}

/// Params for `CountOnShiftToday` (Go `CountOnShiftTodayParams`).
pub struct CountOnShiftTodayParams {
    pub business_id: i64,
}

/// `CountOnShiftToday`: `one`.
pub async fn count_on_shift_today<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountOnShiftTodayParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) FROM hr_shifts WHERE business_id = ? AND date = CURDATE() AND status IN ('scheduled','in_progress')")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `CountPendingTimeOff` (Go `CountPendingTimeOffParams`).
pub struct CountPendingTimeOffParams {
    pub business_id: i64,
}

/// `CountPendingTimeOff`: `one`.
pub async fn count_pending_time_off<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountPendingTimeOffParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query(
        "SELECT COUNT(*) FROM hr_time_off_requests WHERE business_id = ? AND status = 'pending'",
    )
    .bind(params.business_id)
    .fetch_optional(exec)
    .await
}

/// Params for `CreateStaff` (Go `CreateStaffParams`).
pub struct CreateStaffParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub user_id: Option<i64>,
    pub department_id: Option<i64>,
    pub manager_staff_id: Option<i64>,
    pub first_name: String,
    pub last_name: String,
    pub email: String,
    pub phone: Option<String>,
    pub role: String,
    pub employment_type: String,
    pub status: String,
    pub salary: Option<String>,
    pub hourly_rate: Option<String>,
    pub overtime_rate: Option<String>,
    pub start_date: chrono::NaiveDate,
    pub end_date: Option<chrono::DateTime<chrono::Utc>>,
    pub address: Option<String>,
    pub emergency_contact_name: Option<String>,
    pub emergency_contact_phone: Option<String>,
    pub apps: String,
    pub notes: Option<String>,
    pub photo_url: Option<String>,
    pub bio: Option<String>,
}

/// `CreateStaff`: `execresult`.
pub async fn create_staff<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateStaffParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO hr_staff\n(owner_id, business_id, user_id, department_id, manager_staff_id, first_name, last_name, email, phone, role,\n employment_type, status, salary, hourly_rate, overtime_rate, start_date, end_date,\n address, emergency_contact_name, emergency_contact_phone, apps, notes, photo_url, bio)\nVALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(params.user_id)
        .bind(params.department_id)
        .bind(params.manager_staff_id)
        .bind(&params.first_name)
        .bind(&params.last_name)
        .bind(&params.email)
        .bind(&params.phone)
        .bind(&params.role)
        .bind(&params.employment_type)
        .bind(&params.status)
        .bind(&params.salary)
        .bind(&params.hourly_rate)
        .bind(&params.overtime_rate)
        .bind(params.start_date)
        .bind(params.end_date)
        .bind(&params.address)
        .bind(&params.emergency_contact_name)
        .bind(&params.emergency_contact_phone)
        .bind(&params.apps)
        .bind(&params.notes)
        .bind(&params.photo_url)
        .bind(&params.bio)
        .execute(exec)
        .await
}

/// Params for `DeleteStaffForUser` (Go `DeleteStaffForUserParams`).
pub struct DeleteStaffForUserParams {
    pub business_id: i64,
    pub user_id: Option<i64>,
}

/// `DeleteStaffForUser`: `exec`.
pub async fn delete_staff_for_user<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DeleteStaffForUserParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("DELETE FROM hr_staff\nWHERE business_id = ? AND user_id = ?")
        .bind(params.business_id)
        .bind(params.user_id)
        .execute(exec)
        .await
}

/// Params for `EnsureStaffForUser` (Go `EnsureStaffForUserParams`).
pub struct EnsureStaffForUserParams {
    pub role: String,
    pub business_id: i64,
    pub user_id: Option<i64>,
}

/// `EnsureStaffForUser`: `exec`.
pub async fn ensure_staff_for_user<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &EnsureStaffForUserParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO hr_staff\n    (owner_id, business_id, user_id, employee_id, department_id, first_name, last_name, email,\n     phone, role, employment_type, status, salary, hourly_rate, overtime_rate, start_date,\n     end_date, address, emergency_contact_name, emergency_contact_phone, apps, notes, photo_url, bio)\nSELECT e.business_id, e.business_id, e.user_id, e.id, e.department_id, e.first_name, e.last_name, e.email,\n       e.phone, ?,\n       CASE e.employment_type\n           WHEN 'part_time' THEN 'part_time'\n           WHEN 'contractor' THEN 'contractor'\n           WHEN 'intern' THEN 'intern'\n           ELSE 'full_time'\n       END,\n       'active', e.salary, e.hourly_rate, e.overtime_rate, COALESCE(e.start_date, CURDATE()),\n       e.end_date, e.address, e.emergency_contact_name, e.emergency_contact_phone,\n       JSON_ARRAY(), e.notes, NULL, NULL\nFROM hr_employees e\nWHERE e.business_id = ?\n  AND e.user_id = ?\n  AND NOT EXISTS (\n      SELECT 1\n      FROM hr_staff s\n      WHERE s.business_id = e.business_id\n        AND s.user_id = e.user_id\n  )")
        .bind(&params.role)
        .bind(params.business_id)
        .bind(params.user_id)
        .execute(exec)
        .await
}

/// Params for `FindEmployeeIDByEmail` (Go `FindEmployeeIDByEmailParams`).
pub struct FindEmployeeIDByEmailParams {
    pub business_id: i64,
    pub email: String,
}

/// `FindEmployeeIDByEmail`: `one`.
pub async fn find_employee_id_by_email<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &FindEmployeeIDByEmailParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id FROM hr_employees\nWHERE business_id = ? AND email = ?\nORDER BY id ASC\nLIMIT 1")
        .bind(params.business_id)
        .bind(&params.email)
        .fetch_optional(exec)
        .await
}

/// Params for `FindEmployeeIDByUser` (Go `FindEmployeeIDByUserParams`).
pub struct FindEmployeeIDByUserParams {
    pub business_id: i64,
    pub user_id: Option<i64>,
}

/// `FindEmployeeIDByUser`: `one`.
pub async fn find_employee_id_by_user<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &FindEmployeeIDByUserParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id FROM hr_employees\nWHERE business_id = ? AND user_id = ?\nORDER BY id ASC\nLIMIT 1")
        .bind(params.business_id)
        .bind(params.user_id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetStaffByID` (Go `GetStaffByIDParams`).
pub struct GetStaffByIDParams {
    pub id: i64,
    pub business_id: i64,
}

/// `GetStaffByID`: `one`.
pub async fn get_staff_by_id<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetStaffByIDParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, user_id, department_id, manager_staff_id, first_name, last_name, email, phone,\n       role, employment_type, status, salary, hourly_rate, overtime_rate,\n       start_date, end_date, address, emergency_contact_name, emergency_contact_phone,\n       apps, notes, photo_url, bio, created_at, updated_at\nFROM hr_staff\nWHERE id = ? AND business_id = ?")
        .bind(params.id)
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetStaffEmployeeSyncRow` (Go `GetStaffEmployeeSyncRowParams`).
pub struct GetStaffEmployeeSyncRowParams {
    pub business_id: i64,
    pub id: i64,
}

/// `GetStaffEmployeeSyncRow`: `one`.
pub async fn get_staff_employee_sync_row<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetStaffEmployeeSyncRowParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT owner_id, business_id, id, employee_id, user_id, department_id, first_name, last_name,\n       email, phone, employment_type, status, salary, hourly_rate, overtime_rate,\n       start_date, end_date, address, emergency_contact_name, emergency_contact_phone, notes\nFROM hr_staff\nWHERE business_id = ? AND id = ?")
        .bind(params.business_id)
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `InsertEmployeeFromStaff` (Go `InsertEmployeeFromStaffParams`).
pub struct InsertEmployeeFromStaffParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub user_id: Option<i64>,
    pub department_id: Option<i64>,
    pub first_name: String,
    pub last_name: String,
    pub email: String,
    pub phone: Option<String>,
    pub employment_type: String,
    pub status: String,
    pub salary: Option<String>,
    pub hourly_rate: Option<String>,
    pub overtime_rate: Option<String>,
    pub start_date: Option<chrono::DateTime<chrono::Utc>>,
    pub end_date: Option<chrono::DateTime<chrono::Utc>>,
    pub address: Option<String>,
    pub emergency_contact_name: Option<String>,
    pub emergency_contact_phone: Option<String>,
    pub notes: Option<String>,
}

/// `InsertEmployeeFromStaff`: `execresult`.
pub async fn insert_employee_from_staff<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &InsertEmployeeFromStaffParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO hr_employees\n    (owner_id, business_id, user_id, department_id, first_name, last_name, email, phone,\n     employment_type, status, salary, hourly_rate, overtime_rate, start_date,\n     end_date, address, emergency_contact_name, emergency_contact_phone, notes)\nVALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(params.user_id)
        .bind(params.department_id)
        .bind(&params.first_name)
        .bind(&params.last_name)
        .bind(&params.email)
        .bind(&params.phone)
        .bind(&params.employment_type)
        .bind(&params.status)
        .bind(&params.salary)
        .bind(&params.hourly_rate)
        .bind(&params.overtime_rate)
        .bind(params.start_date)
        .bind(params.end_date)
        .bind(&params.address)
        .bind(&params.emergency_contact_name)
        .bind(&params.emergency_contact_phone)
        .bind(&params.notes)
        .execute(exec)
        .await
}

/// Params for `LinkStaffEmployee` (Go `LinkStaffEmployeeParams`).
pub struct LinkStaffEmployeeParams {
    pub employee_id: Option<i64>,
    pub id: i64,
    pub business_id: i64,
}

/// `LinkStaffEmployee`: `exec`.
pub async fn link_staff_employee<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &LinkStaffEmployeeParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE hr_staff SET employee_id = ? WHERE id = ? AND business_id = ?")
        .bind(params.employee_id)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `LinkStaffEmployeeForUser` (Go `LinkStaffEmployeeForUserParams`).
pub struct LinkStaffEmployeeForUserParams {
    pub business_id: i64,
    pub user_id: Option<i64>,
}

/// `LinkStaffEmployeeForUser`: `exec`.
pub async fn link_staff_employee_for_user<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &LinkStaffEmployeeForUserParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE hr_staff s\nJOIN hr_employees e\n  ON e.business_id = s.business_id\n AND e.user_id = s.user_id\nSET s.employee_id = e.id\nWHERE s.business_id = ?\n  AND s.user_id = ?\n  AND (s.employee_id IS NULL OR s.employee_id = 0)")
        .bind(params.business_id)
        .bind(params.user_id)
        .execute(exec)
        .await
}

/// Params for `ListManagedUserIDs` (Go `ListManagedUserIDsParams`).
pub struct ListManagedUserIDsParams {
    pub user_id: Option<i64>,
    pub business_id: i64,
    pub user_id_2: Option<i64>,
    pub business_id_2: i64,
}

/// `ListManagedUserIDs`: `many`.
pub async fn list_managed_user_i_ds<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListManagedUserIDsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT DISTINCT s.user_id\nFROM hr_staff s\nJOIN hr_departments d ON d.id = s.department_id AND d.business_id = s.business_id\nJOIN hr_staff m ON m.id = d.manager_staff_id AND m.business_id = s.business_id\nWHERE m.user_id = ?\n  AND s.business_id = ?\n  AND s.user_id IS NOT NULL\n  AND s.user_id <> m.user_id\n  AND s.status <> 'terminated'\nUNION\nSELECT DISTINCT s2.user_id\nFROM hr_staff s2\nJOIN hr_staff m2 ON m2.id = s2.manager_staff_id AND m2.business_id = s2.business_id\nWHERE m2.user_id = ?\n  AND s2.business_id = ?\n  AND s2.user_id IS NOT NULL\n  AND s2.user_id <> m2.user_id\n  AND s2.manager_staff_id <> s2.id\n  AND s2.status <> 'terminated'")
        .bind(params.user_id)
        .bind(params.business_id)
        .bind(params.user_id_2)
        .bind(params.business_id_2)
        .execute(exec)
        .await
}

/// Params for `TerminateStaff` (Go `TerminateStaffParams`).
pub struct TerminateStaffParams {
    pub id: i64,
    pub business_id: i64,
}

/// `TerminateStaff`: `exec`.
pub async fn terminate_staff<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &TerminateStaffParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE hr_staff SET status = 'terminated', updated_at = NOW() WHERE id = ? AND business_id = ?")
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `UpdateEmployeeFromStaff` (Go `UpdateEmployeeFromStaffParams`).
pub struct UpdateEmployeeFromStaffParams {
    pub user_id: Option<i64>,
    pub department_id: Option<i64>,
    pub first_name: String,
    pub last_name: String,
    pub email: String,
    pub phone: Option<String>,
    pub employment_type: String,
    pub status: String,
    pub salary: Option<String>,
    pub hourly_rate: Option<String>,
    pub overtime_rate: Option<String>,
    pub start_date: Option<chrono::DateTime<chrono::Utc>>,
    pub end_date: Option<chrono::DateTime<chrono::Utc>>,
    pub address: Option<String>,
    pub emergency_contact_name: Option<String>,
    pub emergency_contact_phone: Option<String>,
    pub notes: Option<String>,
    pub id: i64,
    pub business_id: i64,
}

/// `UpdateEmployeeFromStaff`: `exec`.
pub async fn update_employee_from_staff<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateEmployeeFromStaffParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE hr_employees\nSET user_id = COALESCE(?, user_id),\n    department_id = ?, first_name = ?, last_name = ?, email = ?, phone = ?,\n    employment_type = ?, status = ?, salary = ?, hourly_rate = ?,\n    overtime_rate = ?, start_date = ?, end_date = ?, address = ?,\n    emergency_contact_name = ?, emergency_contact_phone = ?, notes = ?,\n    updated_at = NOW()\nWHERE id = ? AND business_id = ?")
        .bind(params.user_id)
        .bind(params.department_id)
        .bind(&params.first_name)
        .bind(&params.last_name)
        .bind(&params.email)
        .bind(&params.phone)
        .bind(&params.employment_type)
        .bind(&params.status)
        .bind(&params.salary)
        .bind(&params.hourly_rate)
        .bind(&params.overtime_rate)
        .bind(params.start_date)
        .bind(params.end_date)
        .bind(&params.address)
        .bind(&params.emergency_contact_name)
        .bind(&params.emergency_contact_phone)
        .bind(&params.notes)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `UpdateStaff` (Go `UpdateStaffParams`).
pub struct UpdateStaffParams {
    pub department_id: Option<i64>,
    pub manager_staff_id: Option<i64>,
    pub first_name: String,
    pub last_name: String,
    pub email: String,
    pub phone: Option<String>,
    pub role: String,
    pub employment_type: String,
    pub status: String,
    pub salary: Option<String>,
    pub hourly_rate: Option<String>,
    pub overtime_rate: Option<String>,
    pub end_date: Option<chrono::DateTime<chrono::Utc>>,
    pub address: Option<String>,
    pub emergency_contact_name: Option<String>,
    pub emergency_contact_phone: Option<String>,
    pub apps: String,
    pub notes: Option<String>,
    pub photo_url: Option<String>,
    pub bio: Option<String>,
    pub id: i64,
    pub business_id: i64,
}

/// `UpdateStaff`: `exec`.
pub async fn update_staff<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateStaffParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE hr_staff\nSET department_id = ?, manager_staff_id = ?, first_name = ?, last_name = ?, email = ?, phone = ?,\n    role = ?, employment_type = ?, status = ?, salary = ?, hourly_rate = ?,\n    overtime_rate = ?, end_date = ?, address = ?, emergency_contact_name = ?,\n    emergency_contact_phone = ?, apps = ?, notes = ?, photo_url = ?, bio = ?,\n    updated_at = NOW()\nWHERE id = ? AND business_id = ?")
        .bind(params.department_id)
        .bind(params.manager_staff_id)
        .bind(&params.first_name)
        .bind(&params.last_name)
        .bind(&params.email)
        .bind(&params.phone)
        .bind(&params.role)
        .bind(&params.employment_type)
        .bind(&params.status)
        .bind(&params.salary)
        .bind(&params.hourly_rate)
        .bind(&params.overtime_rate)
        .bind(params.end_date)
        .bind(&params.address)
        .bind(&params.emergency_contact_name)
        .bind(&params.emergency_contact_phone)
        .bind(&params.apps)
        .bind(&params.notes)
        .bind(&params.photo_url)
        .bind(&params.bio)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}
