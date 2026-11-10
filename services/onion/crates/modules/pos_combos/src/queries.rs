//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `CreateCombo` (Go `CreateComboParams`).
pub struct CreateComboParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub name: String,
    pub description: Option<String>,
    pub price: String,
    pub items: String,
    pub is_active: Option<bool>,
}

/// `CreateCombo`: `execresult`.
pub async fn create_combo<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateComboParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO pos_combos (owner_id, business_id, name, description, price, items, is_active)\nVALUES (?, ?, ?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(&params.name)
        .bind(&params.description)
        .bind(&params.price)
        .bind(&params.items)
        .bind(params.is_active)
        .execute(exec)
        .await
}

/// Params for `DeleteCombo` (Go `DeleteComboParams`).
pub struct DeleteComboParams {
    pub id: i64,
    pub business_id: i64,
}

/// `DeleteCombo`: `execresult`.
pub async fn delete_combo<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DeleteComboParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("DELETE FROM pos_combos WHERE id = ? AND business_id = ?")
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `ListCombosByOwner` (Go `ListCombosByOwnerParams`).
pub struct ListCombosByOwnerParams {
    pub business_id: i64,
}

/// `ListCombosByOwner`: `many`.
pub async fn list_combos_by_owner<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListCombosByOwnerParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, name, description, price, items, is_active, created_at, updated_at\nFROM pos_combos WHERE business_id = ? ORDER BY name\nLIMIT 200")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `UpdateCombo` (Go `UpdateComboParams`).
pub struct UpdateComboParams {
    pub name: String,
    pub description: Option<String>,
    pub price: String,
    pub items: String,
    pub is_active: Option<bool>,
    pub id: i64,
    pub business_id: i64,
}

/// `UpdateCombo`: `execresult`.
pub async fn update_combo<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateComboParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE pos_combos SET name = ?, description = ?, price = ?, items = ?, is_active = ?\nWHERE id = ? AND business_id = ?")
        .bind(&params.name)
        .bind(&params.description)
        .bind(&params.price)
        .bind(&params.items)
        .bind(params.is_active)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}
