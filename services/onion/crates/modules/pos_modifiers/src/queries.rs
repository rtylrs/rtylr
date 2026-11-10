//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `CreateModifier` (Go `CreateModifierParams`).
pub struct CreateModifierParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub name: String,
    pub price_delta: String,
    pub group_name: Option<String>,
    pub is_required: Option<bool>,
}

/// `CreateModifier`: `execresult`.
pub async fn create_modifier<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateModifierParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO pos_modifiers (owner_id, business_id, name, price_delta, group_name, is_required)\nVALUES (?, ?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(&params.name)
        .bind(&params.price_delta)
        .bind(&params.group_name)
        .bind(params.is_required)
        .execute(exec)
        .await
}

/// Params for `DeleteModifier` (Go `DeleteModifierParams`).
pub struct DeleteModifierParams {
    pub id: i64,
    pub business_id: i64,
}

/// `DeleteModifier`: `execresult`.
pub async fn delete_modifier<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DeleteModifierParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("DELETE FROM pos_modifiers WHERE id = ? AND business_id = ?")
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `ListModifiersByOwner` (Go `ListModifiersByOwnerParams`).
pub struct ListModifiersByOwnerParams {
    pub business_id: i64,
}

/// `ListModifiersByOwner`: `many`.
pub async fn list_modifiers_by_owner<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListModifiersByOwnerParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, name, price_delta, group_name, is_required, created_at, updated_at\nFROM pos_modifiers WHERE business_id = ? ORDER BY group_name, name")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `UpdateModifier` (Go `UpdateModifierParams`).
pub struct UpdateModifierParams {
    pub name: String,
    pub price_delta: String,
    pub group_name: Option<String>,
    pub is_required: Option<bool>,
    pub id: i64,
    pub business_id: i64,
}

/// `UpdateModifier`: `execresult`.
pub async fn update_modifier<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateModifierParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE pos_modifiers SET name = ?, price_delta = ?, group_name = ?, is_required = ?\nWHERE id = ? AND business_id = ?")
        .bind(&params.name)
        .bind(&params.price_delta)
        .bind(&params.group_name)
        .bind(params.is_required)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}
