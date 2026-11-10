//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `InsertFact` (Go `InsertFactParams`).
pub struct InsertFactParams {
    pub business_id: i64,
    pub event_type: String,
    pub subject_type: String,
    pub subject_id: i64,
    pub payload: String,
    pub dedupe_key: String,
    pub occurred_at: chrono::NaiveDate,
}

/// `InsertFact`: `exec`.
pub async fn insert_fact<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &InsertFactParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO pos_outbox (business_id, event_type, subject_type, subject_id, payload, dedupe_key, occurred_at)\nVALUES (?, ?, ?, ?, ?, ?, ?)\nON DUPLICATE KEY UPDATE id = id")
        .bind(params.business_id)
        .bind(&params.event_type)
        .bind(&params.subject_type)
        .bind(params.subject_id)
        .bind(&params.payload)
        .bind(&params.dedupe_key)
        .bind(params.occurred_at)
        .execute(exec)
        .await
}

/// Params for `MarkRelayed` (Go `MarkRelayedParams`).
pub struct MarkRelayedParams {
    pub id: u64,
}

/// `MarkRelayed`: `exec`.
pub async fn mark_relayed<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &MarkRelayedParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query(
        "UPDATE pos_outbox SET relayed_at = CURRENT_TIMESTAMP WHERE id = ? AND relayed_at IS NULL",
    )
    .bind(params.id)
    .execute(exec)
    .await
}

/// Params for `PendingFacts` (Go `PendingFactsParams`).
pub struct PendingFactsParams {
    pub limit: i32,
}

/// `PendingFacts`: `many`.
pub async fn pending_facts<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &PendingFactsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, business_id, event_type, subject_type, subject_id, payload, dedupe_key, occurred_at\nFROM pos_outbox WHERE relayed_at IS NULL ORDER BY id LIMIT ?")
        .bind(params.limit)
        .execute(exec)
        .await
}
