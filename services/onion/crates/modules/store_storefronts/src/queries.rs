//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `CountThemeForOwner` (Go `CountThemeForOwnerParams`).
pub struct CountThemeForOwnerParams {
    pub id: i64,
    pub business_id: i64,
}

/// `CountThemeForOwner`: `one`.
pub async fn count_theme_for_owner<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountThemeForOwnerParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) FROM store_themes\nWHERE id = ? AND business_id = ?")
        .bind(params.id)
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `CreateStorefront` (Go `CreateStorefrontParams`).
pub struct CreateStorefrontParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub name: String,
    pub domain: Option<String>,
    pub logo_url: Option<String>,
    pub theme_id: Option<i64>,
    pub is_published: Option<bool>,
}

/// `CreateStorefront`: `execresult`.
pub async fn create_storefront<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateStorefrontParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO store_sites (owner_id, business_id, name, domain, logo_url, theme_id, is_published)\nVALUES (?, ?, ?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(&params.name)
        .bind(&params.domain)
        .bind(&params.logo_url)
        .bind(params.theme_id)
        .bind(params.is_published)
        .execute(exec)
        .await
}

/// Params for `ListStorefronts` (Go `ListStorefrontsParams`).
pub struct ListStorefrontsParams {
    pub business_id: i64,
}

/// `ListStorefronts`: `many`.
pub async fn list_storefronts<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListStorefrontsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, name, domain, logo_url, theme_id, is_published, created_at, updated_at\nFROM store_sites\nWHERE business_id = ?\nORDER BY name ASC\nLIMIT 200")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `UpdateStorefront` (Go `UpdateStorefrontParams`).
pub struct UpdateStorefrontParams {
    pub name: String,
    pub domain: Option<String>,
    pub logo_url: Option<String>,
    pub theme_id: Option<i64>,
    pub is_published: Option<bool>,
    pub business_id: i64,
    pub id: i64,
}

/// `UpdateStorefront`: `exec`.
pub async fn update_storefront<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateStorefrontParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE store_sites\nSET name = ?, domain = ?, logo_url = ?, theme_id = ?, is_published = ?\nWHERE business_id = ? AND id = ?")
        .bind(&params.name)
        .bind(&params.domain)
        .bind(&params.logo_url)
        .bind(params.theme_id)
        .bind(params.is_published)
        .bind(params.business_id)
        .bind(params.id)
        .execute(exec)
        .await
}
