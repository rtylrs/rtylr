//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `CreatePayrollRecord` (Go `CreatePayrollRecordParams`).
pub struct CreatePayrollRecordParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub staff_id: i64,
    pub period_start: chrono::NaiveDate,
    pub period_end: chrono::NaiveDate,
    pub period_label: String,
    pub regular_hours: Option<String>,
    pub overtime_hours: Option<String>,
    pub base_pay: String,
    pub overtime_pay: Option<String>,
    pub net_pay: String,
}

/// `CreatePayrollRecord`: `execresult`.
pub async fn create_payroll_record<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreatePayrollRecordParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO hr_payroll_records\n(owner_id, business_id, staff_id, period_start, period_end, period_label, regular_hours, overtime_hours, base_pay, overtime_pay, net_pay)\nVALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(params.staff_id)
        .bind(params.period_start)
        .bind(params.period_end)
        .bind(&params.period_label)
        .bind(&params.regular_hours)
        .bind(&params.overtime_hours)
        .bind(&params.base_pay)
        .bind(&params.overtime_pay)
        .bind(&params.net_pay)
        .execute(exec)
        .await
}

/// Params for `DeleteUnpaidPayrollRecords` (Go `DeleteUnpaidPayrollRecordsParams`).
pub struct DeleteUnpaidPayrollRecordsParams {
    pub business_id: i64,
    pub period_label: String,
}

/// `DeleteUnpaidPayrollRecords`: `exec`.
pub async fn delete_unpaid_payroll_records<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DeleteUnpaidPayrollRecordsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("DELETE FROM hr_payroll_records\nWHERE business_id = ? AND period_label = ? AND status <> 'paid'")
        .bind(params.business_id)
        .bind(&params.period_label)
        .execute(exec)
        .await
}

/// Params for `GetCompletedShiftsForPeriod` (Go `GetCompletedShiftsForPeriodParams`).
pub struct GetCompletedShiftsForPeriodParams {
    pub business_id: i64,
    pub date: Option<chrono::DateTime<chrono::Utc>>,
    pub date_2: Option<chrono::DateTime<chrono::Utc>>,
}

/// `GetCompletedShiftsForPeriod`: `many`.
pub async fn get_completed_shifts_for_period<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetCompletedShiftsForPeriodParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT staff_id, SUM(hours_worked) AS total_hours, SUM(overtime_hours) AS total_overtime\nFROM hr_shifts\nWHERE business_id = ? AND status = 'completed'\n  AND date >= ? AND date <= ?\nGROUP BY staff_id\nLIMIT 200")
        .bind(params.business_id)
        .bind(params.date)
        .bind(params.date_2)
        .execute(exec)
        .await
}

/// Params for `GetOvertimeConfigForLegacyPayroll` (Go `GetOvertimeConfigForLegacyPayrollParams`).
pub struct GetOvertimeConfigForLegacyPayrollParams {
    pub business_id: i64,
}

/// `GetOvertimeConfigForLegacyPayroll`: `one`.
pub async fn get_overtime_config_for_legacy_payroll<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetOvertimeConfigForLegacyPayrollParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT default_multiplier\nFROM hr_overtime_config\nWHERE business_id = ?")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetPayrollRecordPeriod` (Go `GetPayrollRecordPeriodParams`).
pub struct GetPayrollRecordPeriodParams {
    pub id: i64,
    pub business_id: i64,
}

/// `GetPayrollRecordPeriod`: `one`.
pub async fn get_payroll_record_period<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetPayrollRecordPeriodParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query(
        "SELECT period_start, period_end FROM hr_payroll_records WHERE id = ? AND business_id = ?",
    )
    .bind(params.id)
    .bind(params.business_id)
    .fetch_optional(exec)
    .await
}

/// Params for `GetStaffPayRates` (Go `GetStaffPayRatesParams`).
pub struct GetStaffPayRatesParams {
    pub id: i64,
    pub business_id: i64,
}

/// `GetStaffPayRates`: `one`.
pub async fn get_staff_pay_rates<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetStaffPayRatesParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query(
        "SELECT salary, hourly_rate, overtime_rate FROM hr_staff WHERE id = ? AND business_id = ?",
    )
    .bind(params.id)
    .bind(params.business_id)
    .fetch_optional(exec)
    .await
}

/// Params for `GetStaffPayRatesByIDs` (Go `GetStaffPayRatesByIDsParams`).
pub struct GetStaffPayRatesByIDsParams {
    pub business_id: i64,
    pub staff_ids: Vec<i64>,
}

/// `GetStaffPayRatesByIDs`: `many`.
pub async fn get_staff_pay_rates_by_i_ds<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetStaffPayRatesByIDsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    let mut sql_template: String = "SELECT id, salary, hourly_rate, overtime_rate\nFROM hr_staff\nWHERE business_id = ? AND id IN ({SLICE:staff_ids})".to_string();

    if params.staff_ids.is_empty() {
        sql_template = sql_template.replace("{SLICE:staff_ids}", "NULL");
    } else {
        sql_template = sql_template.replace(
            "{SLICE:staff_ids}",
            &std::iter::repeat_n("?", params.staff_ids.len())
                .collect::<Vec<_>>()
                .join(","),
        );
    }
    let query = sqlx::query(&sql_template);
    let mut query = query.bind(params.business_id);
    for value in &params.staff_ids {
        query = query.bind(value);
    }
    query.execute(exec).await
}

/// Params for `ListPayrollPeriodRowsForUpdate` (Go `ListPayrollPeriodRowsForUpdateParams`).
pub struct ListPayrollPeriodRowsForUpdateParams {
    pub business_id: i64,
    pub period_start: chrono::NaiveDate,
    pub period_end: chrono::NaiveDate,
}

/// `ListPayrollPeriodRowsForUpdate`: `many`.
pub async fn list_payroll_period_rows_for_update<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListPayrollPeriodRowsForUpdateParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, staff_id, status FROM hr_payroll_records\nWHERE business_id = ? AND period_start = ? AND period_end = ?\nFOR UPDATE")
        .bind(params.business_id)
        .bind(params.period_start)
        .bind(params.period_end)
        .execute(exec)
        .await
}

/// Params for `ListPayrollRecords` (Go `ListPayrollRecordsParams`).
pub struct ListPayrollRecordsParams {
    pub business_id: i64,
}

/// `ListPayrollRecords`: `many`.
pub async fn list_payroll_records<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListPayrollRecordsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, staff_id, period_start, period_end, period_label,\n       regular_hours, overtime_hours, base_pay, overtime_pay, bonuses,\n       deductions, tax_withheld, net_pay, status, paid_at, notes, created_at\nFROM hr_payroll_records\nWHERE business_id = ?\nORDER BY period_start DESC, staff_id ASC\nLIMIT 200")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `ListPayrollRecordsByPeriod` (Go `ListPayrollRecordsByPeriodParams`).
pub struct ListPayrollRecordsByPeriodParams {
    pub business_id: i64,
    pub period_label: String,
}

/// `ListPayrollRecordsByPeriod`: `many`.
pub async fn list_payroll_records_by_period<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListPayrollRecordsByPeriodParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, staff_id, period_start, period_end, period_label,\n       regular_hours, overtime_hours, base_pay, overtime_pay, bonuses,\n       deductions, tax_withheld, net_pay, status, paid_at, notes, created_at\nFROM hr_payroll_records\nWHERE business_id = ? AND period_label = ?\nORDER BY period_start DESC, staff_id ASC\nLIMIT 200")
        .bind(params.business_id)
        .bind(&params.period_label)
        .execute(exec)
        .await
}

/// Params for `MarkPayrollPaid` (Go `MarkPayrollPaidParams`).
pub struct MarkPayrollPaidParams {
    pub id: i64,
    pub business_id: i64,
}

/// `MarkPayrollPaid`: `execresult`.
pub async fn mark_payroll_paid<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &MarkPayrollPaidParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE hr_payroll_records\nSET status = 'paid', paid_at = NOW()\nWHERE id = ? AND business_id = ? AND status != 'paid'")
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}
