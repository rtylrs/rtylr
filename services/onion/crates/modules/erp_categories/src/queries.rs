//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `CountCategories` (Go `CountCategoriesParams`).
pub struct CountCategoriesParams {
    pub business_id: i64,
}

/// `CountCategories`: `one`.
pub async fn count_categories<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountCategoriesParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) FROM erp_categories WHERE business_id = ?")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `CreateCategory` (Go `CreateCategoryParams`).
pub struct CreateCategoryParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub name: String,
    pub description: Option<String>,
}

/// `CreateCategory`: `execresult`.
pub async fn create_category<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateCategoryParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO erp_categories (owner_id, business_id, name, description)\nVALUES (?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(&params.name)
        .bind(&params.description)
        .execute(exec)
        .await
}

/// Params for `DeleteCategory` (Go `DeleteCategoryParams`).
pub struct DeleteCategoryParams {
    pub business_id: i64,
    pub id: i64,
}

/// `DeleteCategory`: `exec`.
pub async fn delete_category<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DeleteCategoryParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("DELETE FROM erp_categories WHERE business_id = ? AND id = ?")
        .bind(params.business_id)
        .bind(params.id)
        .execute(exec)
        .await
}

/// Params for `ListCategories` (Go `ListCategoriesParams`).
pub struct ListCategoriesParams {
    pub business_id: i64,
}

/// `ListCategories`: `many`.
pub async fn list_categories<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListCategoriesParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, name, description, created_at, updated_at\nFROM erp_categories\nWHERE business_id = ?\nORDER BY name ASC, id ASC\nLIMIT 100")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `ListCategoriesPaginated` (Go `ListCategoriesPaginatedParams`).
pub struct ListCategoriesPaginatedParams {
    pub business_id: i64,
    pub limit: i32,
    pub offset: i32,
}

/// `ListCategoriesPaginated`: `many`.
pub async fn list_categories_paginated<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListCategoriesPaginatedParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, name, description, created_at, updated_at\nFROM erp_categories\nWHERE business_id = ?\nORDER BY name ASC, id ASC\nLIMIT ? OFFSET ?")
        .bind(params.business_id)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `UpdateCategory` (Go `UpdateCategoryParams`).
pub struct UpdateCategoryParams {
    pub name: String,
    pub description: Option<String>,
    pub business_id: i64,
    pub id: i64,
}

/// `UpdateCategory`: `exec`.
pub async fn update_category<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateCategoryParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query(
        "UPDATE erp_categories\nSET name = ?, description = ?\nWHERE business_id = ? AND id = ?",
    )
    .bind(&params.name)
    .bind(&params.description)
    .bind(params.business_id)
    .bind(params.id)
    .execute(exec)
    .await
}
