//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `DeleteCompetingCustomDomainClaims` (Go `DeleteCompetingCustomDomainClaimsParams`).
pub struct DeleteCompetingCustomDomainClaimsParams {
    pub domain: String,
    pub business_id: i64,
}

/// `DeleteCompetingCustomDomainClaims`: `exec`.
pub async fn delete_competing_custom_domain_claims<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DeleteCompetingCustomDomainClaimsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query(
        "DELETE FROM custom_domains WHERE domain = ? AND business_id <> ? AND status <> 'verified'",
    )
    .bind(&params.domain)
    .bind(params.business_id)
    .execute(exec)
    .await
}

/// Params for `GetCustomDomain` (Go `GetCustomDomainParams`).
pub struct GetCustomDomainParams {
    pub business_id: i64,
}

/// `GetCustomDomain`: `one`.
pub async fn get_custom_domain<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetCustomDomainParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, domain, verification_token, status, ssl_status, verified_at, created_at, updated_at\nFROM custom_domains WHERE business_id = ? LIMIT 1")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetVerifiedCustomDomainByDomain` (Go `GetVerifiedCustomDomainByDomainParams`).
pub struct GetVerifiedCustomDomainByDomainParams {
    pub domain: String,
}

/// `GetVerifiedCustomDomainByDomain`: `one`.
pub async fn get_verified_custom_domain_by_domain<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetVerifiedCustomDomainByDomainParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, domain, verification_token, status, ssl_status, verified_at, created_at, updated_at\nFROM custom_domains WHERE domain = ? AND status = 'verified' LIMIT 1")
        .bind(&params.domain)
        .fetch_optional(exec)
        .await
}

/// Params for `RegisterCustomDomain` (Go `RegisterCustomDomainParams`).
pub struct RegisterCustomDomainParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub domain: String,
    pub verification_token: String,
}

/// `RegisterCustomDomain`: `execresult`.
pub async fn register_custom_domain<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &RegisterCustomDomainParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO custom_domains (owner_id, business_id, domain, verification_token)\nVALUES (?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(&params.domain)
        .bind(&params.verification_token)
        .execute(exec)
        .await
}

/// Params for `RemoveCustomDomain` (Go `RemoveCustomDomainParams`).
pub struct RemoveCustomDomainParams {
    pub business_id: i64,
}

/// `RemoveCustomDomain`: `exec`.
pub async fn remove_custom_domain<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &RemoveCustomDomainParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("DELETE FROM custom_domains WHERE business_id = ?")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `ReplaceCustomDomain` (Go `ReplaceCustomDomainParams`).
pub struct ReplaceCustomDomainParams {
    pub owner_id: i64,
    pub domain: String,
    pub verification_token: String,
    pub business_id: i64,
}

/// `ReplaceCustomDomain`: `exec`.
pub async fn replace_custom_domain<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ReplaceCustomDomainParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE custom_domains SET owner_id = ?, domain = ?, verification_token = ?, status = 'pending_verification', ssl_status = 'pending', verified_at = NULL\nWHERE business_id = ?")
        .bind(params.owner_id)
        .bind(&params.domain)
        .bind(&params.verification_token)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `RotateCustomDomainChallenge` (Go `RotateCustomDomainChallengeParams`).
pub struct RotateCustomDomainChallengeParams {
    pub verification_token: String,
    pub business_id: i64,
    pub domain: String,
}

/// `RotateCustomDomainChallenge`: `exec`.
pub async fn rotate_custom_domain_challenge<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &RotateCustomDomainChallengeParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE custom_domains SET verification_token = ?, status = 'pending_verification', ssl_status = 'pending', verified_at = NULL\nWHERE business_id = ? AND domain = ?")
        .bind(&params.verification_token)
        .bind(params.business_id)
        .bind(&params.domain)
        .execute(exec)
        .await
}

/// Params for `VerifyCustomDomain` (Go `VerifyCustomDomainParams`).
pub struct VerifyCustomDomainParams {
    pub business_id: i64,
    pub verification_token: String,
}

/// `VerifyCustomDomain`: `execresult`.
pub async fn verify_custom_domain<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &VerifyCustomDomainParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE custom_domains SET status = 'verified', ssl_status = 'pending', verified_at = CURRENT_TIMESTAMP\nWHERE business_id = ? AND verification_token = ?")
        .bind(params.business_id)
        .bind(&params.verification_token)
        .execute(exec)
        .await
}
