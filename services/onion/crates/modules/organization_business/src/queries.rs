//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `CheckSubBusinessOwnership` (Go `CheckSubBusinessOwnershipParams`).
pub struct CheckSubBusinessOwnershipParams {
    pub id: i64,
    pub owner_id: i64,
}

/// `CheckSubBusinessOwnership`: `one`.
pub async fn check_sub_business_ownership<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CheckSubBusinessOwnershipParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query(
        "SELECT COUNT(*) AS cnt FROM businesses WHERE id = ? AND owner_id = ? AND is_active = TRUE",
    )
    .bind(params.id)
    .bind(params.owner_id)
    .fetch_optional(exec)
    .await
}

/// Params for `CountActiveOwnedBusinesses` (Go `CountActiveOwnedBusinessesParams`).
pub struct CountActiveOwnedBusinessesParams {
    pub owner_id: i64,
}

/// `CountActiveOwnedBusinesses`: `one`.
pub async fn count_active_owned_businesses<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountActiveOwnedBusinessesParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) AS cnt\nFROM businesses\nWHERE owner_id = ? AND is_active = TRUE")
        .bind(params.owner_id)
        .fetch_optional(exec)
        .await
}

/// Params for `DeactivateBusiness` (Go `DeactivateBusinessParams`).
pub struct DeactivateBusinessParams {
    pub id: i64,
    pub owner_id: i64,
}

/// `DeactivateBusiness`: `execresult`.
pub async fn deactivate_business<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DeactivateBusinessParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE businesses\nSET is_active = FALSE\nWHERE id = ? AND owner_id = ? AND is_active = TRUE")
        .bind(params.id)
        .bind(params.owner_id)
        .execute(exec)
        .await
}

/// Params for `GetBusinessOwnerID` (Go `GetBusinessOwnerIDParams`).
pub struct GetBusinessOwnerIDParams {
    pub id: i64,
}

/// `GetBusinessOwnerID`: `one`.
pub async fn get_business_owner_id<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetBusinessOwnerIDParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT owner_id FROM businesses WHERE id = ? AND is_active = TRUE LIMIT 1")
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetBusinessProfileByBusiness` (Go `GetBusinessProfileByBusinessParams`).
pub struct GetBusinessProfileByBusinessParams {
    pub id: i64,
}

/// `GetBusinessProfileByBusiness`: `one`.
pub async fn get_business_profile_by_business<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetBusinessProfileByBusinessParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT\n    b.id,\n    b.name,\n    COALESCE(NULLIF(c.site_name, ''), b.name) AS site_name,\n    COALESCE(NULLIF(c.logo_url, ''), COALESCE(b.logo_url, '')) AS logo_url,\n    COALESCE(c.contact_email, '') AS contact_email,\n    COALESCE(c.contact_number, '') AS contact_number,\n    COALESCE(b.currency, 'USD') AS currency,\n    COALESCE(b.timezone, 'UTC') AS timezone\nFROM businesses b\nLEFT JOIN config c ON c.business_id = b.id\nWHERE b.id = ? AND b.is_active = TRUE\nLIMIT 1")
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetPublicBusiness` (Go `GetPublicBusinessParams`).
pub struct GetPublicBusinessParams {
    pub id: i64,
}

/// `GetPublicBusiness`: `one`.
pub async fn get_public_business<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetPublicBusinessParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, name, COALESCE(slug, '') AS slug, COALESCE(logo_url, '') AS logo_url,\n       COALESCE(currency, 'USD') AS currency, COALESCE(timezone, 'UTC') AS timezone\nFROM businesses\nWHERE id = ? AND is_active = TRUE\nORDER BY id ASC\nLIMIT 1")
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetReportingPreferences` (Go `GetReportingPreferencesParams`).
pub struct GetReportingPreferencesParams {
    pub id: i64,
}

/// `GetReportingPreferences`: `one`.
pub async fn get_reporting_preferences<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetReportingPreferencesParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COALESCE(NULLIF(UPPER(TRIM(currency)), ''), 'USD') AS currency,\n       COALESCE(NULLIF(TRIM(timezone), ''), 'UTC') AS timezone\nFROM businesses WHERE id = ? AND is_active = TRUE")
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `InsertSubBusiness` (Go `InsertSubBusinessParams`).
pub struct InsertSubBusinessParams {
    pub name: String,
    pub owner_id: i64,
    pub parent_id: Option<i64>,
}

/// `InsertSubBusiness`: `execresult`.
pub async fn insert_sub_business<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &InsertSubBusinessParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO businesses (name, owner_id, parent_id) VALUES (?, ?, ?)")
        .bind(&params.name)
        .bind(params.owner_id)
        .bind(params.parent_id)
        .execute(exec)
        .await
}

/// Params for `ListUserBusinesses` (Go `ListUserBusinessesParams`).
pub struct ListUserBusinessesParams {
    pub user_id: i64,
}

/// `ListUserBusinesses`: `many`.
pub async fn list_user_businesses<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListUserBusinessesParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT b.id, b.name, COALESCE(b.logo_url, '') AS logo_url, COALESCE(b.parent_id, 0) AS parent_id,\n       b.is_active, (b.owner_id = ?) AS is_owned\nFROM businesses b\nWHERE b.is_active = TRUE\n  AND (b.owner_id = ? OR EXISTS (\n    SELECT 1\n    FROM authorization_business_members member\n    WHERE member.user_id = ?\n      AND member.business_id = b.id\n      AND member.status = 'active'\n  ))\nORDER BY b.id ASC\nLIMIT 100")
        .bind(params.user_id)
        .bind(params.user_id)
        .bind(params.user_id)
        .execute(exec)
        .await
}

/// Params for `LookupBusinessBySlug` (Go `LookupBusinessBySlugParams`).
pub struct LookupBusinessBySlugParams {
    pub slug: Option<String>,
}

/// `LookupBusinessBySlug`: `one`.
pub async fn lookup_business_by_slug<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &LookupBusinessBySlugParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, name, COALESCE(slug, '') AS slug, COALESCE(logo_url, '') AS logo_url,\n       COALESCE(currency, 'USD') AS currency, COALESCE(timezone, 'UTC') AS timezone,\n       COALESCE(owner_id, 0) AS owner_id\nFROM businesses\nWHERE slug = ? AND is_active = TRUE\nLIMIT 1")
        .bind(&params.slug)
        .fetch_optional(exec)
        .await
}

/// Params for `RenameBusiness` (Go `RenameBusinessParams`).
pub struct RenameBusinessParams {
    pub name: String,
    pub id: i64,
    pub owner_id: i64,
}

/// `RenameBusiness`: `execresult`.
pub async fn rename_business<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &RenameBusinessParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query(
        "UPDATE businesses\nSET name = ?\nWHERE id = ? AND owner_id = ? AND is_active = TRUE",
    )
    .bind(&params.name)
    .bind(params.id)
    .bind(params.owner_id)
    .execute(exec)
    .await
}

/// Params for `SetBusinessLogo` (Go `SetBusinessLogoParams`).
pub struct SetBusinessLogoParams {
    pub logo_url: Option<String>,
    pub id: i64,
    pub owner_id: i64,
}

/// `SetBusinessLogo`: `execresult`.
pub async fn set_business_logo<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &SetBusinessLogoParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query(
        "UPDATE businesses\nSET logo_url = ?\nWHERE id = ? AND owner_id = ? AND is_active = TRUE",
    )
    .bind(&params.logo_url)
    .bind(params.id)
    .bind(params.owner_id)
    .execute(exec)
    .await
}

/// Params for `UpdateBusinessSlug` (Go `UpdateBusinessSlugParams`).
pub struct UpdateBusinessSlugParams {
    pub slug: Option<String>,
    pub id: i64,
    pub owner_id: i64,
}

/// `UpdateBusinessSlug`: `execresult`.
pub async fn update_business_slug<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateBusinessSlugParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query(
        "UPDATE businesses\nSET slug = ?\nWHERE id = ? AND owner_id = ? AND is_active = TRUE",
    )
    .bind(&params.slug)
    .bind(params.id)
    .bind(params.owner_id)
    .execute(exec)
    .await
}
