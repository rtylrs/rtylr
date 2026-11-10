//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `CreateFloor` (Go `CreateFloorParams`).
pub struct CreateFloorParams {
    pub business_id: i64,
    pub branch_id: i64,
    pub name: String,
    pub sort_order: Option<i32>,
}

/// `CreateFloor`: `execresult`.
pub async fn create_floor<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateFloorParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO pos_floors (business_id, branch_id, name, sort_order, is_active)\nVALUES (?, ?, ?, ?, TRUE)")
        .bind(params.business_id)
        .bind(params.branch_id)
        .bind(&params.name)
        .bind(params.sort_order)
        .execute(exec)
        .await
}

/// Params for `ListFloorsByBranch` (Go `ListFloorsByBranchParams`).
pub struct ListFloorsByBranchParams {
    pub business_id: i64,
    pub branch_id: i64,
}

/// `ListFloorsByBranch`: `many`.
pub async fn list_floors_by_branch<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListFloorsByBranchParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, business_id, branch_id, name, sort_order, is_active, created_at\nFROM pos_floors WHERE business_id = ? AND branch_id = ? AND is_active = TRUE ORDER BY sort_order\nLIMIT 200")
        .bind(params.business_id)
        .bind(params.branch_id)
        .execute(exec)
        .await
}

/// Params for `SoftDeleteFloor` (Go `SoftDeleteFloorParams`).
pub struct SoftDeleteFloorParams {
    pub id: i64,
    pub business_id: i64,
}

/// `SoftDeleteFloor`: `exec`.
pub async fn soft_delete_floor<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &SoftDeleteFloorParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE pos_floors SET is_active = FALSE\nWHERE id = ? AND business_id = ?")
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `UpdateFloor` (Go `UpdateFloorParams`).
pub struct UpdateFloorParams {
    pub name: String,
    pub sort_order: Option<i32>,
    pub id: i64,
    pub business_id: i64,
}

/// `UpdateFloor`: `exec`.
pub async fn update_floor<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateFloorParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE pos_floors SET name = ?, sort_order = ?\nWHERE id = ? AND business_id = ?")
        .bind(&params.name)
        .bind(params.sort_order)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}
