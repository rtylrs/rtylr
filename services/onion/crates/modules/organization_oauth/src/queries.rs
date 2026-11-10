//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `ConsumeAuthCode` (Go `ConsumeAuthCodeParams`).
pub struct ConsumeAuthCodeParams {
    pub code: String,
}

/// `ConsumeAuthCode`: `execresult`.
pub async fn consume_auth_code<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ConsumeAuthCodeParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE oauth_auth_codes SET used = 1 WHERE code = ? AND used = 0")
        .bind(&params.code)
        .execute(exec)
        .await
}

/// Params for `ConsumeRefreshToken` (Go `ConsumeRefreshTokenParams`).
pub struct ConsumeRefreshTokenParams {
    pub refresh_token: String,
}

/// `ConsumeRefreshToken`: `execresult`.
pub async fn consume_refresh_token<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ConsumeRefreshTokenParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE oauth_tokens SET revoked = 1 WHERE refresh_token = ? AND revoked = 0")
        .bind(&params.refresh_token)
        .execute(exec)
        .await
}

/// Params for `GetAuthCode` (Go `GetAuthCodeParams`).
pub struct GetAuthCodeParams {
    pub code: String,
}

/// `GetAuthCode`: `one`.
pub async fn get_auth_code<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetAuthCodeParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT code, client_id, user_id, redirect_uri, scope, code_challenge, code_challenge_method, expires_at, used\nFROM oauth_auth_codes WHERE code = ?")
        .bind(&params.code)
        .fetch_optional(exec)
        .await
}

/// Params for `GetOAuthClient` (Go `GetOAuthClientParams`).
pub struct GetOAuthClientParams {
    pub id: String,
}

/// `GetOAuthClient`: `one`.
pub async fn get_o_auth_client<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetOAuthClientParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, name, redirect_uris, scopes, is_public, created_at\nFROM oauth_clients WHERE id = ?")
        .bind(&params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetTokenByRefresh` (Go `GetTokenByRefreshParams`).
pub struct GetTokenByRefreshParams {
    pub refresh_token: String,
}

/// `GetTokenByRefresh`: `one`.
pub async fn get_token_by_refresh<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetTokenByRefreshParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, access_token, refresh_token, client_id, user_id, scope, expires_at, revoked\nFROM oauth_tokens WHERE refresh_token = ?")
        .bind(&params.refresh_token)
        .fetch_optional(exec)
        .await
}

/// Params for `GetUserBusinessID` (Go `GetUserBusinessIDParams`).
pub struct GetUserBusinessIDParams {
    pub id: i64,
}

/// `GetUserBusinessID`: `one`.
pub async fn get_user_business_id<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetUserBusinessIDParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT business_id FROM users WHERE id = ?")
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetUserIDByEmail` (Go `GetUserIDByEmailParams`).
pub struct GetUserIDByEmailParams {
    pub email: String,
}

/// `GetUserIDByEmail`: `one`.
pub async fn get_user_id_by_email<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetUserIDByEmailParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id FROM users WHERE email = ? ORDER BY id ASC LIMIT 1")
        .bind(&params.email)
        .fetch_optional(exec)
        .await
}

/// Params for `GetUserPasswordHash` (Go `GetUserPasswordHashParams`).
pub struct GetUserPasswordHashParams {
    pub id: i64,
}

/// `GetUserPasswordHash`: `one`.
pub async fn get_user_password_hash<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetUserPasswordHashParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT password_hash FROM users WHERE id = ?")
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetUserTokenEpoch` (Go `GetUserTokenEpochParams`).
pub struct GetUserTokenEpochParams {
    pub id: i64,
}

/// `GetUserTokenEpoch`: `one`.
pub async fn get_user_token_epoch<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetUserTokenEpochParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT token_epoch FROM users WHERE id = ?")
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `MarkAuthCodeUsed` (Go `MarkAuthCodeUsedParams`).
pub struct MarkAuthCodeUsedParams {
    pub code: String,
}

/// `MarkAuthCodeUsed`: `exec`.
pub async fn mark_auth_code_used<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &MarkAuthCodeUsedParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE oauth_auth_codes SET used = 1 WHERE code = ?")
        .bind(&params.code)
        .execute(exec)
        .await
}

/// Params for `RevokeToken` (Go `RevokeTokenParams`).
pub struct RevokeTokenParams {
    pub refresh_token: String,
}

/// `RevokeToken`: `exec`.
pub async fn revoke_token<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &RevokeTokenParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE oauth_tokens SET revoked = 1 WHERE refresh_token = ?")
        .bind(&params.refresh_token)
        .execute(exec)
        .await
}

/// Params for `SaveAuthCode` (Go `SaveAuthCodeParams`).
pub struct SaveAuthCodeParams {
    pub code: String,
    pub client_id: String,
    pub user_id: i64,
    pub redirect_uri: String,
    pub scope: String,
    pub code_challenge: String,
    pub code_challenge_method: String,
    pub expires_at: chrono::NaiveDate,
}

/// `SaveAuthCode`: `exec`.
pub async fn save_auth_code<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &SaveAuthCodeParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO oauth_auth_codes\n(code, client_id, user_id, redirect_uri, scope, code_challenge, code_challenge_method, expires_at)\nVALUES (?, ?, ?, ?, ?, ?, ?, ?)")
        .bind(&params.code)
        .bind(&params.client_id)
        .bind(params.user_id)
        .bind(&params.redirect_uri)
        .bind(&params.scope)
        .bind(&params.code_challenge)
        .bind(&params.code_challenge_method)
        .bind(params.expires_at)
        .execute(exec)
        .await
}

/// Params for `SaveToken` (Go `SaveTokenParams`).
pub struct SaveTokenParams {
    pub access_token: String,
    pub refresh_token: String,
    pub client_id: String,
    pub user_id: i64,
    pub scope: String,
    pub expires_at: chrono::NaiveDate,
}

/// `SaveToken`: `exec`.
pub async fn save_token<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &SaveTokenParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO oauth_tokens (access_token, refresh_token, client_id, user_id, scope, expires_at)\nVALUES (?, ?, ?, ?, ?, ?)")
        .bind(&params.access_token)
        .bind(&params.refresh_token)
        .bind(&params.client_id)
        .bind(params.user_id)
        .bind(&params.scope)
        .bind(params.expires_at)
        .execute(exec)
        .await
}
