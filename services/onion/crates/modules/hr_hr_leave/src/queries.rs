//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `AddToEntitledDays` (Go `AddToEntitledDaysParams`).
pub struct AddToEntitledDaysParams {
    pub entitled_days: String,
    pub last_accrual_month: i32,
    pub business_id: i64,
    pub employee_id: i64,
    pub leave_type_id: i64,
    pub year: i32,
    pub last_accrual_month_2: i32,
}

/// `AddToEntitledDays`: `execresult`.
pub async fn add_to_entitled_days<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &AddToEntitledDaysParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE hr_leave_balances\nSET entitled_days = entitled_days + ?, last_accrual_month = ?, updated_at = NOW()\nWHERE business_id = ? AND employee_id = ? AND leave_type_id = ? AND year = ?\n  AND last_accrual_month < ?")
        .bind(&params.entitled_days)
        .bind(params.last_accrual_month)
        .bind(params.business_id)
        .bind(params.employee_id)
        .bind(params.leave_type_id)
        .bind(params.year)
        .bind(params.last_accrual_month_2)
        .execute(exec)
        .await
}

/// Params for `AddToPendingDays` (Go `AddToPendingDaysParams`).
pub struct AddToPendingDaysParams {
    pub pending_days: String,
    pub business_id: i64,
    pub employee_id: i64,
    pub leave_type_id: i64,
    pub year: i32,
}

/// `AddToPendingDays`: `execresult`.
pub async fn add_to_pending_days<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &AddToPendingDaysParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE hr_leave_balances\nSET pending_days = pending_days + ?, updated_at = NOW()\nWHERE business_id = ? AND employee_id = ? AND leave_type_id = ? AND year = ?")
        .bind(&params.pending_days)
        .bind(params.business_id)
        .bind(params.employee_id)
        .bind(params.leave_type_id)
        .bind(params.year)
        .execute(exec)
        .await
}

/// Params for `AddToUsedDays` (Go `AddToUsedDaysParams`).
pub struct AddToUsedDaysParams {
    pub used_days: String,
    pub pending_days: String,
    pub business_id: i64,
    pub employee_id: i64,
    pub leave_type_id: i64,
    pub year: i32,
}

/// `AddToUsedDays`: `execresult`.
pub async fn add_to_used_days<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &AddToUsedDaysParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE hr_leave_balances\nSET used_days = used_days + ?, pending_days = GREATEST(pending_days - ?, 0), updated_at = NOW()\nWHERE business_id = ? AND employee_id = ? AND leave_type_id = ? AND year = ?")
        .bind(&params.used_days)
        .bind(&params.pending_days)
        .bind(params.business_id)
        .bind(params.employee_id)
        .bind(params.leave_type_id)
        .bind(params.year)
        .execute(exec)
        .await
}

/// Params for `AdvanceLeaveApprovalLevel` (Go `AdvanceLeaveApprovalLevelParams`).
pub struct AdvanceLeaveApprovalLevelParams {
    pub id: i64,
    pub business_id: i64,
    pub current_approval_level: i32,
}

/// `AdvanceLeaveApprovalLevel`: `execresult`.
pub async fn advance_leave_approval_level<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &AdvanceLeaveApprovalLevelParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE hr_leave_requests\nSET current_approval_level = current_approval_level + 1, updated_at = NOW()\nWHERE id = ? AND business_id = ? AND status = 'pending' AND current_approval_level = ?")
        .bind(params.id)
        .bind(params.business_id)
        .bind(params.current_approval_level)
        .execute(exec)
        .await
}

/// Params for `ApplyYearEndCarryover` (Go `ApplyYearEndCarryoverParams`).
pub struct ApplyYearEndCarryoverParams {
    pub year: i32,
    pub business_id: i64,
    pub year_2: i32,
}

/// `ApplyYearEndCarryover`: `execresult`.
pub async fn apply_year_end_carryover<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ApplyYearEndCarryoverParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO hr_leave_balances (\n    owner_id, business_id, employee_id, leave_type_id, year,\n    entitled_days, used_days, pending_days, last_accrual_month\n)\nSELECT previous.owner_id, previous.business_id, previous.employee_id, previous.leave_type_id, ?,\n       LEAST(GREATEST(previous.entitled_days - previous.used_days - previous.pending_days, 0), leave_type.max_carryover_days),\n       0, 0, -1\nFROM hr_leave_balances AS previous\nJOIN hr_leave_types AS leave_type\n  ON leave_type.business_id = previous.business_id\n AND leave_type.id = previous.leave_type_id\nWHERE previous.business_id = ? AND previous.year = ? AND leave_type.is_active = TRUE\nON DUPLICATE KEY UPDATE\n    entitled_days = entitled_days + IF(last_accrual_month = 0, VALUES(entitled_days), 0),\n    last_accrual_month = IF(last_accrual_month = 0, -1, last_accrual_month),\n    updated_at = NOW()")
        .bind(params.year)
        .bind(params.business_id)
        .bind(params.year_2)
        .execute(exec)
        .await
}

/// Params for `ApproveLeaveRequest` (Go `ApproveLeaveRequestParams`).
pub struct ApproveLeaveRequestParams {
    pub id: i64,
    pub business_id: i64,
}

/// `ApproveLeaveRequest`: `execresult`.
pub async fn approve_leave_request<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ApproveLeaveRequestParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE hr_leave_requests\nSET status = 'approved', updated_at = NOW()\nWHERE id = ? AND business_id = ? AND status = 'pending'")
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `CancelLeaveRequest` (Go `CancelLeaveRequestParams`).
pub struct CancelLeaveRequestParams {
    pub id: i64,
    pub business_id: i64,
}

/// `CancelLeaveRequest`: `execresult`.
pub async fn cancel_leave_request<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CancelLeaveRequestParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE hr_leave_requests\nSET status = 'canceled', updated_at = NOW()\nWHERE id = ? AND business_id = ? AND status = 'pending'")
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `CountLeaveBalances` (Go `CountLeaveBalancesParams`).
pub struct CountLeaveBalancesParams {
    pub business_id: i64,
    pub employee_id: Option<i64>,
    pub year: i32,
}

/// `CountLeaveBalances`: `one`.
pub async fn count_leave_balances<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountLeaveBalancesParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) AS total\nFROM hr_leave_balances\nWHERE business_id = ?\n  AND (? IS NULL OR employee_id = ?)\n  AND year = ?")
        .bind(params.business_id)
        .bind(params.employee_id)
        .bind(params.employee_id)
        .bind(params.year)
        .fetch_optional(exec)
        .await
}

/// Params for `CountLeaveRequests` (Go `CountLeaveRequestsParams`).
pub struct CountLeaveRequestsParams {
    pub business_id: i64,
    pub column2: String,
    pub employee_id: i64,
    pub column4: String,
    pub status: String,
    pub from_date: Option<chrono::DateTime<chrono::Utc>>,
    pub to_date: Option<chrono::DateTime<chrono::Utc>>,
}

/// `CountLeaveRequests`: `one`.
pub async fn count_leave_requests<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountLeaveRequestsParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) AS total\nFROM hr_leave_requests\nWHERE business_id = ?\n  AND (? = 0 OR employee_id = ?)\n  AND (? = '' OR status = ?)\n  AND (? IS NULL OR end_date >= ?)\n  AND (? IS NULL OR start_date <= ?)")
        .bind(params.business_id)
        .bind(&params.column2)
        .bind(params.employee_id)
        .bind(&params.column4)
        .bind(&params.status)
        .bind(params.from_date)
        .bind(params.from_date)
        .bind(params.to_date)
        .bind(params.to_date)
        .fetch_optional(exec)
        .await
}

/// Params for `CountLeaveTypes` (Go `CountLeaveTypesParams`).
pub struct CountLeaveTypesParams {
    pub business_id: i64,
}

/// `CountLeaveTypes`: `one`.
pub async fn count_leave_types<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountLeaveTypesParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query(
        "SELECT COUNT(*) AS total FROM hr_leave_types\nWHERE business_id = ? AND is_active = TRUE",
    )
    .bind(params.business_id)
    .fetch_optional(exec)
    .await
}

/// Params for `CreateLeaveRequest` (Go `CreateLeaveRequestParams`).
pub struct CreateLeaveRequestParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub employee_id: i64,
    pub leave_type_id: i64,
    pub start_date: chrono::NaiveDate,
    pub end_date: chrono::NaiveDate,
    pub total_days: String,
    pub notes: Option<String>,
    pub status: String,
    pub current_approval_level: i32,
}

/// `CreateLeaveRequest`: `execresult`.
pub async fn create_leave_request<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateLeaveRequestParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO hr_leave_requests (\n    owner_id, business_id, employee_id, leave_type_id, start_date, end_date, total_days,\n    notes, status, current_approval_level\n) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(params.employee_id)
        .bind(params.leave_type_id)
        .bind(params.start_date)
        .bind(params.end_date)
        .bind(&params.total_days)
        .bind(&params.notes)
        .bind(&params.status)
        .bind(params.current_approval_level)
        .execute(exec)
        .await
}

/// Params for `CreateLeaveType` (Go `CreateLeaveTypeParams`).
pub struct CreateLeaveTypeParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub name: String,
    pub is_paid: bool,
    pub requires_approval: bool,
    pub accrual_rule: Option<String>,
    pub accrual_amount: Option<String>,
    pub max_carryover_days: i32,
    pub min_notice_days: i32,
    pub requires_documentation_after_days: Option<i32>,
    pub is_active: bool,
}

/// `CreateLeaveType`: `execresult`.
pub async fn create_leave_type<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateLeaveTypeParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO hr_leave_types (\n    owner_id, business_id, name, is_paid, requires_approval, accrual_rule, accrual_amount,\n    max_carryover_days, min_notice_days, requires_documentation_after_days, is_active\n) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(&params.name)
        .bind(params.is_paid)
        .bind(params.requires_approval)
        .bind(&params.accrual_rule)
        .bind(&params.accrual_amount)
        .bind(params.max_carryover_days)
        .bind(params.min_notice_days)
        .bind(params.requires_documentation_after_days)
        .bind(params.is_active)
        .execute(exec)
        .await
}

/// Params for `GetLeaveBalance` (Go `GetLeaveBalanceParams`).
pub struct GetLeaveBalanceParams {
    pub business_id: i64,
    pub employee_id: i64,
    pub leave_type_id: i64,
    pub year: i32,
}

/// `GetLeaveBalance`: `one`.
pub async fn get_leave_balance<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetLeaveBalanceParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, employee_id, leave_type_id, year, entitled_days, used_days,\n       pending_days, last_accrual_month, created_at, updated_at\nFROM hr_leave_balances\nWHERE business_id = ? AND employee_id = ? AND leave_type_id = ? AND year = ?")
        .bind(params.business_id)
        .bind(params.employee_id)
        .bind(params.leave_type_id)
        .bind(params.year)
        .fetch_optional(exec)
        .await
}

/// Params for `GetLeaveBalanceForUpdate` (Go `GetLeaveBalanceForUpdateParams`).
pub struct GetLeaveBalanceForUpdateParams {
    pub business_id: i64,
    pub employee_id: i64,
    pub leave_type_id: i64,
    pub year: i32,
}

/// `GetLeaveBalanceForUpdate`: `one`.
pub async fn get_leave_balance_for_update<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetLeaveBalanceForUpdateParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, employee_id, leave_type_id, year, entitled_days, used_days,\n       pending_days, last_accrual_month, created_at, updated_at\nFROM hr_leave_balances\nWHERE business_id = ? AND employee_id = ? AND leave_type_id = ? AND year = ?\nFOR UPDATE")
        .bind(params.business_id)
        .bind(params.employee_id)
        .bind(params.leave_type_id)
        .bind(params.year)
        .fetch_optional(exec)
        .await
}

/// Params for `GetLeaveRequest` (Go `GetLeaveRequestParams`).
pub struct GetLeaveRequestParams {
    pub business_id: i64,
    pub id: i64,
}

/// `GetLeaveRequest`: `one`.
pub async fn get_leave_request<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetLeaveRequestParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, employee_id, leave_type_id, start_date, end_date, total_days,\n       notes, status, current_approval_level, rejection_reason, created_at, updated_at\nFROM hr_leave_requests\nWHERE business_id = ? AND id = ?")
        .bind(params.business_id)
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetLeaveType` (Go `GetLeaveTypeParams`).
pub struct GetLeaveTypeParams {
    pub business_id: i64,
    pub id: i64,
}

/// `GetLeaveType`: `one`.
pub async fn get_leave_type<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetLeaveTypeParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, name, is_paid, requires_approval, accrual_rule, accrual_amount,\n       max_carryover_days, min_notice_days, requires_documentation_after_days,\n       is_active, created_at, updated_at\nFROM hr_leave_types\nWHERE business_id = ? AND id = ?")
        .bind(params.business_id)
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `ListAllActiveLeaveTypesForOwner` (Go `ListAllActiveLeaveTypesForOwnerParams`).
pub struct ListAllActiveLeaveTypesForOwnerParams {
    pub business_id: i64,
}

/// `ListAllActiveLeaveTypesForOwner`: `many`.
pub async fn list_all_active_leave_types_for_owner<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListAllActiveLeaveTypesForOwnerParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, name, is_paid, requires_approval, accrual_rule, accrual_amount,\n       max_carryover_days, min_notice_days, requires_documentation_after_days,\n       is_active, created_at, updated_at\nFROM hr_leave_types\nWHERE business_id = ? AND is_active = TRUE\nLIMIT 500")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `ListBalancesForAccrual` (Go `ListBalancesForAccrualParams`).
pub struct ListBalancesForAccrualParams {
    pub business_id: i64,
    pub year: i32,
    pub last_accrual_month: i32,
}

/// `ListBalancesForAccrual`: `many`.
pub async fn list_balances_for_accrual<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListBalancesForAccrualParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, employee_id, leave_type_id, year, entitled_days, used_days,\n       pending_days, last_accrual_month, created_at, updated_at\nFROM hr_leave_balances\nWHERE business_id = ? AND year = ? AND last_accrual_month < ?\nLIMIT 2000")
        .bind(params.business_id)
        .bind(params.year)
        .bind(params.last_accrual_month)
        .execute(exec)
        .await
}

/// `ListDistinctOwnerIDs`: `many`.
pub async fn list_distinct_owner_i_ds<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query(
        "SELECT DISTINCT business_id FROM hr_leave_types WHERE is_active = TRUE\nLIMIT 5000",
    )
    .execute(exec)
    .await
}

/// Params for `ListLeaveBalances` (Go `ListLeaveBalancesParams`).
pub struct ListLeaveBalancesParams {
    pub business_id: i64,
    pub employee_id: Option<i64>,
    pub year: i32,
}

/// `ListLeaveBalances`: `many`.
pub async fn list_leave_balances<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListLeaveBalancesParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, employee_id, leave_type_id, year, entitled_days, used_days,\n       pending_days, last_accrual_month, created_at, updated_at\nFROM hr_leave_balances\nWHERE business_id = ?\n  AND (? IS NULL OR employee_id = ?)\n  AND year = ?\nORDER BY leave_type_id ASC\nLIMIT 500")
        .bind(params.business_id)
        .bind(params.employee_id)
        .bind(params.employee_id)
        .bind(params.year)
        .execute(exec)
        .await
}

/// Params for `ListLeaveCalendar` (Go `ListLeaveCalendarParams`).
pub struct ListLeaveCalendarParams {
    pub business_id: i64,
    pub date_to: chrono::NaiveDate,
    pub date_from: chrono::NaiveDate,
    pub location_id: Option<i64>,
}

/// `ListLeaveCalendar`: `many`.
pub async fn list_leave_calendar<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListLeaveCalendarParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT lr.id, lr.owner_id, lr.business_id, lr.employee_id, lr.leave_type_id,\n       lr.start_date, lr.end_date, lr.total_days, lr.notes, lr.status,\n       lr.current_approval_level, lr.rejection_reason, lr.created_at, lr.updated_at\nFROM hr_leave_requests AS lr\nLEFT JOIN hr_employees AS employee\n  ON employee.id = lr.employee_id\n AND employee.business_id = lr.business_id\nWHERE lr.business_id = ? AND lr.status IN ('approved', 'pending')\n  AND lr.start_date <= ? AND lr.end_date >= ?\n  AND (? IS NULL OR employee.location_id = ?)\nORDER BY lr.start_date ASC\nLIMIT 1000")
        .bind(params.business_id)
        .bind(params.date_to)
        .bind(params.date_from)
        .bind(params.location_id)
        .bind(params.location_id)
        .execute(exec)
        .await
}

/// Params for `ListLeaveRequests` (Go `ListLeaveRequestsParams`).
pub struct ListLeaveRequestsParams {
    pub business_id: i64,
    pub column2: String,
    pub employee_id: i64,
    pub column4: String,
    pub status: String,
    pub from_date: Option<chrono::DateTime<chrono::Utc>>,
    pub to_date: Option<chrono::DateTime<chrono::Utc>>,
    pub limit: i32,
    pub offset: i32,
}

/// `ListLeaveRequests`: `many`.
pub async fn list_leave_requests<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListLeaveRequestsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, employee_id, leave_type_id, start_date, end_date, total_days,\n       notes, status, current_approval_level, rejection_reason, created_at, updated_at\nFROM hr_leave_requests\nWHERE business_id = ?\n  AND (? = 0 OR employee_id = ?)\n  AND (? = '' OR status = ?)\n  AND (? IS NULL OR end_date >= ?)\n  AND (? IS NULL OR start_date <= ?)\nORDER BY created_at DESC\nLIMIT ? OFFSET ?")
        .bind(params.business_id)
        .bind(&params.column2)
        .bind(params.employee_id)
        .bind(&params.column4)
        .bind(&params.status)
        .bind(params.from_date)
        .bind(params.from_date)
        .bind(params.to_date)
        .bind(params.to_date)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `ListLeaveTypes` (Go `ListLeaveTypesParams`).
pub struct ListLeaveTypesParams {
    pub business_id: i64,
    pub limit: i32,
    pub offset: i32,
}

/// `ListLeaveTypes`: `many`.
pub async fn list_leave_types<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListLeaveTypesParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, name, is_paid, requires_approval, accrual_rule, accrual_amount,\n       max_carryover_days, min_notice_days, requires_documentation_after_days,\n       is_active, created_at, updated_at\nFROM hr_leave_types\nWHERE business_id = ? AND is_active = TRUE\nORDER BY name ASC\nLIMIT ? OFFSET ?")
        .bind(params.business_id)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `ListOverlappingLeaveRequests` (Go `ListOverlappingLeaveRequestsParams`).
pub struct ListOverlappingLeaveRequestsParams {
    pub business_id: i64,
    pub employee_id: i64,
    pub start_date: chrono::NaiveDate,
    pub end_date: chrono::NaiveDate,
}

/// `ListOverlappingLeaveRequests`: `many`.
pub async fn list_overlapping_leave_requests<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListOverlappingLeaveRequestsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, employee_id, leave_type_id, start_date, end_date, total_days,\n       notes, status, current_approval_level, rejection_reason, created_at, updated_at\nFROM hr_leave_requests\nWHERE business_id = ? AND employee_id = ? AND status IN ('pending', 'approved')\n  AND start_date <= ? AND end_date >= ?\nLIMIT 500")
        .bind(params.business_id)
        .bind(params.employee_id)
        .bind(params.start_date)
        .bind(params.end_date)
        .execute(exec)
        .await
}

/// Params for `PatchLeaveType` (Go `PatchLeaveTypeParams`).
pub struct PatchLeaveTypeParams {
    pub name: Option<String>,
    pub is_paid: Option<bool>,
    pub requires_approval: Option<bool>,
    pub accrual_rule: Option<String>,
    pub accrual_amount: Option<String>,
    pub max_carryover_days: Option<i32>,
    pub min_notice_days: Option<i32>,
    pub requires_documentation_after_days: Option<i32>,
    pub is_active: Option<bool>,
    pub id: i64,
    pub business_id: i64,
}

/// `PatchLeaveType`: `execresult`.
pub async fn patch_leave_type<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &PatchLeaveTypeParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE hr_leave_types\nSET name = COALESCE(?, name),\n    is_paid = COALESCE(?, is_paid),\n    requires_approval = COALESCE(?, requires_approval),\n    accrual_rule = COALESCE(?, accrual_rule),\n    accrual_amount = COALESCE(?, accrual_amount),\n    max_carryover_days = COALESCE(?, max_carryover_days),\n    min_notice_days = COALESCE(?, min_notice_days),\n    requires_documentation_after_days = COALESCE(?, requires_documentation_after_days),\n    is_active = COALESCE(?, is_active),\n    updated_at = NOW()\nWHERE id = ? AND business_id = ?")
        .bind(&params.name)
        .bind(params.is_paid)
        .bind(params.requires_approval)
        .bind(&params.accrual_rule)
        .bind(&params.accrual_amount)
        .bind(params.max_carryover_days)
        .bind(params.min_notice_days)
        .bind(params.requires_documentation_after_days)
        .bind(params.is_active)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `RejectLeaveRequest` (Go `RejectLeaveRequestParams`).
pub struct RejectLeaveRequestParams {
    pub rejection_reason: Option<String>,
    pub id: i64,
    pub business_id: i64,
}

/// `RejectLeaveRequest`: `execresult`.
pub async fn reject_leave_request<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &RejectLeaveRequestParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE hr_leave_requests\nSET status = 'rejected', rejection_reason = ?, updated_at = NOW()\nWHERE id = ? AND business_id = ? AND status = 'pending'")
        .bind(&params.rejection_reason)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `SubtractPendingDays` (Go `SubtractPendingDaysParams`).
pub struct SubtractPendingDaysParams {
    pub pending_days: String,
    pub business_id: i64,
    pub employee_id: i64,
    pub leave_type_id: i64,
    pub year: i32,
}

/// `SubtractPendingDays`: `execresult`.
pub async fn subtract_pending_days<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &SubtractPendingDaysParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE hr_leave_balances\nSET pending_days = GREATEST(pending_days - ?, 0), updated_at = NOW()\nWHERE business_id = ? AND employee_id = ? AND leave_type_id = ? AND year = ?")
        .bind(&params.pending_days)
        .bind(params.business_id)
        .bind(params.employee_id)
        .bind(params.leave_type_id)
        .bind(params.year)
        .execute(exec)
        .await
}

/// Params for `UpdateLeaveType` (Go `UpdateLeaveTypeParams`).
pub struct UpdateLeaveTypeParams {
    pub name: String,
    pub is_paid: bool,
    pub requires_approval: bool,
    pub accrual_rule: Option<String>,
    pub accrual_amount: Option<String>,
    pub max_carryover_days: i32,
    pub min_notice_days: i32,
    pub requires_documentation_after_days: Option<i32>,
    pub is_active: bool,
    pub id: i64,
    pub business_id: i64,
}

/// `UpdateLeaveType`: `execresult`.
pub async fn update_leave_type<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateLeaveTypeParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE hr_leave_types\nSET name = ?, is_paid = ?, requires_approval = ?, accrual_rule = ?, accrual_amount = ?,\n    max_carryover_days = ?, min_notice_days = ?, requires_documentation_after_days = ?,\n    is_active = ?, updated_at = NOW()\nWHERE id = ? AND business_id = ?")
        .bind(&params.name)
        .bind(params.is_paid)
        .bind(params.requires_approval)
        .bind(&params.accrual_rule)
        .bind(&params.accrual_amount)
        .bind(params.max_carryover_days)
        .bind(params.min_notice_days)
        .bind(params.requires_documentation_after_days)
        .bind(params.is_active)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `UpsertLeaveBalance` (Go `UpsertLeaveBalanceParams`).
pub struct UpsertLeaveBalanceParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub employee_id: i64,
    pub leave_type_id: i64,
    pub year: i32,
    pub entitled_days: String,
    pub used_days: String,
    pub pending_days: String,
    pub last_accrual_month: i32,
}

/// `UpsertLeaveBalance`: `execresult`.
pub async fn upsert_leave_balance<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpsertLeaveBalanceParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO hr_leave_balances (owner_id, business_id, employee_id, leave_type_id, year, entitled_days, used_days, pending_days, last_accrual_month)\nVALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)\nON DUPLICATE KEY UPDATE entitled_days = VALUES(entitled_days), used_days = VALUES(used_days),\n    pending_days = VALUES(pending_days), last_accrual_month = VALUES(last_accrual_month), updated_at = NOW()")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(params.employee_id)
        .bind(params.leave_type_id)
        .bind(params.year)
        .bind(&params.entitled_days)
        .bind(&params.used_days)
        .bind(&params.pending_days)
        .bind(params.last_accrual_month)
        .execute(exec)
        .await
}
