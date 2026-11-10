//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `AssignDefaultOwnerRole` (Go `AssignDefaultOwnerRoleParams`).
pub struct AssignDefaultOwnerRoleParams {
    pub business_id: i64,
    pub user_id: Option<i64>,
}

/// `AssignDefaultOwnerRole`: `exec`.
pub async fn assign_default_owner_role<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &AssignDefaultOwnerRoleParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT IGNORE INTO hr_employee_roles (owner_id, business_id, employee_id, role_id)\nSELECT e.owner_id, e.business_id, e.id, r.id\nFROM hr_employees e\nJOIN hr_roles r ON r.business_id = e.business_id\nWHERE e.business_id = ?\n  AND e.user_id = ?\n  AND r.name = 'HR Admin'\n  AND r.is_system = TRUE")
        .bind(params.business_id)
        .bind(params.user_id)
        .execute(exec)
        .await
}

/// Params for `AssignRole` (Go `AssignRoleParams`).
pub struct AssignRoleParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub employee_id: i64,
    pub role_id: i64,
}

/// `AssignRole`: `exec`.
pub async fn assign_role<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &AssignRoleParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO hr_employee_roles (owner_id, business_id, employee_id, role_id)\nVALUES (?, ?, ?, ?)\nON DUPLICATE KEY UPDATE role_id = VALUES(role_id), updated_at = NOW()")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(params.employee_id)
        .bind(params.role_id)
        .execute(exec)
        .await
}

/// Params for `CompleteSetupStep` (Go `CompleteSetupStepParams`).
pub struct CompleteSetupStepParams {
    pub business_id: i64,
    pub step_key: String,
}

/// `CompleteSetupStep`: `exec`.
pub async fn complete_setup_step<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CompleteSetupStepParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE hr_setup_steps\nSET is_completed = TRUE, completed_at = NOW(), updated_at = NOW()\nWHERE business_id = ? AND step_key = ?")
        .bind(params.business_id)
        .bind(&params.step_key)
        .execute(exec)
        .await
}

/// Params for `ConsumeInviteToken` (Go `ConsumeInviteTokenParams`).
pub struct ConsumeInviteTokenParams {
    pub token: String,
}

/// `ConsumeInviteToken`: `execresult`.
pub async fn consume_invite_token<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ConsumeInviteTokenParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE hr_employee_invites\nSET consumed_at = NOW(), updated_at = NOW()\nWHERE token = ? AND consumed_at IS NULL")
        .bind(&params.token)
        .execute(exec)
        .await
}

/// Params for `CountEmployees` (Go `CountEmployeesParams`).
pub struct CountEmployeesParams {
    pub business_id: i64,
    pub department_id: Option<i64>,
    pub location_id: Option<i64>,
    pub status: Option<String>,
}

/// `CountEmployees`: `one`.
pub async fn count_employees<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountEmployeesParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) AS total\nFROM hr_employees\nWHERE business_id = ?\n  AND (? IS NULL OR department_id = ?)\n  AND (? IS NULL OR location_id = ?)\n  AND (? IS NULL OR status = ?)")
        .bind(params.business_id)
        .bind(params.department_id)
        .bind(params.department_id)
        .bind(params.location_id)
        .bind(params.location_id)
        .bind(&params.status)
        .bind(&params.status)
        .fetch_optional(exec)
        .await
}

/// Params for `CountWorkSchedules` (Go `CountWorkSchedulesParams`).
pub struct CountWorkSchedulesParams {
    pub business_id: i64,
}

/// `CountWorkSchedules`: `one`.
pub async fn count_work_schedules<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountWorkSchedulesParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) FROM hr_work_schedules WHERE business_id = ?")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `CreateApprovalChainLevel` (Go `CreateApprovalChainLevelParams`).
pub struct CreateApprovalChainLevelParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub approval_chain_id: i64,
    pub level_order: i32,
    pub approver_type: String,
    pub approver_role_id: Option<i64>,
    pub approver_employee_id: Option<i64>,
}

/// `CreateApprovalChainLevel`: `exec`.
pub async fn create_approval_chain_level<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateApprovalChainLevelParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO hr_approval_chain_levels (owner_id, business_id, approval_chain_id, level_order, approver_type, approver_role_id, approver_employee_id)\nVALUES (?, ?, ?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(params.approval_chain_id)
        .bind(params.level_order)
        .bind(&params.approver_type)
        .bind(params.approver_role_id)
        .bind(params.approver_employee_id)
        .execute(exec)
        .await
}

/// Params for `CreateEmployee` (Go `CreateEmployeeParams`).
pub struct CreateEmployeeParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub user_id: Option<i64>,
    pub department_id: Option<i64>,
    pub location_id: Option<i64>,
    pub work_schedule_id: Option<i64>,
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

/// `CreateEmployee`: `execresult`.
pub async fn create_employee<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateEmployeeParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO hr_employees (\n    owner_id, business_id, user_id, department_id, location_id, work_schedule_id,\n    first_name, last_name, email, phone, employment_type, status,\n    salary, hourly_rate, overtime_rate, start_date, end_date,\n    address, emergency_contact_name, emergency_contact_phone, notes\n) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(params.user_id)
        .bind(params.department_id)
        .bind(params.location_id)
        .bind(params.work_schedule_id)
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

/// Params for `CreateHRDepartment` (Go `CreateHRDepartmentParams`).
pub struct CreateHRDepartmentParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub location_id: Option<i64>,
    pub parent_department_id: Option<i64>,
    pub name: String,
    pub default_work_schedule_id: Option<i64>,
}

/// `CreateHRDepartment`: `execresult`.
pub async fn create_hr_department<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateHRDepartmentParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO hr_departments (owner_id, business_id, location_id, parent_department_id, name, default_work_schedule_id)\nVALUES (?, ?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(params.location_id)
        .bind(params.parent_department_id)
        .bind(&params.name)
        .bind(params.default_work_schedule_id)
        .execute(exec)
        .await
}

/// Params for `CreateInviteToken` (Go `CreateInviteTokenParams`).
pub struct CreateInviteTokenParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub employee_id: i64,
    pub token: String,
    pub expires_at: chrono::NaiveDate,
}

/// `CreateInviteToken`: `execresult`.
pub async fn create_invite_token<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateInviteTokenParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO hr_employee_invites (owner_id, business_id, employee_id, token, expires_at)\nVALUES (?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(params.employee_id)
        .bind(&params.token)
        .bind(params.expires_at)
        .execute(exec)
        .await
}

/// Params for `CreateLeaveBalance` (Go `CreateLeaveBalanceParams`).
pub struct CreateLeaveBalanceParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub employee_id: i64,
    pub leave_type_id: i64,
    pub year: i32,
}

/// `CreateLeaveBalance`: `exec`.
pub async fn create_leave_balance<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateLeaveBalanceParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO hr_leave_balances (owner_id, business_id, employee_id, leave_type_id, year)\nVALUES (?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(params.employee_id)
        .bind(params.leave_type_id)
        .bind(params.year)
        .execute(exec)
        .await
}

/// Params for `CreateLocation` (Go `CreateLocationParams`).
pub struct CreateLocationParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub name: String,
    pub address: Option<String>,
    pub city: Option<String>,
    pub country: Option<String>,
    pub timezone: Option<String>,
    pub latitude: Option<f64>,
    pub longitude: Option<f64>,
    pub is_default: bool,
}

/// `CreateLocation`: `execresult`.
pub async fn create_location<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateLocationParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO hr_locations (owner_id, business_id, name, address, city, country, timezone, latitude, longitude, is_default)\nVALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(&params.name)
        .bind(&params.address)
        .bind(&params.city)
        .bind(&params.country)
        .bind(&params.timezone)
        .bind(params.latitude)
        .bind(params.longitude)
        .bind(params.is_default)
        .execute(exec)
        .await
}

/// Params for `CreateRole` (Go `CreateRoleParams`).
pub struct CreateRoleParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub name: String,
    pub description: Option<String>,
    pub permissions: String,
    pub is_system: bool,
}

/// `CreateRole`: `execresult`.
pub async fn create_role<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateRoleParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO hr_roles (owner_id, business_id, name, description, permissions, is_system)\nVALUES (?, ?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(&params.name)
        .bind(&params.description)
        .bind(&params.permissions)
        .bind(params.is_system)
        .execute(exec)
        .await
}

/// Params for `CreateWorkSchedule` (Go `CreateWorkScheduleParams`).
pub struct CreateWorkScheduleParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub name: String,
    pub schedule_data: String,
    pub is_default: bool,
}

/// `CreateWorkSchedule`: `execresult`.
pub async fn create_work_schedule<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateWorkScheduleParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO hr_work_schedules (owner_id, business_id, name, schedule_data, is_default)\nVALUES (?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(&params.name)
        .bind(&params.schedule_data)
        .bind(params.is_default)
        .execute(exec)
        .await
}

/// Params for `DeactivateEmployee` (Go `DeactivateEmployeeParams`).
pub struct DeactivateEmployeeParams {
    pub end_date: Option<chrono::DateTime<chrono::Utc>>,
    pub id: i64,
    pub business_id: i64,
}

/// `DeactivateEmployee`: `execresult`.
pub async fn deactivate_employee<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DeactivateEmployeeParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE hr_employees\nSET status = 'terminated', end_date = ?, updated_at = NOW()\nWHERE id = ? AND business_id = ?")
        .bind(params.end_date)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `DeleteApprovalChainLevels` (Go `DeleteApprovalChainLevelsParams`).
pub struct DeleteApprovalChainLevelsParams {
    pub business_id: i64,
    pub approval_chain_id: i64,
}

/// `DeleteApprovalChainLevels`: `exec`.
pub async fn delete_approval_chain_levels<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DeleteApprovalChainLevelsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query(
        "DELETE FROM hr_approval_chain_levels\nWHERE business_id = ? AND approval_chain_id = ?",
    )
    .bind(params.business_id)
    .bind(params.approval_chain_id)
    .execute(exec)
    .await
}

/// Params for `DeleteEmployeeForUser` (Go `DeleteEmployeeForUserParams`).
pub struct DeleteEmployeeForUserParams {
    pub business_id: i64,
    pub user_id: Option<i64>,
}

/// `DeleteEmployeeForUser`: `exec`.
pub async fn delete_employee_for_user<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DeleteEmployeeForUserParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("DELETE FROM hr_employees\nWHERE business_id = ? AND user_id = ?")
        .bind(params.business_id)
        .bind(params.user_id)
        .execute(exec)
        .await
}

/// Params for `EnsureEmployeeForUser` (Go `EnsureEmployeeForUserParams`).
pub struct EnsureEmployeeForUserParams {
    pub business_id: i64,
    pub user_id: Option<i64>,
    pub first_name: String,
    pub last_name: String,
    pub email: String,
}

/// `EnsureEmployeeForUser`: `exec`.
pub async fn ensure_employee_for_user<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &EnsureEmployeeForUserParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO hr_employees (owner_id, business_id, user_id, first_name, last_name, email, status)\nSELECT ?, ?, ?,\n       ?, ?, ?, 'active'\nWHERE NOT EXISTS (\n    SELECT 1\n    FROM hr_employees e\n    WHERE e.user_id = ?\n)")
        .bind(params.business_id)
        .bind(params.business_id)
        .bind(params.user_id)
        .bind(&params.first_name)
        .bind(&params.last_name)
        .bind(&params.email)
        .bind(params.user_id)
        .execute(exec)
        .await
}

/// Params for `GetApprovalChain` (Go `GetApprovalChainParams`).
pub struct GetApprovalChainParams {
    pub business_id: i64,
    pub request_type: String,
}

/// `GetApprovalChain`: `one`.
pub async fn get_approval_chain<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetApprovalChainParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, request_type, name, created_at, updated_at\nFROM hr_approval_chains\nWHERE business_id = ? AND request_type = ?")
        .bind(params.business_id)
        .bind(&params.request_type)
        .fetch_optional(exec)
        .await
}

/// Params for `GetApprovalChainLevels` (Go `GetApprovalChainLevelsParams`).
pub struct GetApprovalChainLevelsParams {
    pub business_id: i64,
    pub approval_chain_id: i64,
}

/// `GetApprovalChainLevels`: `many`.
pub async fn get_approval_chain_levels<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetApprovalChainLevelsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, approval_chain_id, level_order, approver_type,\n       approver_role_id, approver_employee_id, created_at\nFROM hr_approval_chain_levels\nWHERE business_id = ? AND approval_chain_id = ?\nORDER BY level_order ASC")
        .bind(params.business_id)
        .bind(params.approval_chain_id)
        .execute(exec)
        .await
}

/// Params for `GetEmployee` (Go `GetEmployeeParams`).
pub struct GetEmployeeParams {
    pub business_id: i64,
    pub id: i64,
}

/// `GetEmployee`: `one`.
pub async fn get_employee<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetEmployeeParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, user_id, department_id, location_id, work_schedule_id,\n       first_name, last_name, email, phone, employment_type, status,\n       salary, hourly_rate, overtime_rate, start_date, end_date,\n       address, emergency_contact_name, emergency_contact_phone, notes,\n       clock_in_pin_hash, created_at, updated_at\nFROM hr_employees\nWHERE business_id = ? AND id = ?")
        .bind(params.business_id)
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetEmployeeIDByUserID` (Go `GetEmployeeIDByUserIDParams`).
pub struct GetEmployeeIDByUserIDParams {
    pub business_id: i64,
    pub user_id: Option<i64>,
}

/// `GetEmployeeIDByUserID`: `one`.
pub async fn get_employee_id_by_user_id<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetEmployeeIDByUserIDParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query(
        "SELECT id FROM hr_employees WHERE business_id = ? AND user_id = ? ORDER BY id ASC LIMIT 1",
    )
    .bind(params.business_id)
    .bind(params.user_id)
    .fetch_optional(exec)
    .await
}

/// Params for `GetEmployeeRoles` (Go `GetEmployeeRolesParams`).
pub struct GetEmployeeRolesParams {
    pub business_id: i64,
    pub employee_id: i64,
}

/// `GetEmployeeRoles`: `many`.
pub async fn get_employee_roles<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetEmployeeRolesParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT r.id, r.owner_id, r.business_id, r.name, r.description, r.permissions, r.is_system, r.created_at, r.updated_at\nFROM hr_roles r\nJOIN hr_employee_roles er ON er.role_id = r.id AND er.business_id = r.business_id\nWHERE er.business_id = ? AND er.employee_id = ?")
        .bind(params.business_id)
        .bind(params.employee_id)
        .execute(exec)
        .await
}

/// Params for `GetHRConfig` (Go `GetHRConfigParams`).
pub struct GetHRConfigParams {
    pub business_id: i64,
}

/// `GetHRConfig`: `one`.
pub async fn get_hr_config<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetHRConfigParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT h.id, h.owner_id, h.business_id,\n       h.fiscal_year_start_month, h.default_currency, h.default_timezone,\n       h.default_work_schedule_id, h.overtime_enabled, h.overtime_multiplier,\n       h.directory_enabled, h.created_at, h.updated_at\nFROM hr_company_config h\nWHERE h.business_id = ?")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetHRDepartment` (Go `GetHRDepartmentParams`).
pub struct GetHRDepartmentParams {
    pub business_id: i64,
    pub id: i64,
}

/// `GetHRDepartment`: `one`.
pub async fn get_hr_department<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetHRDepartmentParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, location_id, parent_department_id, name, default_work_schedule_id, created_at, updated_at\nFROM hr_departments\nWHERE business_id = ? AND id = ?")
        .bind(params.business_id)
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetInviteByToken` (Go `GetInviteByTokenParams`).
pub struct GetInviteByTokenParams {
    pub token: String,
}

/// `GetInviteByToken`: `one`.
pub async fn get_invite_by_token<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetInviteByTokenParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, employee_id, token, expires_at, consumed_at, created_at\nFROM hr_employee_invites\nWHERE token = ? AND consumed_at IS NULL")
        .bind(&params.token)
        .fetch_optional(exec)
        .await
}

/// Params for `GetLocation` (Go `GetLocationParams`).
pub struct GetLocationParams {
    pub business_id: i64,
    pub id: i64,
}

/// `GetLocation`: `one`.
pub async fn get_location<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetLocationParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, name, slug, address, city, country, timezone, latitude, longitude, is_default, is_active, created_at, updated_at\nFROM hr_locations\nWHERE business_id = ? AND id = ?")
        .bind(params.business_id)
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetRole` (Go `GetRoleParams`).
pub struct GetRoleParams {
    pub business_id: i64,
    pub id: i64,
}

/// `GetRole`: `one`.
pub async fn get_role<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetRoleParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, name, description, permissions, is_system, created_at, updated_at\nFROM hr_roles\nWHERE business_id = ? AND id = ?")
        .bind(params.business_id)
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetSetupStepByKey` (Go `GetSetupStepByKeyParams`).
pub struct GetSetupStepByKeyParams {
    pub business_id: i64,
    pub step_key: String,
}

/// `GetSetupStepByKey`: `one`.
pub async fn get_setup_step_by_key<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetSetupStepByKeyParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, step_key, step_label, is_completed, completed_at, created_at, updated_at\nFROM hr_setup_steps\nWHERE business_id = ? AND step_key = ?")
        .bind(params.business_id)
        .bind(&params.step_key)
        .fetch_optional(exec)
        .await
}

/// Params for `GetSetupSteps` (Go `GetSetupStepsParams`).
pub struct GetSetupStepsParams {
    pub business_id: i64,
}

/// `GetSetupSteps`: `many`.
pub async fn get_setup_steps<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetSetupStepsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, step_key, step_label, is_completed, completed_at, created_at, updated_at\nFROM hr_setup_steps\nWHERE business_id = ?\nORDER BY id ASC")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `GetWorkSchedule` (Go `GetWorkScheduleParams`).
pub struct GetWorkScheduleParams {
    pub business_id: i64,
    pub id: i64,
}

/// `GetWorkSchedule`: `one`.
pub async fn get_work_schedule<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetWorkScheduleParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, name, schedule_data, is_default, created_at, updated_at\nFROM hr_work_schedules\nWHERE business_id = ? AND id = ?")
        .bind(params.business_id)
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetWorkScheduleForEmployee` (Go `GetWorkScheduleForEmployeeParams`).
pub struct GetWorkScheduleForEmployeeParams {
    pub business_id: i64,
    pub id: i64,
}

/// `GetWorkScheduleForEmployee`: `one`.
pub async fn get_work_schedule_for_employee<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetWorkScheduleForEmployeeParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT ws.id, ws.owner_id, ws.business_id, ws.name, ws.schedule_data, ws.is_default, ws.created_at, ws.updated_at\nFROM hr_work_schedules ws\nJOIN hr_employees e ON e.work_schedule_id = ws.id AND e.business_id = ws.business_id\nWHERE ws.business_id = ? AND e.id = ?")
        .bind(params.business_id)
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `HasEmployeeRole` (Go `HasEmployeeRoleParams`).
pub struct HasEmployeeRoleParams {
    pub business_id: i64,
    pub employee_id: i64,
    pub role_id: i64,
}

/// `HasEmployeeRole`: `one`.
pub async fn has_employee_role<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &HasEmployeeRoleParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*)\nFROM hr_employee_roles\nWHERE business_id = ? AND employee_id = ? AND role_id = ?")
        .bind(params.business_id)
        .bind(params.employee_id)
        .bind(params.role_id)
        .fetch_optional(exec)
        .await
}

/// Params for `InsertRoleIfMissing` (Go `InsertRoleIfMissingParams`).
pub struct InsertRoleIfMissingParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub name: String,
    pub description: Option<String>,
    pub permissions: String,
    pub is_system: bool,
}

/// `InsertRoleIfMissing`: `execresult`.
pub async fn insert_role_if_missing<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &InsertRoleIfMissingParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT IGNORE INTO hr_roles (owner_id, business_id, name, description, permissions, is_system)\nVALUES (?, ?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(&params.name)
        .bind(&params.description)
        .bind(&params.permissions)
        .bind(params.is_system)
        .execute(exec)
        .await
}

/// Params for `LinkInviteAcceptance` (Go `LinkInviteAcceptanceParams`).
pub struct LinkInviteAcceptanceParams {
    pub user_id: Option<i64>,
    pub id: i64,
    pub business_id: i64,
}

/// `LinkInviteAcceptance`: `execresult`.
pub async fn link_invite_acceptance<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &LinkInviteAcceptanceParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE hr_employees\nSET user_id = ?, status = 'active', updated_at = NOW()\nWHERE id = ? AND business_id = ? AND status = 'invited'")
        .bind(params.user_id)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `ListEmployees` (Go `ListEmployeesParams`).
pub struct ListEmployeesParams {
    pub business_id: i64,
    pub department_id: Option<i64>,
    pub location_id: Option<i64>,
    pub status: Option<String>,
    pub limit: i32,
    pub offset: i32,
}

/// `ListEmployees`: `many`.
pub async fn list_employees<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListEmployeesParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, user_id, department_id, location_id, work_schedule_id,\n       first_name, last_name, email, phone, employment_type, status,\n       salary, hourly_rate, overtime_rate, start_date, end_date,\n       address, emergency_contact_name, emergency_contact_phone, notes,\n       clock_in_pin_hash, created_at, updated_at\nFROM hr_employees\nWHERE business_id = ?\n  AND (? IS NULL OR department_id = ?)\n  AND (? IS NULL OR location_id = ?)\n  AND (? IS NULL OR status = ?)\nORDER BY created_at DESC\nLIMIT ? OFFSET ?")
        .bind(params.business_id)
        .bind(params.department_id)
        .bind(params.department_id)
        .bind(params.location_id)
        .bind(params.location_id)
        .bind(&params.status)
        .bind(&params.status)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `ListHRDepartments` (Go `ListHRDepartmentsParams`).
pub struct ListHRDepartmentsParams {
    pub business_id: i64,
    pub location_id: Option<i64>,
}

/// `ListHRDepartments`: `many`.
pub async fn list_hr_departments<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListHRDepartmentsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, location_id, parent_department_id, name, default_work_schedule_id, created_at, updated_at\nFROM hr_departments\nWHERE business_id = ?\n  AND (? = 0 OR location_id = ?)\nORDER BY name ASC\nLIMIT 500")
        .bind(params.business_id)
        .bind(params.location_id)
        .bind(params.location_id)
        .execute(exec)
        .await
}

/// Params for `ListLocations` (Go `ListLocationsParams`).
pub struct ListLocationsParams {
    pub business_id: i64,
}

/// `ListLocations`: `many`.
pub async fn list_locations<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListLocationsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, name, slug, address, city, country, timezone, latitude, longitude, is_default, is_active, created_at, updated_at\nFROM hr_locations\nWHERE business_id = ?\nORDER BY name ASC\nLIMIT 500")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `ListPublicLocations` (Go `ListPublicLocationsParams`).
pub struct ListPublicLocationsParams {
    pub business_id: i64,
}

/// `ListPublicLocations`: `many`.
pub async fn list_public_locations<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListPublicLocationsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, name, COALESCE(slug, '') AS slug, COALESCE(address, '') AS address,\n       COALESCE(timezone, '') AS timezone, is_active\nFROM hr_locations\nWHERE business_id = ? AND is_active = TRUE\nORDER BY is_default DESC, name ASC\nLIMIT 100")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `ListRoles` (Go `ListRolesParams`).
pub struct ListRolesParams {
    pub business_id: i64,
}

/// `ListRoles`: `many`.
pub async fn list_roles<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListRolesParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, name, description, permissions, is_system, created_at, updated_at\nFROM hr_roles\nWHERE business_id = ?\nORDER BY is_system DESC, name ASC\nLIMIT 500")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// `ListTerminatedEmployees`: `many`.
pub async fn list_terminated_employees<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT DISTINCT business_id, user_id\nFROM hr_employees\nWHERE status = 'terminated' AND user_id IS NOT NULL\nORDER BY business_id, user_id")
        .execute(exec)
        .await
}

/// Params for `ListWorkSchedules` (Go `ListWorkSchedulesParams`).
pub struct ListWorkSchedulesParams {
    pub business_id: i64,
    pub limit: i32,
    pub offset: i32,
}

/// `ListWorkSchedules`: `many`.
pub async fn list_work_schedules<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListWorkSchedulesParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, name, schedule_data, is_default, created_at, updated_at\nFROM hr_work_schedules\nWHERE business_id = ?\nORDER BY name ASC\nLIMIT ? OFFSET ?")
        .bind(params.business_id)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `UpdateEmployee` (Go `UpdateEmployeeParams`).
pub struct UpdateEmployeeParams {
    pub user_id: Option<i64>,
    pub department_id: Option<i64>,
    pub location_id: Option<i64>,
    pub work_schedule_id: Option<i64>,
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

/// `UpdateEmployee`: `execrows`.
pub async fn update_employee<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateEmployeeParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE hr_employees\nSET user_id = ?, department_id = ?, location_id = ?, work_schedule_id = ?,\n    first_name = ?, last_name = ?, email = ?, phone = ?,\n    employment_type = ?, status = ?, salary = ?, hourly_rate = ?,\n    overtime_rate = ?, start_date = ?, end_date = ?,\n    address = ?, emergency_contact_name = ?, emergency_contact_phone = ?,\n    notes = ?, updated_at = NOW()\nWHERE id = ? AND business_id = ?")
        .bind(params.user_id)
        .bind(params.department_id)
        .bind(params.location_id)
        .bind(params.work_schedule_id)
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

/// Params for `UpdateHRDepartment` (Go `UpdateHRDepartmentParams`).
pub struct UpdateHRDepartmentParams {
    pub location_id: Option<i64>,
    pub parent_department_id: Option<i64>,
    pub name: String,
    pub default_work_schedule_id: Option<i64>,
    pub id: i64,
    pub business_id: i64,
}

/// `UpdateHRDepartment`: `execresult`.
pub async fn update_hr_department<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateHRDepartmentParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE hr_departments\nSET location_id = ?, parent_department_id = ?, name = ?, default_work_schedule_id = ?, updated_at = NOW()\nWHERE id = ? AND business_id = ?")
        .bind(params.location_id)
        .bind(params.parent_department_id)
        .bind(&params.name)
        .bind(params.default_work_schedule_id)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `UpdateLocation` (Go `UpdateLocationParams`).
pub struct UpdateLocationParams {
    pub name: String,
    pub address: Option<String>,
    pub city: Option<String>,
    pub country: Option<String>,
    pub timezone: Option<String>,
    pub latitude: Option<f64>,
    pub longitude: Option<f64>,
    pub is_default: bool,
    pub id: i64,
    pub business_id: i64,
}

/// `UpdateLocation`: `execresult`.
pub async fn update_location<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateLocationParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE hr_locations\nSET name = ?, address = ?, city = ?, country = ?, timezone = ?, latitude = ?, longitude = ?, is_default = ?, updated_at = NOW()\nWHERE id = ? AND business_id = ?")
        .bind(&params.name)
        .bind(&params.address)
        .bind(&params.city)
        .bind(&params.country)
        .bind(&params.timezone)
        .bind(params.latitude)
        .bind(params.longitude)
        .bind(params.is_default)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `UpdateRole` (Go `UpdateRoleParams`).
pub struct UpdateRoleParams {
    pub name: String,
    pub description: Option<String>,
    pub permissions: String,
    pub id: i64,
    pub business_id: i64,
}

/// `UpdateRole`: `execresult`.
pub async fn update_role<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateRoleParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE hr_roles\nSET name = ?, description = ?, permissions = ?, updated_at = NOW()\nWHERE id = ? AND business_id = ?")
        .bind(&params.name)
        .bind(&params.description)
        .bind(&params.permissions)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `UpdateWorkSchedule` (Go `UpdateWorkScheduleParams`).
pub struct UpdateWorkScheduleParams {
    pub name: String,
    pub schedule_data: String,
    pub is_default: bool,
    pub id: i64,
    pub business_id: i64,
}

/// `UpdateWorkSchedule`: `execresult`.
pub async fn update_work_schedule<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateWorkScheduleParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE hr_work_schedules\nSET name = ?, schedule_data = ?, is_default = ?, updated_at = NOW()\nWHERE id = ? AND business_id = ?")
        .bind(&params.name)
        .bind(&params.schedule_data)
        .bind(params.is_default)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `UpsertApprovalChain` (Go `UpsertApprovalChainParams`).
pub struct UpsertApprovalChainParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub request_type: String,
    pub name: String,
}

/// `UpsertApprovalChain`: `execresult`.
pub async fn upsert_approval_chain<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpsertApprovalChainParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO hr_approval_chains (owner_id, business_id, request_type, name)\nVALUES (?, ?, ?, ?)\nON DUPLICATE KEY UPDATE name = VALUES(name), updated_at = NOW()")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(&params.request_type)
        .bind(&params.name)
        .execute(exec)
        .await
}

/// Params for `UpsertHRConfig` (Go `UpsertHRConfigParams`).
pub struct UpsertHRConfigParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub fiscal_year_start_month: i32,
    pub default_currency: String,
    pub default_timezone: String,
    pub default_work_schedule_id: Option<i64>,
    pub overtime_enabled: bool,
    pub overtime_multiplier: String,
    pub directory_enabled: bool,
}

/// `UpsertHRConfig`: `execresult`.
pub async fn upsert_hr_config<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpsertHRConfigParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO hr_company_config (\n    owner_id, business_id, fiscal_year_start_month, default_currency,\n    default_timezone, default_work_schedule_id, overtime_enabled, overtime_multiplier,\n    directory_enabled\n) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)\nON DUPLICATE KEY UPDATE\n    fiscal_year_start_month = VALUES(fiscal_year_start_month),\n    default_currency = VALUES(default_currency),\n    default_timezone = VALUES(default_timezone),\n    default_work_schedule_id = VALUES(default_work_schedule_id),\n    overtime_enabled = VALUES(overtime_enabled),\n    overtime_multiplier = VALUES(overtime_multiplier),\n    directory_enabled = VALUES(directory_enabled),\n    updated_at = NOW()")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(params.fiscal_year_start_month)
        .bind(&params.default_currency)
        .bind(&params.default_timezone)
        .bind(params.default_work_schedule_id)
        .bind(params.overtime_enabled)
        .bind(&params.overtime_multiplier)
        .bind(params.directory_enabled)
        .execute(exec)
        .await
}
