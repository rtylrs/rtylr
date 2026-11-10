//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `AdvanceApprovalLevel` (Go `AdvanceApprovalLevelParams`).
pub struct AdvanceApprovalLevelParams {
    pub id: i64,
    pub business_id: i64,
    pub current_approval_level: i32,
}

/// `AdvanceApprovalLevel`: `execresult`.
pub async fn advance_approval_level<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &AdvanceApprovalLevelParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE hr_timesheets\nSET current_approval_level = current_approval_level + 1, updated_at = NOW()\nWHERE id = ? AND business_id = ? AND status = 'submitted' AND current_approval_level = ?")
        .bind(params.id)
        .bind(params.business_id)
        .bind(params.current_approval_level)
        .execute(exec)
        .await
}

/// Params for `ApproveTimesheet` (Go `ApproveTimesheetParams`).
pub struct ApproveTimesheetParams {
    pub id: i64,
    pub business_id: i64,
}

/// `ApproveTimesheet`: `exec`.
pub async fn approve_timesheet<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ApproveTimesheetParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE hr_timesheets\nSET status = 'approved', updated_at = NOW()\nWHERE id = ? AND business_id = ?")
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `ApproveTimesheetEntries` (Go `ApproveTimesheetEntriesParams`).
pub struct ApproveTimesheetEntriesParams {
    pub timesheet_id: Option<i64>,
    pub approved_by: Option<i64>,
    pub business_id: i64,
    pub employee_id: i64,
    pub timesheet_id_2: Option<i64>,
    pub date: chrono::NaiveDate,
    pub date_2: chrono::NaiveDate,
}

/// `ApproveTimesheetEntries`: `execrows`.
pub async fn approve_timesheet_entries<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ApproveTimesheetEntriesParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE hr_time_entries\nSET timesheet_id = COALESCE(timesheet_id, ?),\n    status = 'approved',\n    approved_by = ?,\n    rejection_reason = NULL,\n    updated_at = NOW()\nWHERE business_id = ?\n  AND employee_id = ?\n  AND (timesheet_id = ? OR (timesheet_id IS NULL AND date >= ? AND date <= ?))\n  AND status IN ('draft', 'rejected')")
        .bind(params.timesheet_id)
        .bind(params.approved_by)
        .bind(params.business_id)
        .bind(params.employee_id)
        .bind(params.timesheet_id_2)
        .bind(params.date)
        .bind(params.date_2)
        .execute(exec)
        .await
}

/// Params for `CountTimeEntries` (Go `CountTimeEntriesParams`).
pub struct CountTimeEntriesParams {
    pub business_id: i64,
    pub employee_id: Option<i64>,
    pub from_date: Option<chrono::DateTime<chrono::Utc>>,
    pub to_date: Option<chrono::DateTime<chrono::Utc>>,
    pub project_id: Option<i64>,
    pub is_billable: Option<bool>,
    pub status: Option<String>,
}

/// `CountTimeEntries`: `one`.
pub async fn count_time_entries<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountTimeEntriesParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) AS total FROM hr_time_entries\nWHERE business_id = ?\n  AND (? IS NULL OR employee_id = ?)\n  AND (? IS NULL OR date >= ?)\n  AND (? IS NULL OR date <= ?)\n  AND (? IS NULL OR project_id = ?)\n  AND (? IS NULL OR is_billable = ?)\n  AND (? IS NULL OR status = ?)")
        .bind(params.business_id)
        .bind(params.employee_id)
        .bind(params.employee_id)
        .bind(params.from_date)
        .bind(params.from_date)
        .bind(params.to_date)
        .bind(params.to_date)
        .bind(params.project_id)
        .bind(params.project_id)
        .bind(params.is_billable)
        .bind(params.is_billable)
        .bind(&params.status)
        .bind(&params.status)
        .fetch_optional(exec)
        .await
}

/// Params for `CountTimesheets` (Go `CountTimesheetsParams`).
pub struct CountTimesheetsParams {
    pub business_id: i64,
    pub employee_id: Option<i64>,
    pub status: Option<String>,
}

/// `CountTimesheets`: `one`.
pub async fn count_timesheets<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountTimesheetsParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) AS total FROM hr_timesheets\nWHERE business_id = ?\n  AND (? IS NULL OR employee_id = ?)\n  AND (? IS NULL OR status = ?)")
        .bind(params.business_id)
        .bind(params.employee_id)
        .bind(params.employee_id)
        .bind(&params.status)
        .bind(&params.status)
        .fetch_optional(exec)
        .await
}

/// Params for `CountUnbilledBillableTimeEntries` (Go `CountUnbilledBillableTimeEntriesParams`).
pub struct CountUnbilledBillableTimeEntriesParams {
    pub business_id: i64,
}

/// `CountUnbilledBillableTimeEntries`: `one`.
pub async fn count_unbilled_billable_time_entries<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountUnbilledBillableTimeEntriesParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) AS total,\n       CAST(COALESCE(SUM(duration_minutes), 0) AS SIGNED) AS total_minutes\nFROM hr_time_entries\nWHERE business_id = ?\n  AND is_billable = TRUE\n  AND status = 'approved'\n  AND billable_invoice_id IS NULL")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `CreateApprovalAction` (Go `CreateApprovalActionParams`).
pub struct CreateApprovalActionParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub timesheet_id: i64,
    pub approver_id: i64,
    pub action: String,
    pub level: i32,
    pub comment: Option<String>,
}

/// `CreateApprovalAction`: `exec`.
pub async fn create_approval_action<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateApprovalActionParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO hr_timesheet_approval_actions (\n    owner_id, business_id, timesheet_id, approver_id, action, level, comment\n) VALUES (?, ?, ?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(params.timesheet_id)
        .bind(params.approver_id)
        .bind(&params.action)
        .bind(params.level)
        .bind(&params.comment)
        .execute(exec)
        .await
}

/// Params for `CreateFlowWorkdayEntry` (Go `CreateFlowWorkdayEntryParams`).
pub struct CreateFlowWorkdayEntryParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub employee_id: i64,
    pub work_day_id: Option<i64>,
    pub date: chrono::NaiveDate,
    pub project_id: Option<i64>,
    pub task_id: Option<i64>,
    pub work_type_id: Option<i64>,
    pub work_mode: Option<String>,
    pub timezone: Option<String>,
    pub duration_minutes: i32,
    pub description: Option<String>,
    pub is_billable: bool,
    pub client_key: Option<String>,
    pub logged_by: Option<i64>,
}

/// `CreateFlowWorkdayEntry`: `execresult`.
pub async fn create_flow_workday_entry<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateFlowWorkdayEntryParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO hr_time_entries (\n    owner_id, business_id, employee_id, work_day_id, date,\n    project_id, task_id, work_type_id, work_mode, timezone,\n    duration_minutes, description, is_billable, client_key, logged_by, source\n) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'flow')\nON DUPLICATE KEY UPDATE id = LAST_INSERT_ID(id)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(params.employee_id)
        .bind(params.work_day_id)
        .bind(params.date)
        .bind(params.project_id)
        .bind(params.task_id)
        .bind(params.work_type_id)
        .bind(&params.work_mode)
        .bind(&params.timezone)
        .bind(params.duration_minutes)
        .bind(&params.description)
        .bind(params.is_billable)
        .bind(&params.client_key)
        .bind(params.logged_by)
        .execute(exec)
        .await
}

/// Params for `CreateTimeEntry` (Go `CreateTimeEntryParams`).
pub struct CreateTimeEntryParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub employee_id: i64,
    pub timesheet_id: Option<i64>,
    pub project_id: Option<i64>,
    pub task_id: Option<i64>,
    pub category_id: Option<i64>,
    pub date: chrono::NaiveDate,
    pub duration_minutes: i32,
    pub start_time: Option<String>,
    pub end_time: Option<String>,
    pub description: Option<String>,
    pub is_billable: bool,
    pub status: String,
    pub rejection_reason: Option<String>,
    pub approved_by: Option<i64>,
    pub source: String,
}

/// `CreateTimeEntry`: `execresult`.
pub async fn create_time_entry<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateTimeEntryParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO hr_time_entries (\n    owner_id, business_id, employee_id, timesheet_id, project_id, task_id, category_id,\n    date, duration_minutes, start_time, end_time, description, is_billable,\n    status, rejection_reason, approved_by, source\n) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(params.employee_id)
        .bind(params.timesheet_id)
        .bind(params.project_id)
        .bind(params.task_id)
        .bind(params.category_id)
        .bind(params.date)
        .bind(params.duration_minutes)
        .bind(&params.start_time)
        .bind(&params.end_time)
        .bind(&params.description)
        .bind(params.is_billable)
        .bind(&params.status)
        .bind(&params.rejection_reason)
        .bind(params.approved_by)
        .bind(&params.source)
        .execute(exec)
        .await
}

/// Params for `CreateTimesheet` (Go `CreateTimesheetParams`).
pub struct CreateTimesheetParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub employee_id: i64,
    pub period_start: chrono::NaiveDate,
    pub period_end: chrono::NaiveDate,
    pub status: String,
    pub current_approval_level: i32,
    pub total_minutes: i32,
}

/// `CreateTimesheet`: `execresult`.
pub async fn create_timesheet<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateTimesheetParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO hr_timesheets (\n    owner_id, business_id, employee_id, period_start, period_end, status,\n    current_approval_level, total_minutes\n) VALUES (?, ?, ?, ?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(params.employee_id)
        .bind(params.period_start)
        .bind(params.period_end)
        .bind(&params.status)
        .bind(params.current_approval_level)
        .bind(params.total_minutes)
        .execute(exec)
        .await
}

/// Params for `DeleteFlowWorkdayEntry` (Go `DeleteFlowWorkdayEntryParams`).
pub struct DeleteFlowWorkdayEntryParams {
    pub id: i64,
    pub owner_id: i64,
    pub business_id: i64,
}

/// `DeleteFlowWorkdayEntry`: `execresult`.
pub async fn delete_flow_workday_entry<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DeleteFlowWorkdayEntryParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("DELETE FROM hr_time_entries\nWHERE id = ? AND owner_id = ? AND business_id = ? AND source = 'flow'")
        .bind(params.id)
        .bind(params.owner_id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `DeleteTimeEntry` (Go `DeleteTimeEntryParams`).
pub struct DeleteTimeEntryParams {
    pub id: i64,
    pub business_id: i64,
}

/// `DeleteTimeEntry`: `exec`.
pub async fn delete_time_entry<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DeleteTimeEntryParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("DELETE FROM hr_time_entries WHERE id = ? AND business_id = ?")
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `GetFlowWorkdayEntry` (Go `GetFlowWorkdayEntryParams`).
pub struct GetFlowWorkdayEntryParams {
    pub id: i64,
    pub owner_id: i64,
    pub business_id: i64,
}

/// `GetFlowWorkdayEntry`: `one`.
pub async fn get_flow_workday_entry<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetFlowWorkdayEntryParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, work_day_id, project_id, task_id, work_type_id,\n       work_mode, description, duration_minutes, is_billable, status,\n       client_key, logged_by, created_at, updated_at\nFROM hr_time_entries\nWHERE id = ? AND owner_id = ? AND business_id = ? AND source = 'flow'\nLIMIT 1")
        .bind(params.id)
        .bind(params.owner_id)
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetFlowWorkdayEntryOwner` (Go `GetFlowWorkdayEntryOwnerParams`).
pub struct GetFlowWorkdayEntryOwnerParams {
    pub id: i64,
    pub owner_id: i64,
    pub business_id: i64,
}

/// `GetFlowWorkdayEntryOwner`: `one`.
pub async fn get_flow_workday_entry_owner<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetFlowWorkdayEntryOwnerParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT work_day_id, employee_id, task_id\nFROM hr_time_entries\nWHERE id = ? AND owner_id = ? AND business_id = ? AND source = 'flow'\nLIMIT 1")
        .bind(params.id)
        .bind(params.owner_id)
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetTimeEntry` (Go `GetTimeEntryParams`).
pub struct GetTimeEntryParams {
    pub business_id: i64,
    pub id: i64,
}

/// `GetTimeEntry`: `one`.
pub async fn get_time_entry<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetTimeEntryParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, employee_id, timesheet_id, project_id, task_id, category_id,\n       date, duration_minutes, start_time, end_time, description, is_billable,\n       status, rejection_reason, approved_by, source, billable_invoice_id,\n       created_at, updated_at\nFROM hr_time_entries\nWHERE business_id = ? AND id = ?")
        .bind(params.business_id)
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetTimesheet` (Go `GetTimesheetParams`).
pub struct GetTimesheetParams {
    pub business_id: i64,
    pub id: i64,
}

/// `GetTimesheet`: `one`.
pub async fn get_timesheet<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetTimesheetParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, employee_id, period_start, period_end, status,\n       current_approval_level, submitted_at, total_minutes, rejection_reason,\n       created_at, updated_at\nFROM hr_timesheets\nWHERE business_id = ? AND id = ?")
        .bind(params.business_id)
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetTimesheetForPeriod` (Go `GetTimesheetForPeriodParams`).
pub struct GetTimesheetForPeriodParams {
    pub business_id: i64,
    pub employee_id: i64,
    pub period_start: chrono::NaiveDate,
}

/// `GetTimesheetForPeriod`: `one`.
pub async fn get_timesheet_for_period<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetTimesheetForPeriodParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, employee_id, period_start, period_end, status,\n       current_approval_level, submitted_at, total_minutes, rejection_reason,\n       created_at, updated_at\nFROM hr_timesheets\nWHERE business_id = ? AND employee_id = ? AND period_start = ?")
        .bind(params.business_id)
        .bind(params.employee_id)
        .bind(params.period_start)
        .fetch_optional(exec)
        .await
}

/// Params for `ListFlowWorkdayEntries` (Go `ListFlowWorkdayEntriesParams`).
pub struct ListFlowWorkdayEntriesParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub work_day_id: Option<i64>,
}

/// `ListFlowWorkdayEntries`: `many`.
pub async fn list_flow_workday_entries<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListFlowWorkdayEntriesParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, work_day_id, project_id, task_id, work_type_id,\n       work_mode, description, duration_minutes, is_billable, status,\n       client_key, logged_by, created_at, updated_at\nFROM hr_time_entries\nWHERE owner_id = ? AND business_id = ? AND work_day_id = ? AND source = 'flow'\nORDER BY id ASC")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(params.work_day_id)
        .execute(exec)
        .await
}

/// Params for `ListTimeEntries` (Go `ListTimeEntriesParams`).
pub struct ListTimeEntriesParams {
    pub business_id: i64,
    pub employee_id: Option<i64>,
    pub from_date: Option<chrono::DateTime<chrono::Utc>>,
    pub to_date: Option<chrono::DateTime<chrono::Utc>>,
    pub project_id: Option<i64>,
    pub is_billable: Option<bool>,
    pub status: Option<String>,
    pub limit: i32,
    pub offset: i32,
}

/// `ListTimeEntries`: `many`.
pub async fn list_time_entries<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListTimeEntriesParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, employee_id, timesheet_id, project_id, task_id, category_id,\n       date, duration_minutes, start_time, end_time, description, is_billable,\n       status, rejection_reason, approved_by, source, billable_invoice_id,\n       created_at, updated_at\nFROM hr_time_entries\nWHERE business_id = ?\n  AND (? IS NULL OR employee_id = ?)\n  AND (? IS NULL OR date >= ?)\n  AND (? IS NULL OR date <= ?)\n  AND (? IS NULL OR project_id = ?)\n  AND (? IS NULL OR is_billable = ?)\n  AND (? IS NULL OR status = ?)\nORDER BY date DESC, created_at DESC\nLIMIT ? OFFSET ?")
        .bind(params.business_id)
        .bind(params.employee_id)
        .bind(params.employee_id)
        .bind(params.from_date)
        .bind(params.from_date)
        .bind(params.to_date)
        .bind(params.to_date)
        .bind(params.project_id)
        .bind(params.project_id)
        .bind(params.is_billable)
        .bind(params.is_billable)
        .bind(&params.status)
        .bind(&params.status)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// `ListTimesheetBusinessIDs`: `many`.
pub async fn list_timesheet_business_i_ds<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT DISTINCT business_id FROM hr_time_entries\nUNION\nSELECT DISTINCT business_id FROM hr_timesheets\nORDER BY business_id ASC\nLIMIT 500")
        .execute(exec)
        .await
}

/// Params for `ListTimesheets` (Go `ListTimesheetsParams`).
pub struct ListTimesheetsParams {
    pub business_id: i64,
    pub employee_id: Option<i64>,
    pub status: Option<String>,
    pub limit: i32,
    pub offset: i32,
}

/// `ListTimesheets`: `many`.
pub async fn list_timesheets<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListTimesheetsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, employee_id, period_start, period_end, status,\n       current_approval_level, submitted_at, total_minutes, rejection_reason,\n       created_at, updated_at\nFROM hr_timesheets\nWHERE business_id = ?\n  AND (? IS NULL OR employee_id = ?)\n  AND (? IS NULL OR status = ?)\nORDER BY period_start DESC\nLIMIT ? OFFSET ?")
        .bind(params.business_id)
        .bind(params.employee_id)
        .bind(params.employee_id)
        .bind(&params.status)
        .bind(&params.status)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `ListUnbilledBillableTimeEntries` (Go `ListUnbilledBillableTimeEntriesParams`).
pub struct ListUnbilledBillableTimeEntriesParams {
    pub business_id: i64,
    pub limit: i32,
    pub offset: i32,
}

/// `ListUnbilledBillableTimeEntries`: `many`.
pub async fn list_unbilled_billable_time_entries<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListUnbilledBillableTimeEntriesParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, employee_id, timesheet_id, project_id, task_id, category_id,\n       date, duration_minutes, start_time, end_time, description, is_billable,\n       status, rejection_reason, approved_by, source, billable_invoice_id,\n       created_at, updated_at\nFROM hr_time_entries\nWHERE business_id = ?\n  AND is_billable = TRUE\n  AND status = 'approved'\n  AND billable_invoice_id IS NULL\nORDER BY date ASC, id ASC\nLIMIT ? OFFSET ?")
        .bind(params.business_id)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `ListUnsubmittedEmployees` (Go `ListUnsubmittedEmployeesParams`).
pub struct ListUnsubmittedEmployeesParams {
    pub period_start: chrono::NaiveDate,
    pub business_id: i64,
    pub date: chrono::NaiveDate,
    pub date_2: chrono::NaiveDate,
}

/// `ListUnsubmittedEmployees`: `many`.
pub async fn list_unsubmitted_employees<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListUnsubmittedEmployeesParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT DISTINCT e.employee_id\nFROM hr_time_entries e\nLEFT JOIN hr_timesheets t ON t.business_id = e.business_id\n    AND t.employee_id = e.employee_id\n    AND t.period_start = ?\nWHERE e.business_id = ?\n    AND e.date >= ? AND e.date <= ?\n    AND (t.id IS NULL OR t.status = 'draft')\nORDER BY e.employee_id ASC\nLIMIT 500")
        .bind(params.period_start)
        .bind(params.business_id)
        .bind(params.date)
        .bind(params.date_2)
        .execute(exec)
        .await
}

/// Params for `LockTimesheet` (Go `LockTimesheetParams`).
pub struct LockTimesheetParams {
    pub id: i64,
    pub business_id: i64,
}

/// `LockTimesheet`: `execresult`.
pub async fn lock_timesheet<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &LockTimesheetParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE hr_timesheets\nSET status = 'locked', updated_at = NOW()\nWHERE id = ? AND business_id = ? AND status = 'submitted'")
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `MarkEntryInvoiced` (Go `MarkEntryInvoicedParams`).
pub struct MarkEntryInvoicedParams {
    pub billable_invoice_id: Option<i64>,
    pub business_id: i64,
    pub id: i64,
}

/// `MarkEntryInvoiced`: `execrows`.
pub async fn mark_entry_invoiced<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &MarkEntryInvoicedParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE hr_time_entries\nSET billable_invoice_id = ?, updated_at = NOW()\nWHERE business_id = ?\n  AND id = ?\n  AND is_billable = TRUE\n  AND status = 'approved'\n  AND billable_invoice_id IS NULL")
        .bind(params.billable_invoice_id)
        .bind(params.business_id)
        .bind(params.id)
        .execute(exec)
        .await
}

/// Params for `RejectTimesheet` (Go `RejectTimesheetParams`).
pub struct RejectTimesheetParams {
    pub rejection_reason: Option<String>,
    pub id: i64,
    pub business_id: i64,
}

/// `RejectTimesheet`: `execrows`.
pub async fn reject_timesheet<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &RejectTimesheetParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE hr_timesheets\nSET status = 'rejected', rejection_reason = ?, updated_at = NOW()\nWHERE id = ? AND business_id = ? AND status = 'submitted'")
        .bind(&params.rejection_reason)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `SubmitTimesheet` (Go `SubmitTimesheetParams`).
pub struct SubmitTimesheetParams {
    pub id: i64,
    pub business_id: i64,
}

/// `SubmitTimesheet`: `execresult`.
pub async fn submit_timesheet<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &SubmitTimesheetParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE hr_timesheets\nSET status = 'submitted', submitted_at = NOW(), current_approval_level = 1, updated_at = NOW()\nWHERE id = ? AND business_id = ? AND status IN ('draft', 'rejected')")
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `SumFlowWorkdayEntryMinutes` (Go `SumFlowWorkdayEntryMinutesParams`).
pub struct SumFlowWorkdayEntryMinutesParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub work_day_id: Option<i64>,
}

/// `SumFlowWorkdayEntryMinutes`: `one`.
pub async fn sum_flow_workday_entry_minutes<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &SumFlowWorkdayEntryMinutesParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COALESCE(SUM(duration_minutes), 0)\nFROM hr_time_entries\nWHERE owner_id = ? AND business_id = ? AND work_day_id = ? AND source = 'flow'")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(params.work_day_id)
        .fetch_optional(exec)
        .await
}

/// Params for `SumFlowWorkdayEntryMinutesByDayIDs` (Go `SumFlowWorkdayEntryMinutesByDayIDsParams`).
pub struct SumFlowWorkdayEntryMinutesByDayIDsParams {
    pub business_id: i64,
    pub work_day_ids: Vec<Option<i64>>,
}

/// `SumFlowWorkdayEntryMinutesByDayIDs`: `many`.
pub async fn sum_flow_workday_entry_minutes_by_day_i_ds<
    'e,
    E: sqlx::Executor<'e, Database = MySql>,
>(
    exec: E,
    params: &SumFlowWorkdayEntryMinutesByDayIDsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    let mut sql_template: String = "SELECT work_day_id, COALESCE(SUM(duration_minutes), 0) AS logged_minutes\nFROM hr_time_entries\nWHERE business_id = ? AND source = 'flow'\n  AND work_day_id IN ({SLICE:work_day_ids})\nGROUP BY work_day_id".to_string();

    if params.work_day_ids.is_empty() {
        sql_template = sql_template.replace("{SLICE:work_day_ids}", "NULL");
    } else {
        sql_template = sql_template.replace(
            "{SLICE:work_day_ids}",
            &std::iter::repeat_n("?", params.work_day_ids.len())
                .collect::<Vec<_>>()
                .join(","),
        );
    }
    let query = sqlx::query(&sql_template);
    let mut query = query.bind(params.business_id);
    for value in &params.work_day_ids {
        query = query.bind(value);
    }
    query.execute(exec).await
}

/// Params for `SumFlowWorkdayEntryMinutesByMemberAndTaskIDs` (Go `SumFlowWorkdayEntryMinutesByMemberAndTaskIDsParams`).
pub struct SumFlowWorkdayEntryMinutesByMemberAndTaskIDsParams {
    pub business_id: i64,
    pub task_ids: Vec<Option<i64>>,
}

/// `SumFlowWorkdayEntryMinutesByMemberAndTaskIDs`: `many`.
pub async fn sum_flow_workday_entry_minutes_by_member_and_task_i_ds<
    'e,
    E: sqlx::Executor<'e, Database = MySql>,
>(
    exec: E,
    params: &SumFlowWorkdayEntryMinutesByMemberAndTaskIDsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    let mut sql_template: String = "SELECT employee_id, CAST(COALESCE(SUM(duration_minutes), 0) AS SIGNED) AS logged_minutes\nFROM hr_time_entries\nWHERE business_id = ? AND source = 'flow' AND task_id IS NOT NULL\n  AND task_id IN ({SLICE:task_ids})\nGROUP BY employee_id".to_string();

    if params.task_ids.is_empty() {
        sql_template = sql_template.replace("{SLICE:task_ids}", "NULL");
    } else {
        sql_template = sql_template.replace(
            "{SLICE:task_ids}",
            &std::iter::repeat_n("?", params.task_ids.len())
                .collect::<Vec<_>>()
                .join(","),
        );
    }
    let query = sqlx::query(&sql_template);
    let mut query = query.bind(params.business_id);
    for value in &params.task_ids {
        query = query.bind(value);
    }
    query.execute(exec).await
}

/// Params for `SumFlowWorkdayEntryMinutesByTaskIDs` (Go `SumFlowWorkdayEntryMinutesByTaskIDsParams`).
pub struct SumFlowWorkdayEntryMinutesByTaskIDsParams {
    pub business_id: i64,
    pub task_ids: Vec<Option<i64>>,
}

/// `SumFlowWorkdayEntryMinutesByTaskIDs`: `many`.
pub async fn sum_flow_workday_entry_minutes_by_task_i_ds<
    'e,
    E: sqlx::Executor<'e, Database = MySql>,
>(
    exec: E,
    params: &SumFlowWorkdayEntryMinutesByTaskIDsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    let mut sql_template: String = "SELECT task_id, CAST(COALESCE(SUM(duration_minutes), 0) AS SIGNED) AS logged_minutes\nFROM hr_time_entries\nWHERE business_id = ? AND source = 'flow' AND task_id IS NOT NULL\n  AND task_id IN ({SLICE:task_ids})\nGROUP BY task_id".to_string();

    if params.task_ids.is_empty() {
        sql_template = sql_template.replace("{SLICE:task_ids}", "NULL");
    } else {
        sql_template = sql_template.replace(
            "{SLICE:task_ids}",
            &std::iter::repeat_n("?", params.task_ids.len())
                .collect::<Vec<_>>()
                .join(","),
        );
    }
    let query = sqlx::query(&sql_template);
    let mut query = query.bind(params.business_id);
    for value in &params.task_ids {
        query = query.bind(value);
    }
    query.execute(exec).await
}

/// Params for `UpdateFlowWorkdayEntry` (Go `UpdateFlowWorkdayEntryParams`).
pub struct UpdateFlowWorkdayEntryParams {
    pub project_id: Option<i64>,
    pub task_id: Option<i64>,
    pub work_type_id: Option<i64>,
    pub work_mode: Option<String>,
    pub duration_minutes: i32,
    pub description: Option<String>,
    pub is_billable: bool,
    pub id: i64,
    pub owner_id: i64,
    pub business_id: i64,
}

/// `UpdateFlowWorkdayEntry`: `execresult`.
pub async fn update_flow_workday_entry<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateFlowWorkdayEntryParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE hr_time_entries\nSET project_id = ?, task_id = ?, work_type_id = ?, work_mode = ?,\n    duration_minutes = ?, description = ?, is_billable = ?\nWHERE id = ? AND owner_id = ? AND business_id = ? AND source = 'flow'")
        .bind(params.project_id)
        .bind(params.task_id)
        .bind(params.work_type_id)
        .bind(&params.work_mode)
        .bind(params.duration_minutes)
        .bind(&params.description)
        .bind(params.is_billable)
        .bind(params.id)
        .bind(params.owner_id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `UpdateTimeEntry` (Go `UpdateTimeEntryParams`).
pub struct UpdateTimeEntryParams {
    pub employee_id: i64,
    pub timesheet_id: Option<i64>,
    pub project_id: Option<i64>,
    pub task_id: Option<i64>,
    pub category_id: Option<i64>,
    pub date: chrono::NaiveDate,
    pub duration_minutes: i32,
    pub start_time: Option<String>,
    pub end_time: Option<String>,
    pub description: Option<String>,
    pub is_billable: bool,
    pub status: String,
    pub rejection_reason: Option<String>,
    pub approved_by: Option<i64>,
    pub source: String,
    pub id: i64,
    pub business_id: i64,
}

/// `UpdateTimeEntry`: `exec`.
pub async fn update_time_entry<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateTimeEntryParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE hr_time_entries\nSET employee_id = ?, timesheet_id = ?, project_id = ?, task_id = ?, category_id = ?,\n    date = ?, duration_minutes = ?, start_time = ?, end_time = ?, description = ?,\n    is_billable = ?, status = ?, rejection_reason = ?, approved_by = ?, source = ?,\n    updated_at = NOW()\nWHERE id = ? AND business_id = ?")
        .bind(params.employee_id)
        .bind(params.timesheet_id)
        .bind(params.project_id)
        .bind(params.task_id)
        .bind(params.category_id)
        .bind(params.date)
        .bind(params.duration_minutes)
        .bind(&params.start_time)
        .bind(&params.end_time)
        .bind(&params.description)
        .bind(params.is_billable)
        .bind(&params.status)
        .bind(&params.rejection_reason)
        .bind(params.approved_by)
        .bind(&params.source)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}
