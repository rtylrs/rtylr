//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `CountLocationsByBusiness` (Go `CountLocationsByBusinessParams`).
pub struct CountLocationsByBusinessParams {
    pub business_id: i64,
}

/// `CountLocationsByBusiness`: `one`.
pub async fn count_locations_by_business<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountLocationsByBusinessParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) FROM pos_locations WHERE business_id = ?")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `CreateLocation` (Go `CreateLocationParams`).
pub struct CreateLocationParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub name: String,
    pub address: Option<String>,
    pub phone: Option<String>,
    pub is_active: Option<bool>,
}

/// `CreateLocation`: `execresult`.
pub async fn create_location<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateLocationParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO pos_locations (owner_id, business_id, name, address, phone, is_active)\nVALUES (?, ?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(&params.name)
        .bind(&params.address)
        .bind(&params.phone)
        .bind(params.is_active)
        .execute(exec)
        .await
}

/// Params for `EnsureLocationLock` (Go `EnsureLocationLockParams`).
pub struct EnsureLocationLockParams {
    pub business_id: i64,
}

/// `EnsureLocationLock`: `exec`.
pub async fn ensure_location_lock<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &EnsureLocationLockParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO pos_location_creation_locks (business_id) VALUES (?) ON DUPLICATE KEY UPDATE business_id = business_id")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `ListLocationsByOwner` (Go `ListLocationsByOwnerParams`).
pub struct ListLocationsByOwnerParams {
    pub business_id: i64,
}

/// `ListLocationsByOwner`: `many`.
pub async fn list_locations_by_owner<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListLocationsByOwnerParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, name, address, phone, is_active, created_at\nFROM pos_locations WHERE business_id = ? ORDER BY name")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `LockLocationCreation` (Go `LockLocationCreationParams`).
pub struct LockLocationCreationParams {
    pub business_id: i64,
}

/// `LockLocationCreation`: `one`.
pub async fn lock_location_creation<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &LockLocationCreationParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query(
        "SELECT business_id FROM pos_location_creation_locks WHERE business_id = ? FOR UPDATE",
    )
    .bind(params.business_id)
    .fetch_optional(exec)
    .await
}

/// Params for `StockBranchExists` (Go `StockBranchExistsParams`).
pub struct StockBranchExistsParams {
    pub business_id: i64,
    pub id: i64,
}

/// `StockBranchExists`: `one`.
pub async fn stock_branch_exists<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &StockBranchExistsParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) > 0 FROM pos_locations WHERE business_id = ? AND id = ?")
        .bind(params.business_id)
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `UpdateLocation` (Go `UpdateLocationParams`).
pub struct UpdateLocationParams {
    pub name: String,
    pub address: Option<String>,
    pub phone: Option<String>,
    pub is_active: Option<bool>,
    pub id: i64,
    pub business_id: i64,
}

/// `UpdateLocation`: `execresult`.
pub async fn update_location<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateLocationParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE pos_locations SET name = ?, address = ?, phone = ?, is_active = COALESCE(?, is_active)\nWHERE id = ? AND business_id = ?")
        .bind(&params.name)
        .bind(&params.address)
        .bind(&params.phone)
        .bind(params.is_active)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}
