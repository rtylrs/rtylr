//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `ConnectIntegration` (Go `ConnectIntegrationParams`).
pub struct ConnectIntegrationParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub provider: String,
    pub config: String,
}

/// `ConnectIntegration`: `execresult`.
pub async fn connect_integration<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ConnectIntegrationParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO hub_integrations (owner_id, business_id, provider, status, config, connected_at)\nVALUES (?, ?, ?, 'connected', ?, NOW())")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(&params.provider)
        .bind(&params.config)
        .execute(exec)
        .await
}

/// Params for `DisconnectIntegration` (Go `DisconnectIntegrationParams`).
pub struct DisconnectIntegrationParams {
    pub id: i64,
    pub business_id: i64,
}

/// `DisconnectIntegration`: `execresult`.
pub async fn disconnect_integration<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DisconnectIntegrationParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("DELETE FROM hub_integrations WHERE id = ? AND business_id = ?")
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `FindIntegrationByProvider` (Go `FindIntegrationByProviderParams`).
pub struct FindIntegrationByProviderParams {
    pub business_id: i64,
    pub provider: String,
}

/// `FindIntegrationByProvider`: `one`.
pub async fn find_integration_by_provider<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &FindIntegrationByProviderParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id FROM hub_integrations WHERE business_id = ? AND provider = ? ORDER BY id LIMIT 1")
        .bind(params.business_id)
        .bind(&params.provider)
        .fetch_optional(exec)
        .await
}

/// Params for `ListIntegrations` (Go `ListIntegrationsParams`).
pub struct ListIntegrationsParams {
    pub business_id: i64,
}

/// `ListIntegrations`: `many`.
pub async fn list_integrations<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListIntegrationsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, provider, status, config, connected_at, created_at\nFROM hub_integrations\nWHERE business_id = ?\nORDER BY created_at DESC\nLIMIT 200")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `ReconnectIntegration` (Go `ReconnectIntegrationParams`).
pub struct ReconnectIntegrationParams {
    pub config: String,
    pub id: i64,
    pub business_id: i64,
}

/// `ReconnectIntegration`: `exec`.
pub async fn reconnect_integration<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ReconnectIntegrationParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE hub_integrations SET status = 'connected', config = ?, connected_at = NOW()\nWHERE id = ? AND business_id = ?")
        .bind(&params.config)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}
