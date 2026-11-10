//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `CreateTerminal` (Go `CreateTerminalParams`).
pub struct CreateTerminalParams {
    pub business_id: i64,
    pub branch_id: i64,
    pub name: String,
    pub device_id: Option<String>,
}

/// `CreateTerminal`: `execresult`.
pub async fn create_terminal<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateTerminalParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO pos_terminals (business_id, branch_id, name, device_id, is_active)\nVALUES (?, ?, ?, ?, TRUE)")
        .bind(params.business_id)
        .bind(params.branch_id)
        .bind(&params.name)
        .bind(&params.device_id)
        .execute(exec)
        .await
}

/// Params for `ListTerminalsByBusiness` (Go `ListTerminalsByBusinessParams`).
pub struct ListTerminalsByBusinessParams {
    pub business_id: i64,
}

/// `ListTerminalsByBusiness`: `many`.
pub async fn list_terminals_by_business<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListTerminalsByBusinessParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, business_id, branch_id, name, device_id, is_active, last_seen_at, created_at\nFROM pos_terminals WHERE business_id = ? ORDER BY name")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `UpdateTerminal` (Go `UpdateTerminalParams`).
pub struct UpdateTerminalParams {
    pub branch_id: i64,
    pub name: String,
    pub device_id: Option<String>,
    pub is_active: Option<bool>,
    pub id: i64,
    pub business_id: i64,
}

/// `UpdateTerminal`: `exec`.
pub async fn update_terminal<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateTerminalParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE pos_terminals SET branch_id = ?, name = ?, device_id = ?, is_active = ?\nWHERE id = ? AND business_id = ?")
        .bind(params.branch_id)
        .bind(&params.name)
        .bind(&params.device_id)
        .bind(params.is_active)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `UpdateTerminalHeartbeat` (Go `UpdateTerminalHeartbeatParams`).
pub struct UpdateTerminalHeartbeatParams {
    pub id: i64,
    pub business_id: i64,
}

/// `UpdateTerminalHeartbeat`: `execresult`.
pub async fn update_terminal_heartbeat<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateTerminalHeartbeatParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE pos_terminals SET last_seen_at = NOW()\nWHERE id = ? AND business_id = ?")
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}
