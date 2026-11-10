//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `GetHRConfig` (Go `GetHRConfigParams`).
pub struct GetHRConfigParams {
    pub business_id: i64,
}

/// `GetHRConfig`: `one`.
pub async fn get_hr_config<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetHRConfigParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, business_id, setup_completed_at, created_at, updated_at\nFROM hr_configs\nWHERE business_id = ?\nLIMIT 1")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `ListSetupSteps` (Go `ListSetupStepsParams`).
pub struct ListSetupStepsParams {
    pub business_id: i64,
}

/// `ListSetupSteps`: `many`.
pub async fn list_setup_steps<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListSetupStepsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, step_key, completed_at\nFROM hr_setup_steps\nWHERE business_id = ?\nORDER BY id ASC\nLIMIT 500")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `UpsertHRConfig` (Go `UpsertHRConfigParams`).
pub struct UpsertHRConfigParams {
    pub business_id: i64,
    pub setup_completed_at: Option<chrono::DateTime<chrono::Utc>>,
}

/// `UpsertHRConfig`: `exec`.
pub async fn upsert_hr_config<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpsertHRConfigParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO hr_configs (owner_id, business_id, setup_completed_at)\nVALUES (?, ?, ?)\nON DUPLICATE KEY UPDATE setup_completed_at = VALUES(setup_completed_at), updated_at = NOW()")
        .bind(params.business_id)
        .bind(params.business_id)
        .bind(params.setup_completed_at)
        .execute(exec)
        .await
}

/// Params for `UpsertSetupStep` (Go `UpsertSetupStepParams`).
pub struct UpsertSetupStepParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub step_key: String,
    pub completed_at: Option<chrono::DateTime<chrono::Utc>>,
}

/// `UpsertSetupStep`: `exec`.
pub async fn upsert_setup_step<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpsertSetupStepParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO hr_setup_steps (owner_id, business_id, step_key, completed_at)\nVALUES (?, ?, ?, ?)\nON DUPLICATE KEY UPDATE completed_at = VALUES(completed_at)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(&params.step_key)
        .bind(params.completed_at)
        .execute(exec)
        .await
}
