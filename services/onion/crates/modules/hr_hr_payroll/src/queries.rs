//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `ClaimPayrollRunForExport` (Go `ClaimPayrollRunForExportParams`).
pub struct ClaimPayrollRunForExportParams {
    pub business_id: i64,
    pub id: i64,
}

/// `ClaimPayrollRunForExport`: `execresult`.
pub async fn claim_payroll_run_for_export<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ClaimPayrollRunForExportParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE hr_payroll_runs\nSET status = 'exporting'\nWHERE business_id = ? AND id = ? AND status = 'approved' AND exported_at IS NULL")
        .bind(params.business_id)
        .bind(params.id)
        .execute(exec)
        .await
}

/// Params for `CountInactivePayrollEmployees` (Go `CountInactivePayrollEmployeesParams`).
pub struct CountInactivePayrollEmployeesParams {
    pub business_id: i64,
    pub payroll_run_id: i64,
}

/// `CountInactivePayrollEmployees`: `one`.
pub async fn count_inactive_payroll_employees<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountInactivePayrollEmployeesParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) AS inactive_count\nFROM hr_payroll_entries e\nLEFT JOIN hr_employees emp\n  ON emp.business_id = e.business_id AND emp.id = e.employee_id\nWHERE e.business_id = ?\n  AND e.payroll_run_id = ?\n  AND (emp.id IS NULL OR emp.status <> 'active')")
        .bind(params.business_id)
        .bind(params.payroll_run_id)
        .fetch_optional(exec)
        .await
}

/// Params for `CountPayrollRuns` (Go `CountPayrollRunsParams`).
pub struct CountPayrollRunsParams {
    pub business_id: i64,
}

/// `CountPayrollRuns`: `one`.
pub async fn count_payroll_runs<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountPayrollRunsParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) FROM hr_payroll_runs WHERE business_id = ?")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `CountPayrollRunsForEmployee` (Go `CountPayrollRunsForEmployeeParams`).
pub struct CountPayrollRunsForEmployeeParams {
    pub business_id: i64,
    pub employee_id: i64,
    pub status: String,
    pub status_2: String,
}

/// `CountPayrollRunsForEmployee`: `one`.
pub async fn count_payroll_runs_for_employee<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountPayrollRunsForEmployeeParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) FROM hr_payroll_runs r\nJOIN hr_payroll_entries e ON e.payroll_run_id = r.id AND e.business_id = r.business_id\nWHERE r.business_id = ? AND e.employee_id = ? AND r.status IN (?, ?)")
        .bind(params.business_id)
        .bind(params.employee_id)
        .bind(&params.status)
        .bind(&params.status_2)
        .fetch_optional(exec)
        .await
}

/// Params for `CreatePayrollEntry` (Go `CreatePayrollEntryParams`).
pub struct CreatePayrollEntryParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub payroll_run_id: i64,
    pub employee_id: i64,
    pub hours_worked: String,
    pub overtime_hours: String,
    pub base_pay: String,
    pub overtime_pay: String,
    pub bonuses: String,
    pub deductions: String,
    pub net_pay: String,
}

/// `CreatePayrollEntry`: `execresult`.
pub async fn create_payroll_entry<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreatePayrollEntryParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO hr_payroll_entries (owner_id, business_id, payroll_run_id, employee_id, hours_worked, overtime_hours, base_pay, overtime_pay, bonuses, deductions, net_pay)\nVALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(params.payroll_run_id)
        .bind(params.employee_id)
        .bind(&params.hours_worked)
        .bind(&params.overtime_hours)
        .bind(&params.base_pay)
        .bind(&params.overtime_pay)
        .bind(&params.bonuses)
        .bind(&params.deductions)
        .bind(&params.net_pay)
        .execute(exec)
        .await
}

/// Params for `CreatePayrollRun` (Go `CreatePayrollRunParams`).
pub struct CreatePayrollRunParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub period_start: chrono::NaiveDate,
    pub period_end: chrono::NaiveDate,
    pub status: String,
    pub created_by: i64,
    pub total_gross: String,
    pub total_net: String,
}

/// `CreatePayrollRun`: `execresult`.
pub async fn create_payroll_run<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreatePayrollRunParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO hr_payroll_runs (owner_id, business_id, period_start, period_end, status, created_by, total_gross, total_net)\nVALUES (?, ?, ?, ?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(params.period_start)
        .bind(params.period_end)
        .bind(&params.status)
        .bind(params.created_by)
        .bind(&params.total_gross)
        .bind(&params.total_net)
        .execute(exec)
        .await
}

/// Params for `GetOvertimeConfigForPayroll` (Go `GetOvertimeConfigForPayrollParams`).
pub struct GetOvertimeConfigForPayrollParams {
    pub business_id: i64,
}

/// `GetOvertimeConfigForPayroll`: `one`.
pub async fn get_overtime_config_for_payroll<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetOvertimeConfigForPayrollParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT default_multiplier, auto_detect_enabled\nFROM hr_overtime_config\nWHERE business_id = ?")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetPayrollConfig` (Go `GetPayrollConfigParams`).
pub struct GetPayrollConfigParams {
    pub business_id: i64,
}

/// `GetPayrollConfig`: `one`.
pub async fn get_payroll_config<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetPayrollConfigParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, period_type, pay_day_of_month, overtime_enabled, created_at, updated_at\nFROM hr_payroll_config\nWHERE business_id = ?")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetPayrollEntryForEmployee` (Go `GetPayrollEntryForEmployeeParams`).
pub struct GetPayrollEntryForEmployeeParams {
    pub business_id: i64,
    pub payroll_run_id: i64,
    pub employee_id: i64,
}

/// `GetPayrollEntryForEmployee`: `one`.
pub async fn get_payroll_entry_for_employee<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetPayrollEntryForEmployeeParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, payroll_run_id, employee_id, hours_worked, overtime_hours, base_pay, overtime_pay, bonuses, deductions, net_pay, created_at, updated_at\nFROM hr_payroll_entries\nWHERE business_id = ? AND payroll_run_id = ? AND employee_id = ?")
        .bind(params.business_id)
        .bind(params.payroll_run_id)
        .bind(params.employee_id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetPayrollRun` (Go `GetPayrollRunParams`).
pub struct GetPayrollRunParams {
    pub business_id: i64,
    pub id: i64,
}

/// `GetPayrollRun`: `one`.
pub async fn get_payroll_run<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetPayrollRunParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, period_start, period_end, status, created_by, approved_by, exported_at, total_gross, total_net, created_at, updated_at\nFROM hr_payroll_runs\nWHERE business_id = ? AND id = ?")
        .bind(params.business_id)
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetPayrollRunByPeriod` (Go `GetPayrollRunByPeriodParams`).
pub struct GetPayrollRunByPeriodParams {
    pub business_id: i64,
    pub period_start: chrono::NaiveDate,
    pub period_end: chrono::NaiveDate,
}

/// `GetPayrollRunByPeriod`: `one`.
pub async fn get_payroll_run_by_period<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetPayrollRunByPeriodParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, period_start, period_end, status, created_by, approved_by, exported_at, total_gross, total_net, created_at, updated_at\nFROM hr_payroll_runs\nWHERE business_id = ? AND period_start = ? AND period_end = ?\nORDER BY id ASC\nLIMIT 1")
        .bind(params.business_id)
        .bind(params.period_start)
        .bind(params.period_end)
        .fetch_optional(exec)
        .await
}

/// Params for `GetProjectLabourCost` (Go `GetProjectLabourCostParams`).
pub struct GetProjectLabourCostParams {
    pub business_id: i64,
    pub business_id_2: i64,
    pub project_id: Option<i64>,
}

/// `GetProjectLabourCost`: `one`.
pub async fn get_project_labour_cost<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetProjectLabourCostParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT\n    e.business_id,\n    e.project_id,\n    COALESCE(SUM(CASE\n        WHEN emp.hourly_rate IS NOT NULL\n        THEN ROUND(emp.hourly_rate * e.duration_minutes / 60 * 100)\n        ELSE 0\n    END), 0) AS cost_cents,\n    COALESCE(SUM(CASE\n        WHEN emp.hourly_rate IS NOT NULL THEN e.duration_minutes\n        ELSE 0\n    END), 0) AS costed_minutes,\n    COALESCE(SUM(CASE\n        WHEN emp.hourly_rate IS NULL THEN e.duration_minutes\n        ELSE 0\n    END), 0) AS uncosted_minutes,\n    COUNT(DISTINCT CASE\n        WHEN emp.hourly_rate IS NULL THEN e.employee_id\n        ELSE NULL\n    END) AS uncosted_employee_count\nFROM hr_time_entries e\nJOIN hr_employees emp\n  ON emp.id = e.employee_id\n AND emp.business_id = ?\nWHERE e.business_id = ?\n  AND e.project_id = ?\n  AND e.is_billable = TRUE\nGROUP BY e.business_id, e.project_id")
        .bind(params.business_id)
        .bind(params.business_id_2)
        .bind(params.project_id)
        .fetch_optional(exec)
        .await
}

/// `ListAllPayrollConfigs`: `many`.
pub async fn list_all_payroll_configs<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, period_type, pay_day_of_month, overtime_enabled, created_at, updated_at\nFROM hr_payroll_config\nORDER BY business_id\nLIMIT 5000")
        .execute(exec)
        .await
}

/// Params for `ListApprovedOvertimeSummaries` (Go `ListApprovedOvertimeSummariesParams`).
pub struct ListApprovedOvertimeSummariesParams {
    pub business_id: i64,
    pub date: chrono::NaiveDate,
    pub date_2: chrono::NaiveDate,
}

/// `ListApprovedOvertimeSummaries`: `many`.
pub async fn list_approved_overtime_summaries<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListApprovedOvertimeSummariesParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT employee_id,\n       SUM(planned_extra_minutes) AS total_minutes,\n       SUM(planned_extra_minutes * COALESCE(multiplier_applied, 0)) AS weighted_minutes,\n       SUM(CASE WHEN multiplier_applied IS NULL THEN planned_extra_minutes ELSE 0 END) AS unweighted_minutes\nFROM hr_overtime_requests\nWHERE business_id = ? AND status = 'approved' AND date >= ? AND date <= ?\nGROUP BY employee_id\nORDER BY employee_id\nLIMIT 5001")
        .bind(params.business_id)
        .bind(params.date)
        .bind(params.date_2)
        .execute(exec)
        .await
}

/// Params for `ListApprovedTimesheetSummaries` (Go `ListApprovedTimesheetSummariesParams`).
pub struct ListApprovedTimesheetSummariesParams {
    pub period_end: chrono::NaiveDate,
    pub period_start: chrono::NaiveDate,
    pub business_id: i64,
}

/// `ListApprovedTimesheetSummaries`: `many`.
pub async fn list_approved_timesheet_summaries<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListApprovedTimesheetSummariesParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT t.employee_id,\n       ROUND(SUM(\n           (\n               SELECT COALESCE(SUM(e.duration_minutes), 0)\n               FROM hr_time_entries e\n               WHERE e.business_id = t.business_id\n                 AND e.employee_id = t.employee_id\n                 AND e.date >= t.period_start\n                 AND e.date <= t.period_end\n           ) *\n           (DATEDIFF(\n                LEAST(t.period_end, ?),\n                GREATEST(t.period_start, ?)\n            ) + 1) /\n           (DATEDIFF(t.period_end, t.period_start) + 1)\n       ), 0) AS total_minutes\nFROM hr_timesheets t\nWHERE t.business_id = ?\n  AND t.status IN ('approved', 'locked')\n  AND t.period_start <= ?\n  AND t.period_end >= ?\nGROUP BY t.employee_id\nORDER BY t.employee_id\nLIMIT 5001")
        .bind(params.period_end)
        .bind(params.period_start)
        .bind(params.business_id)
        .bind(params.period_end)
        .bind(params.period_start)
        .execute(exec)
        .await
}

/// `ListDistinctPayrollOwnerIDs`: `many`.
pub async fn list_distinct_payroll_owner_i_ds<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query(
        "SELECT DISTINCT business_id FROM hr_payroll_config\nORDER BY business_id\nLIMIT 5000",
    )
    .execute(exec)
    .await
}

/// Params for `ListEmployeePayInfo` (Go `ListEmployeePayInfoParams`).
pub struct ListEmployeePayInfoParams {
    pub business_id: i64,
}

/// `ListEmployeePayInfo`: `many`.
pub async fn list_employee_pay_info<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListEmployeePayInfoParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, salary, hourly_rate\nFROM hr_employees\nWHERE business_id = ? AND status = 'active'\nORDER BY id\nLIMIT 5001")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `ListPayrollEntries` (Go `ListPayrollEntriesParams`).
pub struct ListPayrollEntriesParams {
    pub business_id: i64,
    pub payroll_run_id: i64,
}

/// `ListPayrollEntries`: `many`.
pub async fn list_payroll_entries<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListPayrollEntriesParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, payroll_run_id, employee_id, hours_worked, overtime_hours, base_pay, overtime_pay, bonuses, deductions, net_pay, created_at, updated_at\nFROM hr_payroll_entries\nWHERE business_id = ? AND payroll_run_id = ?\nORDER BY employee_id\nLIMIT 5001")
        .bind(params.business_id)
        .bind(params.payroll_run_id)
        .execute(exec)
        .await
}

/// Params for `ListPayrollRuns` (Go `ListPayrollRunsParams`).
pub struct ListPayrollRunsParams {
    pub business_id: i64,
    pub limit: i32,
    pub offset: i32,
}

/// `ListPayrollRuns`: `many`.
pub async fn list_payroll_runs<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListPayrollRunsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, period_start, period_end, status, created_by, approved_by, exported_at, total_gross, total_net, created_at, updated_at\nFROM hr_payroll_runs\nWHERE business_id = ?\nORDER BY created_at DESC\nLIMIT ? OFFSET ?")
        .bind(params.business_id)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `ListPayrollRunsForEmployee` (Go `ListPayrollRunsForEmployeeParams`).
pub struct ListPayrollRunsForEmployeeParams {
    pub business_id: i64,
    pub employee_id: i64,
    pub status: String,
    pub status_2: String,
    pub limit: i32,
    pub offset: i32,
}

/// `ListPayrollRunsForEmployee`: `many`.
pub async fn list_payroll_runs_for_employee<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListPayrollRunsForEmployeeParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT r.id, r.owner_id, r.business_id, r.period_start, r.period_end, r.status, r.created_by, r.approved_by, r.exported_at, r.total_gross, r.total_net, r.created_at, r.updated_at\nFROM hr_payroll_runs r\nJOIN hr_payroll_entries e ON e.payroll_run_id = r.id AND e.business_id = r.business_id\nWHERE r.business_id = ? AND e.employee_id = ? AND r.status IN (?, ?)\nORDER BY r.created_at DESC\nLIMIT ? OFFSET ?")
        .bind(params.business_id)
        .bind(params.employee_id)
        .bind(&params.status)
        .bind(&params.status_2)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `ReleasePayrollRunExportClaim` (Go `ReleasePayrollRunExportClaimParams`).
pub struct ReleasePayrollRunExportClaimParams {
    pub business_id: i64,
    pub id: i64,
}

/// `ReleasePayrollRunExportClaim`: `exec`.
pub async fn release_payroll_run_export_claim<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ReleasePayrollRunExportClaimParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE hr_payroll_runs\nSET status = 'approved'\nWHERE business_id = ? AND id = ? AND status = 'exporting' AND exported_at IS NULL")
        .bind(params.business_id)
        .bind(params.id)
        .execute(exec)
        .await
}

/// Params for `SetPayrollRunApproved` (Go `SetPayrollRunApprovedParams`).
pub struct SetPayrollRunApprovedParams {
    pub approved_by: Option<i64>,
    pub business_id: i64,
    pub id: i64,
}

/// `SetPayrollRunApproved`: `execresult`.
pub async fn set_payroll_run_approved<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &SetPayrollRunApprovedParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE hr_payroll_runs\nSET status = 'approved', approved_by = ?\nWHERE business_id = ? AND id = ? AND status = 'in_review'")
        .bind(params.approved_by)
        .bind(params.business_id)
        .bind(params.id)
        .execute(exec)
        .await
}

/// Params for `SetPayrollRunExported` (Go `SetPayrollRunExportedParams`).
pub struct SetPayrollRunExportedParams {
    pub business_id: i64,
    pub id: i64,
}

/// `SetPayrollRunExported`: `execresult`.
pub async fn set_payroll_run_exported<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &SetPayrollRunExportedParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE hr_payroll_runs\nSET status = 'exported', exported_at = CURRENT_TIMESTAMP\nWHERE business_id = ? AND id = ? AND status = 'exporting' AND exported_at IS NULL")
        .bind(params.business_id)
        .bind(params.id)
        .execute(exec)
        .await
}

/// Params for `UpsertPayrollConfig` (Go `UpsertPayrollConfigParams`).
pub struct UpsertPayrollConfigParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub period_type: String,
    pub pay_day_of_month: i32,
    pub overtime_enabled: bool,
}

/// `UpsertPayrollConfig`: `execresult`.
pub async fn upsert_payroll_config<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpsertPayrollConfigParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO hr_payroll_config (owner_id, business_id, period_type, pay_day_of_month, overtime_enabled)\nVALUES (?, ?, ?, ?, ?)\nON DUPLICATE KEY UPDATE\n    period_type = VALUES(period_type),\n    pay_day_of_month = VALUES(pay_day_of_month),\n    overtime_enabled = VALUES(overtime_enabled)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(&params.period_type)
        .bind(params.pay_day_of_month)
        .bind(params.overtime_enabled)
        .execute(exec)
        .await
}
