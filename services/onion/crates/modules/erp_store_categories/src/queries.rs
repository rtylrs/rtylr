//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `CreateStoreCategory` (Go `CreateStoreCategoryParams`).
pub struct CreateStoreCategoryParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub name: String,
    pub slug: Option<String>,
    pub description: Option<String>,
    pub image_url: Option<String>,
    pub parent_id: Option<i64>,
    pub sort_order: Option<i32>,
}

/// `CreateStoreCategory`: `execresult`.
pub async fn create_store_category<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateStoreCategoryParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO store_categories (owner_id, business_id, name, slug, description, image_url, parent_id, sort_order)\nVALUES (?, ?, ?, ?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(&params.name)
        .bind(&params.slug)
        .bind(&params.description)
        .bind(&params.image_url)
        .bind(params.parent_id)
        .bind(params.sort_order)
        .execute(exec)
        .await
}

/// Params for `DeleteStoreCategory` (Go `DeleteStoreCategoryParams`).
pub struct DeleteStoreCategoryParams {
    pub business_id: i64,
    pub id: i64,
}

/// `DeleteStoreCategory`: `exec`.
pub async fn delete_store_category<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DeleteStoreCategoryParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("DELETE FROM store_categories WHERE business_id = ? AND id = ?")
        .bind(params.business_id)
        .bind(params.id)
        .execute(exec)
        .await
}

/// Params for `GetStoreCategory` (Go `GetStoreCategoryParams`).
pub struct GetStoreCategoryParams {
    pub business_id: i64,
    pub id: i64,
}

/// `GetStoreCategory`: `one`.
pub async fn get_store_category<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetStoreCategoryParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, name, slug, description, image_url, parent_id, sort_order, created_at\nFROM store_categories WHERE business_id = ? AND id = ?")
        .bind(params.business_id)
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `ListStoreCategories` (Go `ListStoreCategoriesParams`).
pub struct ListStoreCategoriesParams {
    pub business_id: i64,
}

/// `ListStoreCategories`: `many`.
pub async fn list_store_categories<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListStoreCategoriesParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, name, slug, description, image_url, parent_id, sort_order, created_at\nFROM store_categories WHERE business_id = ? ORDER BY sort_order ASC\nLIMIT 200")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `UpdateStoreCategory` (Go `UpdateStoreCategoryParams`).
pub struct UpdateStoreCategoryParams {
    pub name: String,
    pub slug: Option<String>,
    pub description: Option<String>,
    pub image_url: Option<String>,
    pub parent_id: Option<i64>,
    pub sort_order: Option<i32>,
    pub business_id: i64,
    pub id: i64,
}

/// `UpdateStoreCategory`: `exec`.
pub async fn update_store_category<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateStoreCategoryParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE store_categories SET name = ?, slug = ?, description = ?, image_url = ?, parent_id = ?, sort_order = ?\nWHERE business_id = ? AND id = ?")
        .bind(&params.name)
        .bind(&params.slug)
        .bind(&params.description)
        .bind(&params.image_url)
        .bind(params.parent_id)
        .bind(params.sort_order)
        .bind(params.business_id)
        .bind(params.id)
        .execute(exec)
        .await
}
