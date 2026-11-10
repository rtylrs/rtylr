//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `CountLoginHistory` (Go `CountLoginHistoryParams`).
pub struct CountLoginHistoryParams {
    pub user_id: i64,
}

/// `CountLoginHistory`: `one`.
pub async fn count_login_history<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountLoginHistoryParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) FROM login_history WHERE user_id = ?")
        .bind(params.user_id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetTwoFactorEnabled` (Go `GetTwoFactorEnabledParams`).
pub struct GetTwoFactorEnabledParams {
    pub id: i64,
}

/// `GetTwoFactorEnabled`: `one`.
pub async fn get_two_factor_enabled<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetTwoFactorEnabledParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COALESCE(two_factor_enabled, FALSE) FROM users WHERE id = ?")
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `InvalidateAllSessions` (Go `InvalidateAllSessionsParams`).
pub struct InvalidateAllSessionsParams {
    pub user_id: i64,
}

/// `InvalidateAllSessions`: `exec`.
pub async fn invalidate_all_sessions<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &InvalidateAllSessionsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query(
        "UPDATE user_sessions SET revoked_at = NOW() WHERE user_id = ? AND revoked_at IS NULL",
    )
    .bind(params.user_id)
    .execute(exec)
    .await
}

/// Params for `ListLoginHistory` (Go `ListLoginHistoryParams`).
pub struct ListLoginHistoryParams {
    pub user_id: i64,
    pub limit: i32,
    pub offset: i32,
}

/// `ListLoginHistory`: `many`.
pub async fn list_login_history<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListLoginHistoryParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, email, method, COALESCE(ip_address,''), COALESCE(user_agent,''), success, created_at\nFROM login_history\nWHERE user_id = ?\nORDER BY created_at DESC, id DESC\nLIMIT ? OFFSET ?")
        .bind(params.user_id)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `RecordLogin` (Go `RecordLoginParams`).
pub struct RecordLoginParams {
    pub user_id: i64,
    pub email: String,
    pub method: String,
    pub ip_address: Option<String>,
    pub user_agent: Option<String>,
    pub success: Option<bool>,
}

/// `RecordLogin`: `exec`.
pub async fn record_login<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &RecordLoginParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO login_history (user_id, email, method, ip_address, user_agent, success)\nVALUES (?, ?, ?, ?, ?, ?)")
        .bind(params.user_id)
        .bind(&params.email)
        .bind(&params.method)
        .bind(&params.ip_address)
        .bind(&params.user_agent)
        .bind(params.success)
        .execute(exec)
        .await
}

/// Params for `SetTwoFactor` (Go `SetTwoFactorParams`).
pub struct SetTwoFactorParams {
    pub two_factor_enabled: bool,
    pub id: i64,
}

/// `SetTwoFactor`: `exec`.
pub async fn set_two_factor<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &SetTwoFactorParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE users SET two_factor_enabled = ? WHERE id = ?")
        .bind(params.two_factor_enabled)
        .bind(params.id)
        .execute(exec)
        .await
}
