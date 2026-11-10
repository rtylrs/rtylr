//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `CreateCashMovement` (Go `CreateCashMovementParams`).
pub struct CreateCashMovementParams {
    pub business_id: i64,
    pub branch_id: i64,
    pub terminal_id: Option<i64>,
    pub session_id: i64,
    pub employee_id: i64,
    pub r#type: String,
    pub amount: String,
    pub reason: Option<String>,
    pub approved_by: Option<i64>,
}

/// `CreateCashMovement`: `execresult`.
pub async fn create_cash_movement<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateCashMovementParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO pos_cash_movements (business_id, branch_id, terminal_id, session_id, employee_id, type, amount, reason, approved_by)\nVALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)")
        .bind(params.business_id)
        .bind(params.branch_id)
        .bind(params.terminal_id)
        .bind(params.session_id)
        .bind(params.employee_id)
        .bind(&params.r#type)
        .bind(&params.amount)
        .bind(&params.reason)
        .bind(params.approved_by)
        .execute(exec)
        .await
}

/// Params for `ListCashMovementsBySession` (Go `ListCashMovementsBySessionParams`).
pub struct ListCashMovementsBySessionParams {
    pub business_id: i64,
    pub session_id: i64,
}

/// `ListCashMovementsBySession`: `many`.
pub async fn list_cash_movements_by_session<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListCashMovementsBySessionParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, business_id, branch_id, terminal_id, session_id, employee_id,\n       type, amount, reason, approved_by, created_at\nFROM pos_cash_movements\nWHERE business_id = ? AND session_id = ?\nORDER BY created_at DESC\nLIMIT 200")
        .bind(params.business_id)
        .bind(params.session_id)
        .execute(exec)
        .await
}
