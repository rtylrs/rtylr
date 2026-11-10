//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `CreateMenuBanner` (Go `CreateMenuBannerParams`).
pub struct CreateMenuBannerParams {
    pub business_id: i64,
    pub title: String,
    pub body: Option<String>,
    pub image_url: Option<String>,
    pub link_url: Option<String>,
    pub position: i32,
    pub is_active: bool,
    pub starts_at: Option<chrono::DateTime<chrono::Utc>>,
    pub ends_at: Option<chrono::DateTime<chrono::Utc>>,
}

/// `CreateMenuBanner`: `execresult`.
pub async fn create_menu_banner<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateMenuBannerParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO hub_menu_banners (\n    business_id,\n    title,\n    body,\n    image_url,\n    link_url,\n    position,\n    is_active,\n    starts_at,\n    ends_at\n) VALUES (\n    ?, ?, ?, ?, ?, ?, ?, ?, ?\n)")
        .bind(params.business_id)
        .bind(&params.title)
        .bind(&params.body)
        .bind(&params.image_url)
        .bind(&params.link_url)
        .bind(params.position)
        .bind(params.is_active)
        .bind(params.starts_at)
        .bind(params.ends_at)
        .execute(exec)
        .await
}

/// Params for `DeleteMenuBanner` (Go `DeleteMenuBannerParams`).
pub struct DeleteMenuBannerParams {
    pub business_id: i64,
    pub id: i64,
}

/// `DeleteMenuBanner`: `exec`.
pub async fn delete_menu_banner<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DeleteMenuBannerParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("DELETE FROM hub_menu_banners\nWHERE business_id = ? AND id = ?")
        .bind(params.business_id)
        .bind(params.id)
        .execute(exec)
        .await
}

/// Params for `GetHubNavigationMenu` (Go `GetHubNavigationMenuParams`).
pub struct GetHubNavigationMenuParams {
    pub business_id: i64,
}

/// `GetHubNavigationMenu`: `one`.
pub async fn get_hub_navigation_menu<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetHubNavigationMenuParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT items\nFROM hub_navigation_menus\nWHERE business_id = ?\nORDER BY business_id ASC\nLIMIT 1")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetMenuBanner` (Go `GetMenuBannerParams`).
pub struct GetMenuBannerParams {
    pub business_id: i64,
    pub id: i64,
}

/// `GetMenuBanner`: `one`.
pub async fn get_menu_banner<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetMenuBannerParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT\n    id,\n    business_id,\n    title,\n    COALESCE(body, '') AS body,\n    COALESCE(image_url, '') AS image_url,\n    COALESCE(link_url, '') AS link_url,\n    position,\n    is_active,\n    starts_at,\n    ends_at,\n    created_at,\n    updated_at\nFROM hub_menu_banners\nWHERE business_id = ? AND id = ?\nORDER BY id ASC\nLIMIT 1")
        .bind(params.business_id)
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `ListMenuBanners` (Go `ListMenuBannersParams`).
pub struct ListMenuBannersParams {
    pub business_id: i64,
}

/// `ListMenuBanners`: `many`.
pub async fn list_menu_banners<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListMenuBannersParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT\n    id,\n    business_id,\n    title,\n    COALESCE(body, '') AS body,\n    COALESCE(image_url, '') AS image_url,\n    COALESCE(link_url, '') AS link_url,\n    position,\n    is_active,\n    starts_at,\n    ends_at,\n    created_at,\n    updated_at\nFROM hub_menu_banners\nWHERE business_id = ?\nORDER BY position ASC, id ASC\nLIMIT 200")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `UpdateMenuBanner` (Go `UpdateMenuBannerParams`).
pub struct UpdateMenuBannerParams {
    pub title: String,
    pub body: Option<String>,
    pub image_url: Option<String>,
    pub link_url: Option<String>,
    pub position: i32,
    pub is_active: bool,
    pub starts_at: Option<chrono::DateTime<chrono::Utc>>,
    pub ends_at: Option<chrono::DateTime<chrono::Utc>>,
    pub business_id: i64,
    pub id: i64,
}

/// `UpdateMenuBanner`: `exec`.
pub async fn update_menu_banner<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateMenuBannerParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE hub_menu_banners\nSET title = ?,\n    body = ?,\n    image_url = ?,\n    link_url = ?,\n    position = ?,\n    is_active = ?,\n    starts_at = ?,\n    ends_at = ?,\n    updated_at = CURRENT_TIMESTAMP\nWHERE business_id = ? AND id = ?")
        .bind(&params.title)
        .bind(&params.body)
        .bind(&params.image_url)
        .bind(&params.link_url)
        .bind(params.position)
        .bind(params.is_active)
        .bind(params.starts_at)
        .bind(params.ends_at)
        .bind(params.business_id)
        .bind(params.id)
        .execute(exec)
        .await
}

/// Params for `UpsertHubNavigationMenu` (Go `UpsertHubNavigationMenuParams`).
pub struct UpsertHubNavigationMenuParams {
    pub business_id: i64,
    pub items: String,
}

/// `UpsertHubNavigationMenu`: `exec`.
pub async fn upsert_hub_navigation_menu<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpsertHubNavigationMenuParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO hub_navigation_menus (\n    business_id,\n    items\n) VALUES (\n    ?, ?\n)\nON DUPLICATE KEY UPDATE\n    items = VALUES(items),\n    updated_at = CURRENT_TIMESTAMP")
        .bind(params.business_id)
        .bind(&params.items)
        .execute(exec)
        .await
}
