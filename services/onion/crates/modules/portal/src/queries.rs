//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `EnsureLayout` (Go `EnsureLayoutParams`).
pub struct EnsureLayoutParams {
    pub user_id: i64,
}

/// `EnsureLayout`: `exec`.
pub async fn ensure_layout<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &EnsureLayoutParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT IGNORE INTO portal_preferences (user_id, layout) VALUES (?, '{}')")
        .bind(params.user_id)
        .execute(exec)
        .await
}

/// Params for `GetLayout` (Go `GetLayoutParams`).
pub struct GetLayoutParams {
    pub user_id: i64,
}

/// `GetLayout`: `one`.
pub async fn get_layout<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetLayoutParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT layout FROM portal_preferences WHERE user_id = ?")
        .bind(params.user_id)
        .fetch_optional(exec)
        .await
}

/// Params for `LockLayout` (Go `LockLayoutParams`).
pub struct LockLayoutParams {
    pub user_id: i64,
}

/// `LockLayout`: `one`.
pub async fn lock_layout<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &LockLayoutParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT layout FROM portal_preferences WHERE user_id = ? FOR UPDATE")
        .bind(params.user_id)
        .fetch_optional(exec)
        .await
}

/// Params for `SetLayout` (Go `SetLayoutParams`).
pub struct SetLayoutParams {
    pub layout: String,
    pub user_id: i64,
}

/// `SetLayout`: `exec`.
pub async fn set_layout<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &SetLayoutParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE portal_preferences SET layout = ? WHERE user_id = ?")
        .bind(&params.layout)
        .bind(params.user_id)
        .execute(exec)
        .await
}
