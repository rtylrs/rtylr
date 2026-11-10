//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `CreateAuditLog` (Go `CreateAuditLogParams`).
pub struct CreateAuditLogParams {
    pub business_id: i64,
    pub branch_id: i64,
    pub terminal_id: Option<i64>,
    pub employee_id: i64,
    pub action: String,
    pub entity_type: String,
    pub entity_id: Option<i64>,
    pub details: String,
    pub ip_address: Option<String>,
}

/// `CreateAuditLog`: `execresult`.
pub async fn create_audit_log<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateAuditLogParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO pos_audit_log (business_id, branch_id, terminal_id, employee_id, action, entity_type, entity_id, details, ip_address)\nVALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)")
        .bind(params.business_id)
        .bind(params.branch_id)
        .bind(params.terminal_id)
        .bind(params.employee_id)
        .bind(&params.action)
        .bind(&params.entity_type)
        .bind(params.entity_id)
        .bind(&params.details)
        .bind(&params.ip_address)
        .execute(exec)
        .await
}

/// Params for `ListAuditLogs` (Go `ListAuditLogsParams`).
pub struct ListAuditLogsParams {
    pub business_id: i64,
    pub action: Option<String>,
    pub entity_type: Option<String>,
    pub employee_id: Option<i64>,
    pub start_date: Option<chrono::DateTime<chrono::Utc>>,
    pub end_date: Option<chrono::DateTime<chrono::Utc>>,
    pub cursor_created_at: Option<chrono::DateTime<chrono::Utc>>,
    pub cursor_id: Option<i64>,
    pub limit: i32,
}

/// `ListAuditLogs`: `many`.
pub async fn list_audit_logs<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListAuditLogsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, business_id, branch_id, terminal_id, employee_id, action, entity_type, entity_id, details, ip_address, created_at\nFROM pos_audit_log\nWHERE business_id = ?\n  AND (? IS NULL OR action = ?)\n  AND (? IS NULL OR entity_type = ?)\n  AND (? IS NULL OR employee_id = ?)\n  AND (? IS NULL OR created_at >= ?)\n  AND (? IS NULL OR created_at <= ?)\n  AND (\n    ? IS NULL\n    OR created_at < ?\n    OR (created_at = ? AND id < ?)\n  )\nORDER BY created_at DESC, id DESC\nLIMIT ?")
        .bind(params.business_id)
        .bind(&params.action)
        .bind(&params.action)
        .bind(&params.entity_type)
        .bind(&params.entity_type)
        .bind(params.employee_id)
        .bind(params.employee_id)
        .bind(params.start_date)
        .bind(params.start_date)
        .bind(params.end_date)
        .bind(params.end_date)
        .bind(params.cursor_created_at)
        .bind(params.cursor_created_at)
        .bind(params.cursor_created_at)
        .bind(params.cursor_id)
        .bind(params.limit)
        .execute(exec)
        .await
}
