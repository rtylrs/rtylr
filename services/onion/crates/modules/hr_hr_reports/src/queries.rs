//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `AttendanceSummaryReport` (Go `AttendanceSummaryReportParams`).
pub struct AttendanceSummaryReportParams {
    pub from_date: chrono::NaiveDate,
    pub to_date: chrono::NaiveDate,
    pub business_id: i64,
    pub location_id: Option<i64>,
    pub department_id: Option<i64>,
    pub employee_id: Option<i64>,
}

/// `AttendanceSummaryReport`: `many`.
pub async fn attendance_summary_report<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &AttendanceSummaryReportParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT e.id AS employee_id,\n       CONCAT(e.first_name, ' ', e.last_name) AS employee_name,\n       COALESCE(d.name, 'Unassigned') AS department,\n       COUNT(a.id) AS total_records,\n       SUM(CASE WHEN a.status = 'present' THEN 1 ELSE 0 END) AS days_present,\n       SUM(CASE WHEN a.is_late = TRUE THEN 1 ELSE 0 END) AS days_late,\n       COALESCE(SUM(a.late_minutes), 0) AS total_late_minutes,\n       COALESCE(SUM(a.duration_minutes), 0) AS total_duration_minutes\nFROM hr_employees e\nLEFT JOIN hr_departments d ON d.id = e.department_id AND d.business_id = e.business_id\nLEFT JOIN hr_attendance_records a ON a.employee_id = e.id AND a.business_id = e.business_id AND a.date BETWEEN ? AND ?\nWHERE e.business_id = ?\n  AND e.status = 'active'\n  AND (? IS NULL OR e.location_id = ?)\n  AND (? IS NULL OR e.department_id = ?)\n  AND (? IS NULL OR e.id = ?)\nGROUP BY e.id, e.first_name, e.last_name, d.name\nORDER BY employee_name\nLIMIT 5001")
        .bind(params.from_date)
        .bind(params.to_date)
        .bind(params.business_id)
        .bind(params.location_id)
        .bind(params.location_id)
        .bind(params.department_id)
        .bind(params.department_id)
        .bind(params.employee_id)
        .bind(params.employee_id)
        .execute(exec)
        .await
}

/// Params for `HeadcountReport` (Go `HeadcountReportParams`).
pub struct HeadcountReportParams {
    pub business_id: i64,
    pub created_at: chrono::NaiveDate,
    pub location_id: Option<i64>,
    pub department_id: Option<i64>,
    pub employee_id: Option<i64>,
}

/// `HeadcountReport`: `many`.
pub async fn headcount_report<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &HeadcountReportParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT COALESCE(d.name, 'Unassigned') AS department,\n       COALESCE(l.name, 'Unassigned') AS location,\n       e.employment_type,\n       e.status,\n       COUNT(*) AS employee_count\nFROM hr_employees e\nLEFT JOIN hr_departments d ON d.id = e.department_id AND d.business_id = e.business_id\nLEFT JOIN hr_locations l ON l.id = e.location_id AND l.business_id = e.business_id\nWHERE e.business_id = ?\n  AND e.status = 'active'\n  AND e.created_at <= ?\n  AND (? IS NULL OR e.location_id = ?)\n  AND (? IS NULL OR e.department_id = ?)\n  AND (? IS NULL OR e.id = ?)\nGROUP BY d.name, l.name, e.employment_type, e.status\nORDER BY department, location\nLIMIT 5001")
        .bind(params.business_id)
        .bind(params.created_at)
        .bind(params.location_id)
        .bind(params.location_id)
        .bind(params.department_id)
        .bind(params.department_id)
        .bind(params.employee_id)
        .bind(params.employee_id)
        .execute(exec)
        .await
}

/// Params for `InsightsCountActiveEmployees` (Go `InsightsCountActiveEmployeesParams`).
pub struct InsightsCountActiveEmployeesParams {
    pub business_id: i64,
}

/// `InsightsCountActiveEmployees`: `one`.
pub async fn insights_count_active_employees<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &InsightsCountActiveEmployeesParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query(
        "SELECT COUNT(*) AS count FROM hr_employees WHERE business_id = ? AND status = 'active'",
    )
    .bind(params.business_id)
    .fetch_optional(exec)
    .await
}

/// Params for `InsightsCountActiveStaff` (Go `InsightsCountActiveStaffParams`).
pub struct InsightsCountActiveStaffParams {
    pub business_id: i64,
}

/// `InsightsCountActiveStaff`: `one`.
pub async fn insights_count_active_staff<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &InsightsCountActiveStaffParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query(
        "SELECT COUNT(*) AS count FROM hr_staff WHERE business_id = ? AND status = 'active'",
    )
    .bind(params.business_id)
    .fetch_optional(exec)
    .await
}

/// Params for `InsightsCountDepartures` (Go `InsightsCountDeparturesParams`).
pub struct InsightsCountDeparturesParams {
    pub business_id: i64,
    pub date_from: Option<chrono::DateTime<chrono::Utc>>,
    pub date_to: Option<chrono::DateTime<chrono::Utc>>,
}

/// `InsightsCountDepartures`: `one`.
pub async fn insights_count_departures<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &InsightsCountDeparturesParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) AS count\nFROM hr_employees\nWHERE business_id = ?\n  AND end_date >= ?\n  AND end_date <= ?")
        .bind(params.business_id)
        .bind(params.date_from)
        .bind(params.date_to)
        .fetch_optional(exec)
        .await
}

/// Params for `InsightsCountHires` (Go `InsightsCountHiresParams`).
pub struct InsightsCountHiresParams {
    pub business_id: i64,
    pub date_from: Option<chrono::DateTime<chrono::Utc>>,
    pub date_to: Option<chrono::DateTime<chrono::Utc>>,
}

/// `InsightsCountHires`: `one`.
pub async fn insights_count_hires<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &InsightsCountHiresParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) AS count\nFROM hr_employees\nWHERE business_id = ?\n  AND start_date >= ?\n  AND start_date <= ?")
        .bind(params.business_id)
        .bind(params.date_from)
        .bind(params.date_to)
        .fetch_optional(exec)
        .await
}

/// Params for `InsightsCountLeaveStarting` (Go `InsightsCountLeaveStartingParams`).
pub struct InsightsCountLeaveStartingParams {
    pub business_id: i64,
    pub status: String,
    pub date_from: chrono::NaiveDate,
    pub date_to: chrono::NaiveDate,
}

/// `InsightsCountLeaveStarting`: `one`.
pub async fn insights_count_leave_starting<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &InsightsCountLeaveStartingParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) AS count\nFROM hr_leave_requests\nWHERE business_id = ?\n  AND status = ?\n  AND start_date BETWEEN ? AND ?")
        .bind(params.business_id)
        .bind(&params.status)
        .bind(params.date_from)
        .bind(params.date_to)
        .fetch_optional(exec)
        .await
}

/// Params for `InsightsCountOpenLeave` (Go `InsightsCountOpenLeaveParams`).
pub struct InsightsCountOpenLeaveParams {
    pub business_id: i64,
}

/// `InsightsCountOpenLeave`: `one`.
pub async fn insights_count_open_leave<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &InsightsCountOpenLeaveParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) AS count\nFROM hr_leave_requests\nWHERE business_id = ?\n  AND status NOT IN ('rejected','cancelled','completed')")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `InsightsCountPendingLeave` (Go `InsightsCountPendingLeaveParams`).
pub struct InsightsCountPendingLeaveParams {
    pub business_id: i64,
}

/// `InsightsCountPendingLeave`: `one`.
pub async fn insights_count_pending_leave<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &InsightsCountPendingLeaveParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) AS count FROM hr_leave_requests WHERE business_id = ? AND status = 'pending'")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `InsightsCountShiftsThisWeek` (Go `InsightsCountShiftsThisWeekParams`).
pub struct InsightsCountShiftsThisWeekParams {
    pub business_id: i64,
}

/// `InsightsCountShiftsThisWeek`: `one`.
pub async fn insights_count_shifts_this_week<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &InsightsCountShiftsThisWeekParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) AS count\nFROM hr_shift_assignments\nWHERE business_id = ?\n  AND date >= DATE_SUB(CURDATE(), INTERVAL WEEKDAY(CURDATE()) DAY)\n  AND date <  DATE_ADD(DATE_SUB(CURDATE(), INTERVAL WEEKDAY(CURDATE()) DAY), INTERVAL 7 DAY)")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `InsightsHRLastChange` (Go `InsightsHRLastChangeParams`).
pub struct InsightsHRLastChangeParams {
    pub business_id: i64,
}

/// `InsightsHRLastChange`: `one`.
pub async fn insights_hr_last_change<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &InsightsHRLastChangeParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT CAST(GREATEST(\n  COALESCE((SELECT MAX(e.updated_at) FROM hr_employees e WHERE e.business_id = ?), '1000-01-01'),\n  COALESCE((SELECT MAX(l.updated_at) FROM hr_leave_requests l WHERE l.business_id = ?), '1000-01-01'),\n  COALESCE((SELECT MAX(a.updated_at) FROM hr_shift_assignments a WHERE a.business_id = ?), '1000-01-01')\n) AS DATETIME) AS last_change")
        .bind(params.business_id)
        .bind(params.business_id)
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `InsightsHeadcountAsOf` (Go `InsightsHeadcountAsOfParams`).
pub struct InsightsHeadcountAsOfParams {
    pub business_id: i64,
    pub as_of: Option<chrono::DateTime<chrono::Utc>>,
}

/// `InsightsHeadcountAsOf`: `one`.
pub async fn insights_headcount_as_of<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &InsightsHeadcountAsOfParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) AS count\nFROM hr_employees\nWHERE business_id = ?\n  AND (start_date IS NULL OR start_date <= ?)\n  AND (end_date IS NULL OR end_date > ?)")
        .bind(params.business_id)
        .bind(params.as_of)
        .bind(params.as_of)
        .fetch_optional(exec)
        .await
}

/// Params for `InsightsScheduledHours` (Go `InsightsScheduledHoursParams`).
pub struct InsightsScheduledHoursParams {
    pub business_id: i64,
    pub date_from: chrono::NaiveDate,
    pub date_to: chrono::NaiveDate,
}

/// `InsightsScheduledHours`: `one`.
pub async fn insights_scheduled_hours<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &InsightsScheduledHoursParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COALESCE(SUM(\n  TIME_TO_SEC(STR_TO_DATE(s.end_time, '%H:%i'))\n  - TIME_TO_SEC(STR_TO_DATE(s.start_time, '%H:%i'))\n) / 3600, 0) AS scheduled_hours\nFROM hr_shift_assignments a\nJOIN hr_shifts s ON s.id = a.shift_id AND s.business_id = a.business_id\nWHERE a.business_id = ?\n  AND a.status <> 'cancelled'\n  AND a.date >= ?\n  AND a.date <= ?")
        .bind(params.business_id)
        .bind(params.date_from)
        .bind(params.date_to)
        .fetch_optional(exec)
        .await
}

/// Params for `InsightsShiftMetrics` (Go `InsightsShiftMetricsParams`).
pub struct InsightsShiftMetricsParams {
    pub business_id: i64,
    pub date_from: chrono::NaiveDate,
    pub date_to: chrono::NaiveDate,
}

/// `InsightsShiftMetrics`: `one`.
pub async fn insights_shift_metrics<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &InsightsShiftMetricsParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT\n  SUM(CASE WHEN status <> 'cancelled' THEN 1 ELSE 0 END) AS scheduled,\n  SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END)  AS completed,\n  SUM(CASE WHEN status = 'no_show'   THEN 1 ELSE 0 END)  AS no_shows\nFROM hr_shift_assignments\nWHERE business_id = ?\n  AND date BETWEEN ? AND ?")
        .bind(params.business_id)
        .bind(params.date_from)
        .bind(params.date_to)
        .fetch_optional(exec)
        .await
}

/// Params for `LeaveSummaryReport` (Go `LeaveSummaryReportParams`).
pub struct LeaveSummaryReportParams {
    pub year: chrono::NaiveDate,
    pub business_id: i64,
    pub start_date: chrono::NaiveDate,
    pub end_date: chrono::NaiveDate,
    pub business_id_2: i64,
    pub location_id: Option<i64>,
    pub department_id: Option<i64>,
    pub employee_id: Option<i64>,
}

/// `LeaveSummaryReport`: `many`.
pub async fn leave_summary_report<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &LeaveSummaryReportParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT e.id AS employee_id,\n       CONCAT(e.first_name, ' ', e.last_name) AS employee_name,\n       COALESCE(d.name, 'Unassigned') AS department,\n       lt.name AS leave_type,\n       COALESCE(lb.entitled_days, 0) AS entitled_days,\n       COALESCE(lb.used_days, 0) AS used_days,\n       COALESCE(lb.pending_days, 0) AS pending_days,\n       COALESCE(lb.entitled_days - lb.used_days - lb.pending_days, 0) AS remaining_days,\n       COALESCE(lr.requests_in_period, 0) AS requests_in_period,\n       COALESCE(lr.approved_days_in_period, 0) AS approved_days_in_period\nFROM hr_employees e\nLEFT JOIN hr_departments d ON d.id = e.department_id AND d.business_id = e.business_id\nCROSS JOIN hr_leave_types lt ON lt.business_id = e.business_id AND lt.is_active = TRUE\nLEFT JOIN hr_leave_balances lb ON lb.employee_id = e.id AND lb.business_id = e.business_id AND lb.leave_type_id = lt.id AND lb.year = YEAR(?)\nLEFT JOIN (\n    SELECT req.employee_id, req.business_id, req.leave_type_id,\n           COUNT(*) AS requests_in_period,\n           COALESCE(SUM(CASE WHEN req.status = 'approved' THEN req.total_days ELSE 0 END), 0) AS approved_days_in_period\n    FROM hr_leave_requests req\n    WHERE req.business_id = ? AND req.start_date <= ? AND req.end_date >= ?\n    GROUP BY req.employee_id, req.business_id, req.leave_type_id\n) lr ON lr.employee_id = e.id AND lr.business_id = e.business_id AND lr.leave_type_id = lt.id\nWHERE e.business_id = ?\n  AND e.status = 'active'\n  AND (? IS NULL OR e.location_id = ?)\n  AND (? IS NULL OR e.department_id = ?)\n  AND (? IS NULL OR e.id = ?)\nGROUP BY e.id, e.first_name, e.last_name, d.name, lt.id, lt.name, lb.entitled_days, lb.used_days, lb.pending_days, lr.requests_in_period, lr.approved_days_in_period\nORDER BY employee_name, leave_type\nLIMIT 5001")
        .bind(params.year)
        .bind(params.business_id)
        .bind(params.start_date)
        .bind(params.end_date)
        .bind(params.business_id_2)
        .bind(params.location_id)
        .bind(params.location_id)
        .bind(params.department_id)
        .bind(params.department_id)
        .bind(params.employee_id)
        .bind(params.employee_id)
        .execute(exec)
        .await
}

/// Params for `OvertimeSummaryReport` (Go `OvertimeSummaryReportParams`).
pub struct OvertimeSummaryReportParams {
    pub from_date: chrono::NaiveDate,
    pub to_date: chrono::NaiveDate,
    pub business_id: i64,
    pub location_id: Option<i64>,
    pub department_id: Option<i64>,
    pub employee_id: Option<i64>,
}

/// `OvertimeSummaryReport`: `many`.
pub async fn overtime_summary_report<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &OvertimeSummaryReportParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT e.id AS employee_id,\n       CONCAT(e.first_name, ' ', e.last_name) AS employee_name,\n       COALESCE(d.name, 'Unassigned') AS department,\n       COUNT(o.id) AS total_requests,\n       SUM(CASE WHEN o.status = 'approved' THEN 1 ELSE 0 END) AS approved_requests,\n       SUM(CASE WHEN o.status = 'pending' THEN 1 ELSE 0 END) AS pending_requests,\n       SUM(CASE WHEN o.status = 'rejected' THEN 1 ELSE 0 END) AS rejected_requests,\n       COALESCE(SUM(CASE WHEN o.status = 'approved' THEN o.planned_extra_minutes ELSE 0 END), 0) AS approved_overtime_minutes,\n       COALESCE(AVG(CASE WHEN o.status = 'approved' THEN o.multiplier_applied ELSE NULL END), 0) AS avg_multiplier\nFROM hr_employees e\nLEFT JOIN hr_departments d ON d.id = e.department_id AND d.business_id = e.business_id\nLEFT JOIN hr_overtime_requests o ON o.employee_id = e.id AND o.business_id = e.business_id AND o.date BETWEEN ? AND ?\nWHERE e.business_id = ?\n  AND e.status = 'active'\n  AND (? IS NULL OR e.location_id = ?)\n  AND (? IS NULL OR e.department_id = ?)\n  AND (? IS NULL OR e.id = ?)\nGROUP BY e.id, e.first_name, e.last_name, d.name\nORDER BY employee_name\nLIMIT 5001")
        .bind(params.from_date)
        .bind(params.to_date)
        .bind(params.business_id)
        .bind(params.location_id)
        .bind(params.location_id)
        .bind(params.department_id)
        .bind(params.department_id)
        .bind(params.employee_id)
        .bind(params.employee_id)
        .execute(exec)
        .await
}

/// Params for `PayrollSummaryReport` (Go `PayrollSummaryReportParams`).
pub struct PayrollSummaryReportParams {
    pub period_start: chrono::NaiveDate,
    pub period_end: chrono::NaiveDate,
    pub business_id: i64,
    pub location_id: Option<i64>,
    pub department_id: Option<i64>,
    pub employee_id: Option<i64>,
}

/// `PayrollSummaryReport`: `many`.
pub async fn payroll_summary_report<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &PayrollSummaryReportParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT e.id AS employee_id,\n       CONCAT(e.first_name, ' ', e.last_name) AS employee_name,\n       COALESCE(d.name, 'Unassigned') AS department,\n       COALESCE(SUM(pe.base_pay), 0) AS total_base_pay,\n       COALESCE(SUM(pe.overtime_pay), 0) AS total_overtime_pay,\n       COALESCE(SUM(pe.bonuses), 0) AS total_bonuses,\n       COALESCE(SUM(pe.deductions), 0) AS total_deductions,\n       COALESCE(SUM(pe.base_pay + pe.overtime_pay + pe.bonuses), 0) AS total_gross_pay,\n       COALESCE(SUM(pe.net_pay), 0) AS total_net_pay,\n       COALESCE(SUM(pe.hours_worked), 0) AS total_hours_worked,\n       COALESCE(SUM(pe.overtime_hours), 0) AS total_overtime_hours\nFROM hr_employees e\nLEFT JOIN hr_departments d ON d.id = e.department_id AND d.business_id = e.business_id\nLEFT JOIN hr_payroll_entries pe ON pe.employee_id = e.id AND pe.business_id = e.business_id\nLEFT JOIN hr_payroll_runs pr ON pr.id = pe.payroll_run_id AND pr.business_id = e.business_id AND pr.period_start >= ? AND pr.period_end <= ?\nWHERE e.business_id = ?\n  AND e.status = 'active'\n  AND (? IS NULL OR e.location_id = ?)\n  AND (? IS NULL OR e.department_id = ?)\n  AND (? IS NULL OR e.id = ?)\nGROUP BY e.id, e.first_name, e.last_name, d.name\nORDER BY employee_name\nLIMIT 5001")
        .bind(params.period_start)
        .bind(params.period_end)
        .bind(params.business_id)
        .bind(params.location_id)
        .bind(params.location_id)
        .bind(params.department_id)
        .bind(params.department_id)
        .bind(params.employee_id)
        .bind(params.employee_id)
        .execute(exec)
        .await
}

/// Params for `TimesheetSummaryReport` (Go `TimesheetSummaryReportParams`).
pub struct TimesheetSummaryReportParams {
    pub period_start: chrono::NaiveDate,
    pub period_end: chrono::NaiveDate,
    pub business_id: i64,
    pub location_id: Option<i64>,
    pub department_id: Option<i64>,
    pub employee_id: Option<i64>,
}

/// `TimesheetSummaryReport`: `many`.
pub async fn timesheet_summary_report<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &TimesheetSummaryReportParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT e.id AS employee_id,\n       CONCAT(e.first_name, ' ', e.last_name) AS employee_name,\n       COALESCE(d.name, 'Unassigned') AS department,\n       COUNT(DISTINCT ts.id) AS total_timesheets,\n       COALESCE(SUM(ts.total_minutes), 0) AS total_logged_minutes,\n       COALESCE(SUM(CASE WHEN ts.status = 'approved' THEN ts.total_minutes ELSE 0 END), 0) AS approved_minutes,\n       COALESCE(SUM(CASE WHEN ts.status = 'draft' THEN ts.total_minutes ELSE 0 END), 0) AS draft_minutes,\n       COALESCE(SUM(CASE WHEN ts.status = 'submitted' THEN ts.total_minutes ELSE 0 END), 0) AS submitted_minutes\nFROM hr_employees e\nLEFT JOIN hr_departments d ON d.id = e.department_id AND d.business_id = e.business_id\nLEFT JOIN hr_timesheets ts ON ts.employee_id = e.id AND ts.business_id = e.business_id AND ts.period_start >= ? AND ts.period_end <= ?\nWHERE e.business_id = ?\n  AND e.status = 'active'\n  AND (? IS NULL OR e.location_id = ?)\n  AND (? IS NULL OR e.department_id = ?)\n  AND (? IS NULL OR e.id = ?)\nGROUP BY e.id, e.first_name, e.last_name, d.name\nORDER BY employee_name\nLIMIT 5001")
        .bind(params.period_start)
        .bind(params.period_end)
        .bind(params.business_id)
        .bind(params.location_id)
        .bind(params.location_id)
        .bind(params.department_id)
        .bind(params.department_id)
        .bind(params.employee_id)
        .bind(params.employee_id)
        .execute(exec)
        .await
}
