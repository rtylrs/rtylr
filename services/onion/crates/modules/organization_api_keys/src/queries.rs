//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `APIKeyExists` (Go `APIKeyExistsParams`).
pub struct APIKeyExistsParams {
    pub id: i64,
    pub business_id: i64,
}

/// `APIKeyExists`: `one`.
pub async fn api_key_exists<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &APIKeyExistsParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) FROM hub_api_keys WHERE id = ? AND business_id = ?")
        .bind(params.id)
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `CreateAPIKey` (Go `CreateAPIKeyParams`).
pub struct CreateAPIKeyParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub name: String,
    pub key_hash: String,
    pub key_prefix: String,
    pub scopes: String,
}

/// `CreateAPIKey`: `execresult`.
pub async fn create_api_key<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateAPIKeyParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO hub_api_keys (owner_id, business_id, name, key_hash, key_prefix, scopes)\nVALUES (?, ?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(&params.name)
        .bind(&params.key_hash)
        .bind(&params.key_prefix)
        .bind(&params.scopes)
        .execute(exec)
        .await
}

/// Params for `ListAPIKeys` (Go `ListAPIKeysParams`).
pub struct ListAPIKeysParams {
    pub business_id: i64,
}

/// `ListAPIKeys`: `many`.
pub async fn list_api_keys<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListAPIKeysParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, name, key_prefix, scopes, last_used_at, expires_at, is_active, created_at\nFROM hub_api_keys\nWHERE business_id = ? AND is_active = TRUE\nORDER BY created_at DESC")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `RevokeAPIKey` (Go `RevokeAPIKeyParams`).
pub struct RevokeAPIKeyParams {
    pub id: i64,
    pub business_id: i64,
}

/// `RevokeAPIKey`: `exec`.
pub async fn revoke_api_key<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &RevokeAPIKeyParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE hub_api_keys\nSET is_active = FALSE\nWHERE id = ? AND business_id = ?")
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}
