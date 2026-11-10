//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `CreateTable` (Go `CreateTableParams`).
pub struct CreateTableParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub location_id: Option<i64>,
    pub name: String,
    pub capacity: Option<i32>,
}

/// `CreateTable`: `execresult`.
pub async fn create_table<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateTableParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO pos_tables (owner_id, business_id, location_id, name, capacity, status)\nVALUES (?, ?, ?, ?, ?, 'available')")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(params.location_id)
        .bind(&params.name)
        .bind(params.capacity)
        .execute(exec)
        .await
}

/// Params for `DeleteTable` (Go `DeleteTableParams`).
pub struct DeleteTableParams {
    pub id: i64,
    pub business_id: i64,
}

/// `DeleteTable`: `execresult`.
pub async fn delete_table<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DeleteTableParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("DELETE FROM pos_tables WHERE id = ? AND business_id = ?")
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `ListTablesByOwner` (Go `ListTablesByOwnerParams`).
pub struct ListTablesByOwnerParams {
    pub business_id: i64,
}

/// `ListTablesByOwner`: `many`.
pub async fn list_tables_by_owner<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListTablesByOwnerParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, location_id, name, capacity, status, current_session_order_id, qr_code_url, created_at, updated_at\nFROM pos_tables WHERE business_id = ? ORDER BY name")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `UpdateTable` (Go `UpdateTableParams`).
pub struct UpdateTableParams {
    pub location_id: Option<i64>,
    pub name: String,
    pub capacity: Option<i32>,
    pub status: String,
    pub id: i64,
    pub business_id: i64,
}

/// `UpdateTable`: `exec`.
pub async fn update_table<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateTableParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE pos_tables SET location_id = ?, name = ?, capacity = ?, status = ?\nWHERE id = ? AND business_id = ?")
        .bind(params.location_id)
        .bind(&params.name)
        .bind(params.capacity)
        .bind(&params.status)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `UpdateTableStatus` (Go `UpdateTableStatusParams`).
pub struct UpdateTableStatusParams {
    pub status: String,
    pub id: i64,
    pub business_id: i64,
}

/// `UpdateTableStatus`: `execresult`.
pub async fn update_table_status<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateTableStatusParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE pos_tables SET status = ? WHERE id = ? AND business_id = ?")
        .bind(&params.status)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}
