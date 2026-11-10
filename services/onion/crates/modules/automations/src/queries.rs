//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `ClaimAutomationRun` (Go `ClaimAutomationRunParams`).
pub struct ClaimAutomationRunParams {
    pub id: u64,
}

/// `ClaimAutomationRun`: `execresult`.
pub async fn claim_automation_run<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ClaimAutomationRunParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE automation_runs\nSET status = 'running', attempts = attempts + 1, updated_at = NOW()\nWHERE id = ? AND status IN ('pending', 'failed')")
        .bind(params.id)
        .execute(exec)
        .await
}

/// Params for `CreateAutomation` (Go `CreateAutomationParams`).
pub struct CreateAutomationParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub name: String,
    pub trigger_type: String,
    pub trigger_config: String,
    pub actions: String,
    pub is_active: Option<bool>,
}

/// `CreateAutomation`: `execresult`.
pub async fn create_automation<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateAutomationParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO crm_automations\n  (owner_id, business_id, name, trigger_type, trigger_config, actions, is_active)\nVALUES (?, ?, ?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(&params.name)
        .bind(&params.trigger_type)
        .bind(&params.trigger_config)
        .bind(&params.actions)
        .bind(params.is_active)
        .execute(exec)
        .await
}

/// Params for `DeleteAutomation` (Go `DeleteAutomationParams`).
pub struct DeleteAutomationParams {
    pub id: i64,
    pub business_id: i64,
}

/// `DeleteAutomation`: `execresult`.
pub async fn delete_automation<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DeleteAutomationParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("DELETE FROM crm_automations WHERE id = ? AND business_id = ?")
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `EnqueueAutomationRun` (Go `EnqueueAutomationRunParams`).
pub struct EnqueueAutomationRunParams {
    pub business_id: i64,
    pub automation_id: i64,
    pub event_id: u64,
}

/// `EnqueueAutomationRun`: `exec`.
pub async fn enqueue_automation_run<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &EnqueueAutomationRunParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO automation_runs (business_id, automation_id, event_id)\nVALUES (?, ?, ?)\nON DUPLICATE KEY UPDATE id = id")
        .bind(params.business_id)
        .bind(params.automation_id)
        .bind(params.event_id)
        .execute(exec)
        .await
}

/// Params for `GetAutomationForRun` (Go `GetAutomationForRunParams`).
pub struct GetAutomationForRunParams {
    pub id: i64,
    pub business_id: i64,
}

/// `GetAutomationForRun`: `one`.
pub async fn get_automation_for_run<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetAutomationForRunParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, trigger_type, actions, is_active\nFROM crm_automations\nWHERE id = ? AND business_id = ?")
        .bind(params.id)
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `ListAutomations` (Go `ListAutomationsParams`).
pub struct ListAutomationsParams {
    pub business_id: i64,
}

/// `ListAutomations`: `many`.
pub async fn list_automations<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListAutomationsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, name, trigger_type, trigger_config, actions, is_active,\n       last_run_at, run_count, created_at, updated_at\nFROM crm_automations WHERE business_id = ? ORDER BY created_at DESC LIMIT 100")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `ListAutomationsForEvent` (Go `ListAutomationsForEventParams`).
pub struct ListAutomationsForEventParams {
    pub business_id: i64,
    pub trigger_type: String,
}

/// `ListAutomationsForEvent`: `many`.
pub async fn list_automations_for_event<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListAutomationsForEventParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id\nFROM crm_automations\nWHERE business_id = ? AND trigger_type = ? AND is_active = TRUE")
        .bind(params.business_id)
        .bind(&params.trigger_type)
        .execute(exec)
        .await
}

/// Params for `ListDueAutomationRuns` (Go `ListDueAutomationRunsParams`).
pub struct ListDueAutomationRunsParams {
    pub limit: i32,
}

/// `ListDueAutomationRuns`: `many`.
pub async fn list_due_automation_runs<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListDueAutomationRunsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, business_id, automation_id, event_id, attempts, max_attempts\nFROM automation_runs\nWHERE status IN ('pending', 'failed')\n  AND next_attempt_at <= NOW()\nORDER BY next_attempt_at ASC, id ASC\nLIMIT ?")
        .bind(params.limit)
        .execute(exec)
        .await
}

/// Params for `MarkAutomationRunDead` (Go `MarkAutomationRunDeadParams`).
pub struct MarkAutomationRunDeadParams {
    pub last_error: Option<String>,
    pub id: u64,
}

/// `MarkAutomationRunDead`: `exec`.
pub async fn mark_automation_run_dead<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &MarkAutomationRunDeadParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE automation_runs\nSET status = 'dead', last_error = ?, updated_at = NOW()\nWHERE id = ?")
        .bind(&params.last_error)
        .bind(params.id)
        .execute(exec)
        .await
}

/// Params for `MarkAutomationRunFailed` (Go `MarkAutomationRunFailedParams`).
pub struct MarkAutomationRunFailedParams {
    pub retry_after_seconds: String,
    pub last_error: Option<String>,
    pub id: u64,
}

/// `MarkAutomationRunFailed`: `exec`.
pub async fn mark_automation_run_failed<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &MarkAutomationRunFailedParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE automation_runs\nSET status = IF(attempts >= max_attempts, 'dead', 'failed'),\n    next_attempt_at = DATE_ADD(NOW(), INTERVAL ? SECOND),\n    last_error = ?,\n    updated_at = NOW()\nWHERE id = ?")
        .bind(&params.retry_after_seconds)
        .bind(&params.last_error)
        .bind(params.id)
        .execute(exec)
        .await
}

/// Params for `MarkAutomationRunSucceeded` (Go `MarkAutomationRunSucceededParams`).
pub struct MarkAutomationRunSucceededParams {
    pub id: u64,
}

/// `MarkAutomationRunSucceeded`: `exec`.
pub async fn mark_automation_run_succeeded<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &MarkAutomationRunSucceededParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE automation_runs\nSET status = 'succeeded', completed_at = NOW(), last_error = NULL, updated_at = NOW()\nWHERE id = ?")
        .bind(params.id)
        .execute(exec)
        .await
}

/// Params for `RecordAutomationExecution` (Go `RecordAutomationExecutionParams`).
pub struct RecordAutomationExecutionParams {
    pub id: i64,
    pub business_id: i64,
}

/// `RecordAutomationExecution`: `exec`.
pub async fn record_automation_execution<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &RecordAutomationExecutionParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE crm_automations\nSET run_count = run_count + 1, last_run_at = NOW()\nWHERE id = ? AND business_id = ?")
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `RecoverAbandonedAutomationRuns` (Go `RecoverAbandonedAutomationRunsParams`).
pub struct RecoverAbandonedAutomationRunsParams {
    pub stale_after_seconds: String,
}

/// `RecoverAbandonedAutomationRuns`: `execresult`.
pub async fn recover_abandoned_automation_runs<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &RecoverAbandonedAutomationRunsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE automation_runs\nSET status = IF(attempts >= max_attempts, 'dead', 'failed'),\n    next_attempt_at = NOW(),\n    last_error = 'run was abandoned while executing and has been requeued',\n    updated_at = NOW()\nWHERE status = 'running'\n  AND updated_at < DATE_SUB(NOW(), INTERVAL ? SECOND)")
        .bind(&params.stale_after_seconds)
        .execute(exec)
        .await
}

/// Params for `UpdateAutomation` (Go `UpdateAutomationParams`).
pub struct UpdateAutomationParams {
    pub name: String,
    pub trigger_type: String,
    pub trigger_config: String,
    pub actions: String,
    pub is_active: Option<bool>,
    pub id: i64,
    pub business_id: i64,
}

/// `UpdateAutomation`: `execresult`.
pub async fn update_automation<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateAutomationParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE crm_automations\nSET name = ?, trigger_type = ?, trigger_config = ?, actions = ?, is_active = ?\nWHERE id = ? AND business_id = ?")
        .bind(&params.name)
        .bind(&params.trigger_type)
        .bind(&params.trigger_config)
        .bind(&params.actions)
        .bind(params.is_active)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}
