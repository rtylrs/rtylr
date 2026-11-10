//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `ApproveTimeOffRequest` (Go `ApproveTimeOffRequestParams`).
pub struct ApproveTimeOffRequestParams {
    pub status: String,
    pub approved_by: Option<i64>,
    pub notes: Option<String>,
    pub id: i64,
    pub business_id: i64,
}

/// `ApproveTimeOffRequest`: `execresult`.
pub async fn approve_time_off_request<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ApproveTimeOffRequestParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE hr_time_off_requests\nSET status = ?, approved_by = ?, approved_at = NOW(), notes = COALESCE(?, notes), updated_at = NOW()\nWHERE id = ? AND business_id = ? AND status = 'pending'")
        .bind(&params.status)
        .bind(params.approved_by)
        .bind(&params.notes)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `CreateTimeOffRequest` (Go `CreateTimeOffRequestParams`).
pub struct CreateTimeOffRequestParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub staff_id: i64,
    pub r#type: String,
    pub start_date: chrono::NaiveDate,
    pub end_date: chrono::NaiveDate,
    pub days_count: String,
    pub reason: Option<String>,
    pub notes: Option<String>,
}

/// `CreateTimeOffRequest`: `execresult`.
pub async fn create_time_off_request<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateTimeOffRequestParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO hr_time_off_requests\n(owner_id, business_id, staff_id, type, start_date, end_date, days_count, reason, notes)\nVALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(params.staff_id)
        .bind(&params.r#type)
        .bind(params.start_date)
        .bind(params.end_date)
        .bind(&params.days_count)
        .bind(&params.reason)
        .bind(&params.notes)
        .execute(exec)
        .await
}

/// Params for `ListTimeOffRequests` (Go `ListTimeOffRequestsParams`).
pub struct ListTimeOffRequestsParams {
    pub business_id: i64,
    pub only_user_id: Option<i64>,
}

/// `ListTimeOffRequests`: `many`.
pub async fn list_time_off_requests<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListTimeOffRequestsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT t.id, t.owner_id, t.business_id, t.staff_id, t.type, t.start_date, t.end_date, t.days_count,\n       t.status, t.reason, t.approved_by, t.approved_at, t.notes, t.created_at, t.updated_at,\n       COALESCE(TRIM(CONCAT(s.first_name, ' ', s.last_name)), '') AS staff_name\nFROM hr_time_off_requests t\nLEFT JOIN hr_staff s ON s.id = t.staff_id AND s.business_id = t.business_id\nWHERE t.business_id = ?\n  AND (? IS NULL OR s.user_id = ?)\nORDER BY t.created_at DESC, t.id DESC\nLIMIT 500")
        .bind(params.business_id)
        .bind(params.only_user_id)
        .bind(params.only_user_id)
        .execute(exec)
        .await
}

/// Params for `StaffBelongsToUser` (Go `StaffBelongsToUserParams`).
pub struct StaffBelongsToUserParams {
    pub id: i64,
    pub business_id: i64,
    pub user_id: Option<i64>,
}

/// `StaffBelongsToUser`: `one`.
pub async fn staff_belongs_to_user<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &StaffBelongsToUserParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) FROM hr_staff WHERE id = ? AND business_id = ? AND user_id = ?")
        .bind(params.id)
        .bind(params.business_id)
        .bind(params.user_id)
        .fetch_optional(exec)
        .await
}
