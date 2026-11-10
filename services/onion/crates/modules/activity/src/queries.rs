//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `CountActivity` (Go `CountActivityParams`).
pub struct CountActivityParams {
    pub business_id: i64,
    pub column2: String,
    pub app: String,
}

/// `CountActivity`: `one`.
pub async fn count_activity<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountActivityParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) AS total\nFROM activity_log a\nWHERE a.business_id = ?\n  AND (? = '' OR a.app = ?)")
        .bind(params.business_id)
        .bind(&params.column2)
        .bind(&params.app)
        .fetch_optional(exec)
        .await
}

/// Params for `ListActivity` (Go `ListActivityParams`).
pub struct ListActivityParams {
    pub business_id: i64,
    pub column2: String,
    pub app: String,
    pub limit: i32,
    pub offset: i32,
}

/// `ListActivity`: `many`.
pub async fn list_activity<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListActivityParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT a.id, a.owner_id, a.business_id, a.app, a.entity_type, a.entity_id, a.action, a.description, a.user_id, a.created_at\nFROM activity_log a\nWHERE a.business_id = ?\n  AND (? = '' OR a.app = ?)\nORDER BY a.created_at DESC\nLIMIT ? OFFSET ?")
        .bind(params.business_id)
        .bind(&params.column2)
        .bind(&params.app)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `RecordMutation` (Go `RecordMutationParams`).
pub struct RecordMutationParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub app: String,
    pub entity_type: String,
    pub entity_id: Option<i64>,
    pub action: String,
    pub description: Option<String>,
    pub user_id: Option<i64>,
    pub old_values: String,
    pub new_values: String,
    pub ip_address: Option<String>,
    pub created_at: Option<chrono::DateTime<chrono::Utc>>,
    pub source_event_id: Option<i64>,
}

/// `RecordMutation`: `exec`.
pub async fn record_mutation<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &RecordMutationParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO activity_log (owner_id, business_id, app, entity_type, entity_id, action, description, user_id, old_values, new_values, ip_address, created_at, source_event_id)\nVALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)\nON DUPLICATE KEY UPDATE id = id")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(&params.app)
        .bind(&params.entity_type)
        .bind(params.entity_id)
        .bind(&params.action)
        .bind(&params.description)
        .bind(params.user_id)
        .bind(&params.old_values)
        .bind(&params.new_values)
        .bind(&params.ip_address)
        .bind(params.created_at)
        .bind(params.source_event_id)
        .execute(exec)
        .await
}
