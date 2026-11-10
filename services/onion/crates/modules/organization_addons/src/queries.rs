//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `DisableAddon` (Go `DisableAddonParams`).
pub struct DisableAddonParams {
    pub business_id: i64,
    pub addon_id: String,
}

/// `DisableAddon`: `exec`.
pub async fn disable_addon<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DisableAddonParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE business_addons SET is_enabled = false, disabled_at = NOW()\nWHERE business_id = ? AND addon_id = ?")
        .bind(params.business_id)
        .bind(&params.addon_id)
        .execute(exec)
        .await
}

/// Params for `EnableAddon` (Go `EnableAddonParams`).
pub struct EnableAddonParams {
    pub business_id: i64,
    pub addon_id: String,
}

/// `EnableAddon`: `exec`.
pub async fn enable_addon<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &EnableAddonParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO business_addons (business_id, addon_id, is_enabled, enabled_at)\nVALUES (?, ?, true, NOW())\nON DUPLICATE KEY UPDATE is_enabled = true, enabled_at = NOW(), disabled_at = NULL")
        .bind(params.business_id)
        .bind(&params.addon_id)
        .execute(exec)
        .await
}

/// Params for `HasActiveAddonGrant` (Go `HasActiveAddonGrantParams`).
pub struct HasActiveAddonGrantParams {
    pub business_id: i64,
    pub addon_id: String,
}

/// `HasActiveAddonGrant`: `one`.
pub async fn has_active_addon_grant<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &HasActiveAddonGrantParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) FROM manual_addon_grants\nWHERE business_id = ? AND addon_id = ? AND revoked_at IS NULL\n  AND (expires_at IS NULL OR expires_at > CURRENT_TIMESTAMP)")
        .bind(params.business_id)
        .bind(&params.addon_id)
        .fetch_optional(exec)
        .await
}

/// Params for `IsAddonEnabled` (Go `IsAddonEnabledParams`).
pub struct IsAddonEnabledParams {
    pub business_id: i64,
    pub addon_id: String,
}

/// `IsAddonEnabled`: `one`.
pub async fn is_addon_enabled<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &IsAddonEnabledParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) FROM business_addons\nWHERE business_id = ? AND addon_id = ? AND is_enabled = TRUE")
        .bind(params.business_id)
        .bind(&params.addon_id)
        .fetch_optional(exec)
        .await
}

/// `ListAllAddons`: `many`.
pub async fn list_all_addons<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, name, description, icon, price, is_free, is_default_enabled, sort_order\nFROM addons\nORDER BY sort_order ASC\nLIMIT 100")
        .execute(exec)
        .await
}

/// Params for `ListBusinessAddons` (Go `ListBusinessAddonsParams`).
pub struct ListBusinessAddonsParams {
    pub business_id: i64,
}

/// `ListBusinessAddons`: `many`.
pub async fn list_business_addons<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListBusinessAddonsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT ba.addon_id, ba.is_enabled, ba.enabled_at, ba.disabled_at,\n       a.name, a.description, a.icon, a.price, a.is_free\nFROM business_addons ba\nJOIN addons a ON a.id = ba.addon_id\nWHERE ba.business_id = ?\nLIMIT 100")
        .bind(params.business_id)
        .execute(exec)
        .await
}
