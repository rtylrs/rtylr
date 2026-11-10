//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `AdvanceOvertimeApprovalLevel` (Go `AdvanceOvertimeApprovalLevelParams`).
pub struct AdvanceOvertimeApprovalLevelParams {
    pub id: i64,
    pub business_id: i64,
    pub current_approval_level: i32,
}

/// `AdvanceOvertimeApprovalLevel`: `execresult`.
pub async fn advance_overtime_approval_level<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &AdvanceOvertimeApprovalLevelParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE hr_overtime_requests\nSET current_approval_level = current_approval_level + 1, updated_at = NOW()\nWHERE id = ? AND business_id = ? AND status = 'pending' AND current_approval_level = ?")
        .bind(params.id)
        .bind(params.business_id)
        .bind(params.current_approval_level)
        .execute(exec)
        .await
}

/// Params for `ApproveOvertimeRequest` (Go `ApproveOvertimeRequestParams`).
pub struct ApproveOvertimeRequestParams {
    pub approved_by: Option<i64>,
    pub multiplier_applied: Option<String>,
    pub id: i64,
    pub business_id: i64,
}

/// `ApproveOvertimeRequest`: `execresult`.
pub async fn approve_overtime_request<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ApproveOvertimeRequestParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE hr_overtime_requests\nSET status = 'approved', approved_by = ?, multiplier_applied = ?, rejection_reason = NULL, updated_at = NOW()\nWHERE id = ? AND business_id = ? AND status = 'pending'")
        .bind(params.approved_by)
        .bind(&params.multiplier_applied)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `CountOvertimeRequests` (Go `CountOvertimeRequestsParams`).
pub struct CountOvertimeRequestsParams {
    pub business_id: i64,
    pub employee_id: Option<i64>,
    pub status: Option<String>,
}

/// `CountOvertimeRequests`: `one`.
pub async fn count_overtime_requests<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountOvertimeRequestsParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) AS total\nFROM hr_overtime_requests\nWHERE business_id = ?\n  AND (? IS NULL OR employee_id = ?)\n  AND (? IS NULL OR status = ?)")
        .bind(params.business_id)
        .bind(params.employee_id)
        .bind(params.employee_id)
        .bind(&params.status)
        .bind(&params.status)
        .fetch_optional(exec)
        .await
}

/// Params for `CreateFlaggedOvertimeRequest` (Go `CreateFlaggedOvertimeRequestParams`).
pub struct CreateFlaggedOvertimeRequestParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub employee_id: i64,
    pub date: chrono::NaiveDate,
    pub planned_extra_minutes: i32,
    pub business_id_2: i64,
    pub employee_id_2: i64,
    pub date_2: chrono::NaiveDate,
}

/// `CreateFlaggedOvertimeRequest`: `execresult`.
pub async fn create_flagged_overtime_request<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateFlaggedOvertimeRequestParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO hr_overtime_requests (\n    owner_id, business_id, employee_id, date, planned_extra_minutes, reason, status, current_approval_level\n) SELECT ?, ?, ?, ?, ?, 'auto-detected overtime', 'flagged', 0\nWHERE NOT EXISTS (\n    SELECT 1 FROM hr_overtime_requests AS existing\n    WHERE existing.business_id = ? AND existing.employee_id = ? AND existing.date = ?\n      AND existing.status IN ('flagged', 'pending', 'approved')\n)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(params.employee_id)
        .bind(params.date)
        .bind(params.planned_extra_minutes)
        .bind(params.business_id_2)
        .bind(params.employee_id_2)
        .bind(params.date_2)
        .execute(exec)
        .await
}

/// Params for `CreateOvertimeRequest` (Go `CreateOvertimeRequestParams`).
pub struct CreateOvertimeRequestParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub employee_id: i64,
    pub date: chrono::NaiveDate,
    pub planned_extra_minutes: i32,
    pub reason: Option<String>,
    pub status: String,
    pub current_approval_level: i32,
}

/// `CreateOvertimeRequest`: `execresult`.
pub async fn create_overtime_request<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateOvertimeRequestParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO hr_overtime_requests (\n    owner_id, business_id, employee_id, date, planned_extra_minutes, reason, status, current_approval_level\n) VALUES (?, ?, ?, ?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(params.employee_id)
        .bind(params.date)
        .bind(params.planned_extra_minutes)
        .bind(&params.reason)
        .bind(&params.status)
        .bind(params.current_approval_level)
        .execute(exec)
        .await
}

/// Params for `GetOvertimeConfig` (Go `GetOvertimeConfigParams`).
pub struct GetOvertimeConfigParams {
    pub business_id: i64,
}

/// `GetOvertimeConfig`: `one`.
pub async fn get_overtime_config<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetOvertimeConfigParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, daily_threshold_minutes, default_multiplier, auto_detect_enabled,\n       created_at, updated_at\nFROM hr_overtime_config\nWHERE business_id = ?")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetOvertimeRequest` (Go `GetOvertimeRequestParams`).
pub struct GetOvertimeRequestParams {
    pub business_id: i64,
    pub id: i64,
}

/// `GetOvertimeRequest`: `one`.
pub async fn get_overtime_request<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetOvertimeRequestParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, employee_id, date, planned_extra_minutes, reason, status,\n       current_approval_level, rejection_reason, approved_by, multiplier_applied, created_at, updated_at\nFROM hr_overtime_requests\nWHERE business_id = ? AND id = ?")
        .bind(params.business_id)
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `ListAllConfiguredOwnerIDs` (Go `ListAllConfiguredOwnerIDsParams`).
pub struct ListAllConfiguredOwnerIDsParams {
    pub limit: i32,
    pub offset: i32,
}

/// `ListAllConfiguredOwnerIDs`: `many`.
pub async fn list_all_configured_owner_i_ds<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListAllConfiguredOwnerIDsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT DISTINCT business_id FROM hr_overtime_config WHERE auto_detect_enabled = TRUE\nORDER BY business_id ASC\nLIMIT ? OFFSET ?")
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `ListFlaggedForDate` (Go `ListFlaggedForDateParams`).
pub struct ListFlaggedForDateParams {
    pub business_id: i64,
    pub date: chrono::NaiveDate,
}

/// `ListFlaggedForDate`: `many`.
pub async fn list_flagged_for_date<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListFlaggedForDateParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, employee_id, date, planned_extra_minutes, reason, status,\n       current_approval_level, rejection_reason, approved_by, multiplier_applied, created_at, updated_at\nFROM hr_overtime_requests\nWHERE business_id = ? AND date = ? AND status IN ('flagged', 'pending', 'approved')")
        .bind(params.business_id)
        .bind(params.date)
        .execute(exec)
        .await
}

/// Params for `ListOvertimeCandidates` (Go `ListOvertimeCandidatesParams`).
pub struct ListOvertimeCandidatesParams {
    pub duration_minutes: i32,
    pub business_id: i64,
    pub date: chrono::NaiveDate,
    pub duration_minutes_2: i32,
}

/// `ListOvertimeCandidates`: `many`.
pub async fn list_overtime_candidates<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListOvertimeCandidatesParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT employee_id, date, SUM(duration_minutes) - ? AS extra_minutes\nFROM hr_time_entries\nWHERE business_id = ? AND date = ? AND status IN ('submitted', 'approved')\nGROUP BY employee_id, date\nHAVING SUM(duration_minutes) > ?\nORDER BY employee_id ASC")
        .bind(params.duration_minutes)
        .bind(params.business_id)
        .bind(params.date)
        .bind(params.duration_minutes_2)
        .execute(exec)
        .await
}

/// Params for `ListOvertimeRequests` (Go `ListOvertimeRequestsParams`).
pub struct ListOvertimeRequestsParams {
    pub business_id: i64,
    pub employee_id: Option<i64>,
    pub status: Option<String>,
    pub limit: i32,
    pub offset: i32,
}

/// `ListOvertimeRequests`: `many`.
pub async fn list_overtime_requests<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListOvertimeRequestsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, employee_id, date, planned_extra_minutes, reason, status,\n       current_approval_level, rejection_reason, approved_by, multiplier_applied, created_at, updated_at\nFROM hr_overtime_requests\nWHERE business_id = ?\n  AND (? IS NULL OR employee_id = ?)\n  AND (? IS NULL OR status = ?)\nORDER BY created_at DESC\nLIMIT ? OFFSET ?")
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

/// Params for `RejectOvertimeRequest` (Go `RejectOvertimeRequestParams`).
pub struct RejectOvertimeRequestParams {
    pub rejection_reason: Option<String>,
    pub id: i64,
    pub business_id: i64,
}

/// `RejectOvertimeRequest`: `execresult`.
pub async fn reject_overtime_request<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &RejectOvertimeRequestParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE hr_overtime_requests\nSET status = 'rejected', rejection_reason = ?, updated_at = NOW()\nWHERE id = ? AND business_id = ? AND status = 'pending'")
        .bind(&params.rejection_reason)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `UpsertOvertimeConfig` (Go `UpsertOvertimeConfigParams`).
pub struct UpsertOvertimeConfigParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub daily_threshold_minutes: i32,
    pub default_multiplier: String,
    pub auto_detect_enabled: bool,
}

/// `UpsertOvertimeConfig`: `execresult`.
pub async fn upsert_overtime_config<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpsertOvertimeConfigParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO hr_overtime_config (owner_id, business_id, daily_threshold_minutes, default_multiplier, auto_detect_enabled)\nVALUES (?, ?, ?, ?, ?)\nON DUPLICATE KEY UPDATE daily_threshold_minutes = VALUES(daily_threshold_minutes),\n    default_multiplier = VALUES(default_multiplier), auto_detect_enabled = VALUES(auto_detect_enabled),\n    updated_at = NOW()")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(params.daily_threshold_minutes)
        .bind(&params.default_multiplier)
        .bind(params.auto_detect_enabled)
        .execute(exec)
        .await
}
