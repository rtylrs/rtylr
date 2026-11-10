//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `InsertManagerPIN` (Go `InsertManagerPINParams`).
pub struct InsertManagerPINParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub staff_id: i64,
    pub pin_hash: Option<String>,
}

/// `InsertManagerPIN`: `exec`.
pub async fn insert_manager_pin<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &InsertManagerPINParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO pos_manager_pins (owner_id, business_id, staff_id, pin_hash) VALUES (?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(params.staff_id)
        .bind(&params.pin_hash)
        .execute(exec)
        .await
}

/// Params for `ListManagerPINHashes` (Go `ListManagerPINHashesParams`).
pub struct ListManagerPINHashesParams {
    pub business_id: i64,
}

/// `ListManagerPINHashes`: `many`.
pub async fn list_manager_pin_hashes<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListManagerPINHashesParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT owner_id, staff_id, pin_hash FROM pos_manager_pins WHERE business_id = ? AND pin_hash IS NOT NULL AND pin_hash <> '' ORDER BY id")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `UpdateManagerPINHash` (Go `UpdateManagerPINHashParams`).
pub struct UpdateManagerPINHashParams {
    pub pin_hash: Option<String>,
    pub business_id: i64,
    pub staff_id: i64,
}

/// `UpdateManagerPINHash`: `execrows`.
pub async fn update_manager_pin_hash<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateManagerPINHashParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE pos_manager_pins SET pin_hash = ? WHERE business_id = ? AND staff_id = ?")
        .bind(&params.pin_hash)
        .bind(params.business_id)
        .bind(params.staff_id)
        .execute(exec)
        .await
}
