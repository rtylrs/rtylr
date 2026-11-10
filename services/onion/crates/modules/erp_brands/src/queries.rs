//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `CountBrands` (Go `CountBrandsParams`).
pub struct CountBrandsParams {
    pub business_id: i64,
}

/// `CountBrands`: `one`.
pub async fn count_brands<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountBrandsParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) FROM erp_brands WHERE business_id = ?")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `CreateBrand` (Go `CreateBrandParams`).
pub struct CreateBrandParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub name: String,
    pub logo_url: Option<String>,
    pub description: Option<String>,
}

/// `CreateBrand`: `execresult`.
pub async fn create_brand<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateBrandParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO erp_brands (owner_id, business_id, name, logo_url, description)\nVALUES (?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(&params.name)
        .bind(&params.logo_url)
        .bind(&params.description)
        .execute(exec)
        .await
}

/// Params for `DeleteBrand` (Go `DeleteBrandParams`).
pub struct DeleteBrandParams {
    pub business_id: i64,
    pub id: i64,
}

/// `DeleteBrand`: `exec`.
pub async fn delete_brand<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DeleteBrandParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("DELETE FROM erp_brands WHERE business_id = ? AND id = ?")
        .bind(params.business_id)
        .bind(params.id)
        .execute(exec)
        .await
}

/// Params for `ListBrands` (Go `ListBrandsParams`).
pub struct ListBrandsParams {
    pub business_id: i64,
}

/// `ListBrands`: `many`.
pub async fn list_brands<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListBrandsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("WITH brand_page AS (\n    SELECT b.id, b.owner_id, b.business_id, b.name, b.logo_url, b.description, b.created_at, b.updated_at\n    FROM erp_brands b\n    WHERE b.business_id = ?\n), product_counts AS (\n    -- cross-schema read from erp_products (products module) — include the number of products assigned to each brand.\n    SELECT p.brand_id, COUNT(*) AS product_count\n    FROM erp_products p FORCE INDEX (idx_erp_products_business_brand)\n    JOIN brand_page b ON b.id = p.brand_id AND b.business_id = p.business_id\n    GROUP BY p.brand_id\n)\nSELECT b.id, b.owner_id, b.business_id, b.name, b.logo_url, b.description,\n       COALESCE(pc.product_count, 0) AS product_count,\n       b.created_at, b.updated_at\nFROM brand_page b\nLEFT JOIN product_counts pc ON pc.brand_id = b.id\nORDER BY b.name ASC, b.id ASC")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `ListBrandsPaginated` (Go `ListBrandsPaginatedParams`).
pub struct ListBrandsPaginatedParams {
    pub business_id: i64,
    pub limit: i32,
    pub offset: i32,
}

/// `ListBrandsPaginated`: `many`.
pub async fn list_brands_paginated<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListBrandsPaginatedParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("WITH brand_page AS (\n    SELECT b.id, b.owner_id, b.business_id, b.name, b.logo_url, b.description, b.created_at, b.updated_at\n    FROM erp_brands b\n    WHERE b.business_id = ?\n    ORDER BY b.name ASC, b.id ASC\n    LIMIT ? OFFSET ?\n), product_counts AS (\n    -- cross-schema read from erp_products (products module) — include the number of products assigned to each brand.\n    SELECT p.brand_id, COUNT(*) AS product_count\n    FROM erp_products p FORCE INDEX (idx_erp_products_business_brand)\n    JOIN brand_page b ON b.id = p.brand_id AND b.business_id = p.business_id\n    GROUP BY p.brand_id\n)\nSELECT b.id, b.owner_id, b.business_id, b.name, b.logo_url, b.description,\n       COALESCE(pc.product_count, 0) AS product_count,\n       b.created_at, b.updated_at\nFROM brand_page b\nLEFT JOIN product_counts pc ON pc.brand_id = b.id\nORDER BY b.name ASC, b.id ASC")
        .bind(params.business_id)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `UpdateBrand` (Go `UpdateBrandParams`).
pub struct UpdateBrandParams {
    pub name: String,
    pub logo_url: Option<String>,
    pub description: Option<String>,
    pub business_id: i64,
    pub id: i64,
}

/// `UpdateBrand`: `exec`.
pub async fn update_brand<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateBrandParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE erp_brands\nSET name = ?, logo_url = ?, description = ?\nWHERE business_id = ? AND id = ?")
        .bind(&params.name)
        .bind(&params.logo_url)
        .bind(&params.description)
        .bind(params.business_id)
        .bind(params.id)
        .execute(exec)
        .await
}
