//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `AdvanceRoutine` (Go `AdvanceRoutineParams`).
pub struct AdvanceRoutineParams {
    pub id: i64,
}

/// `AdvanceRoutine`: `exec`.
pub async fn advance_routine<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &AdvanceRoutineParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE caret_buddy_routines\nSET last_run_at = NOW(),\n    next_run_at = DATE_ADD(\n      next_run_at,\n      INTERVAL (FLOOR(TIMESTAMPDIFF(MINUTE, next_run_at, NOW()) / every_minutes) + 1) * every_minutes MINUTE\n    )\nWHERE id = ?")
        .bind(params.id)
        .execute(exec)
        .await
}

/// `ClaimNextRunForUpdate`: `one`.
pub async fn claim_next_run_for_update<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, buddy_id, routine_id, business_id, trigger_kind, status, attempts, lease_until, report, error, tool_calls, tokens, started_at, finished_at, created_at FROM caret_buddy_runs\nWHERE status = 'queued'\nORDER BY id\nLIMIT 1\nFOR UPDATE")
        .fetch_optional(exec)
        .await
}

/// Params for `CountQueuedEventRuns` (Go `CountQueuedEventRunsParams`).
pub struct CountQueuedEventRunsParams {
    pub buddy_id: i64,
}

/// `CountQueuedEventRuns`: `one`.
pub async fn count_queued_event_runs<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountQueuedEventRunsParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) FROM caret_buddy_runs\nWHERE buddy_id = ? AND trigger_kind = 'event' AND status = 'queued'")
        .bind(params.buddy_id)
        .fetch_optional(exec)
        .await
}

/// Params for `CountRunsSince` (Go `CountRunsSinceParams`).
pub struct CountRunsSinceParams {
    pub buddy_id: i64,
}

/// `CountRunsSince`: `one`.
pub async fn count_runs_since<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountRunsSinceParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) FROM caret_buddy_runs\nWHERE buddy_id = ? AND created_at >= DATE_SUB(NOW(), INTERVAL 24 HOUR)")
        .bind(params.buddy_id)
        .fetch_optional(exec)
        .await
}

/// Params for `DecideAction` (Go `DecideActionParams`).
pub struct DecideActionParams {
    pub status: String,
    pub decided_by: Option<i64>,
    pub id: i64,
    pub business_id: i64,
}

/// `DecideAction`: `execrows`.
pub async fn decide_action<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DecideActionParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE caret_buddy_actions\nSET status = ?, decided_by = ?, decided_at = NOW()\nWHERE id = ? AND business_id = ? AND status = 'pending' AND expires_at > NOW()")
        .bind(&params.status)
        .bind(params.decided_by)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `DeleteRoutine` (Go `DeleteRoutineParams`).
pub struct DeleteRoutineParams {
    pub id: i64,
    pub business_id: i64,
    pub user_id: i64,
}

/// `DeleteRoutine`: `execrows`.
pub async fn delete_routine<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DeleteRoutineParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query(
        "DELETE FROM caret_buddy_routines\nWHERE id = ? AND business_id = ? AND user_id = ?",
    )
    .bind(params.id)
    .bind(params.business_id)
    .bind(params.user_id)
    .execute(exec)
    .await
}

/// Params for `EnsureBuddy` (Go `EnsureBuddyParams`).
pub struct EnsureBuddyParams {
    pub business_id: i64,
    pub run_as_user_id: i64,
}

/// `EnsureBuddy`: `exec`.
pub async fn ensure_buddy<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &EnsureBuddyParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT IGNORE INTO caret_buddies (business_id, run_as_user_id, enabled, instructions)\nVALUES (?, ?, 0, '')")
        .bind(params.business_id)
        .bind(params.run_as_user_id)
        .execute(exec)
        .await
}

/// `ExpireActions`: `execrows`.
pub async fn expire_actions<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE caret_buddy_actions\nSET status = 'expired'\nWHERE status = 'pending' AND expires_at < NOW()")
        .execute(exec)
        .await
}

/// Params for `FailExhaustedRuns` (Go `FailExhaustedRunsParams`).
pub struct FailExhaustedRunsParams {
    pub attempts: u8,
}

/// `FailExhaustedRuns`: `execrows`.
pub async fn fail_exhausted_runs<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &FailExhaustedRunsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE caret_buddy_runs\nSET status = 'failed', error = 'run was interrupted too many times', finished_at = NOW(), lease_until = NULL\nWHERE status = 'running' AND lease_until < NOW() AND attempts >= ?")
        .bind(params.attempts)
        .execute(exec)
        .await
}

/// Params for `FinishAction` (Go `FinishActionParams`).
pub struct FinishActionParams {
    pub status: String,
    pub result: String,
    pub id: i64,
    pub business_id: i64,
    pub decided_by: Option<i64>,
}

/// `FinishAction`: `execrows`.
pub async fn finish_action<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &FinishActionParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE caret_buddy_actions\nSET status = ?, result = ?\nWHERE id = ? AND business_id = ? AND status = 'approved' AND decided_by = ?")
        .bind(&params.status)
        .bind(&params.result)
        .bind(params.id)
        .bind(params.business_id)
        .bind(params.decided_by)
        .execute(exec)
        .await
}

/// Params for `FinishRun` (Go `FinishRunParams`).
pub struct FinishRunParams {
    pub status: String,
    pub report: Option<String>,
    pub error: Option<String>,
    pub tool_calls: i32,
    pub tokens: i64,
    pub id: i64,
}

/// `FinishRun`: `exec`.
pub async fn finish_run<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &FinishRunParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE caret_buddy_runs\nSET status = ?, report = ?, error = ?, tool_calls = ?, tokens = ?,\n    finished_at = NOW(), lease_until = NULL\nWHERE id = ?")
        .bind(&params.status)
        .bind(&params.report)
        .bind(&params.error)
        .bind(params.tool_calls)
        .bind(params.tokens)
        .bind(params.id)
        .execute(exec)
        .await
}

/// Params for `GetAction` (Go `GetActionParams`).
pub struct GetActionParams {
    pub id: i64,
    pub business_id: i64,
}

/// `GetAction`: `one`.
pub async fn get_action<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetActionParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, run_id, business_id, tool_name, input, fingerprint, status, decided_by, decided_at, result, expires_at, created_at FROM caret_buddy_actions WHERE id = ? AND business_id = ?")
        .bind(params.id)
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetBuddy` (Go `GetBuddyParams`).
pub struct GetBuddyParams {
    pub id: i64,
}

/// `GetBuddy`: `one`.
pub async fn get_buddy<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetBuddyParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, business_id, run_as_user_id, enabled, instructions, interval_minutes, daily_run_cap, next_run_at, last_run_at, created_at, updated_at FROM caret_buddies WHERE id = ?")
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetBuddyByBusiness` (Go `GetBuddyByBusinessParams`).
pub struct GetBuddyByBusinessParams {
    pub business_id: i64,
}

/// `GetBuddyByBusiness`: `one`.
pub async fn get_buddy_by_business<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetBuddyByBusinessParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, business_id, run_as_user_id, enabled, instructions, interval_minutes, daily_run_cap, next_run_at, last_run_at, created_at, updated_at FROM caret_buddies WHERE business_id = ?")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetBuddyByBusinessForUpdate` (Go `GetBuddyByBusinessForUpdateParams`).
pub struct GetBuddyByBusinessForUpdateParams {
    pub business_id: i64,
}

/// `GetBuddyByBusinessForUpdate`: `one`.
pub async fn get_buddy_by_business_for_update<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetBuddyByBusinessForUpdateParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, business_id, run_as_user_id, enabled, instructions, interval_minutes, daily_run_cap, next_run_at, last_run_at, created_at, updated_at FROM caret_buddies WHERE business_id = ? FOR UPDATE")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetRoutine` (Go `GetRoutineParams`).
pub struct GetRoutineParams {
    pub id: i64,
    pub business_id: i64,
    pub user_id: i64,
}

/// `GetRoutine`: `one`.
pub async fn get_routine<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetRoutineParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, buddy_id, business_id, user_id, name, instructions, every_minutes, enabled, next_run_at, last_run_at, created_at, updated_at FROM caret_buddy_routines\nWHERE id = ? AND business_id = ? AND user_id = ?")
        .bind(params.id)
        .bind(params.business_id)
        .bind(params.user_id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetRoutineByID` (Go `GetRoutineByIDParams`).
pub struct GetRoutineByIDParams {
    pub id: i64,
}

/// `GetRoutineByID`: `one`.
pub async fn get_routine_by_id<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetRoutineByIDParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, buddy_id, business_id, user_id, name, instructions, every_minutes, enabled, next_run_at, last_run_at, created_at, updated_at FROM caret_buddy_routines WHERE id = ?")
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `InsertAction` (Go `InsertActionParams`).
pub struct InsertActionParams {
    pub run_id: i64,
    pub business_id: i64,
    pub tool_name: String,
    pub input: String,
    pub fingerprint: String,
    pub ttl_hours: i64,
}

/// `InsertAction`: `execresult`.
pub async fn insert_action<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &InsertActionParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT IGNORE INTO caret_buddy_actions (run_id, business_id, tool_name, input, fingerprint, expires_at)\nVALUES (?, ?, ?, ?, ?, DATE_ADD(NOW(), INTERVAL CAST(? AS SIGNED) HOUR))")
        .bind(params.run_id)
        .bind(params.business_id)
        .bind(&params.tool_name)
        .bind(&params.input)
        .bind(&params.fingerprint)
        .bind(params.ttl_hours)
        .execute(exec)
        .await
}

/// Params for `InsertRoutine` (Go `InsertRoutineParams`).
pub struct InsertRoutineParams {
    pub buddy_id: i64,
    pub business_id: i64,
    pub user_id: i64,
    pub name: String,
    pub instructions: String,
    pub every_minutes: i32,
    pub next_run_at: Option<chrono::DateTime<chrono::Utc>>,
}

/// `InsertRoutine`: `execresult`.
pub async fn insert_routine<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &InsertRoutineParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO caret_buddy_routines (buddy_id, business_id, user_id, name, instructions, every_minutes, enabled, next_run_at)\nVALUES (?, ?, ?, ?, ?, ?, 1, ?)")
        .bind(params.buddy_id)
        .bind(params.business_id)
        .bind(params.user_id)
        .bind(&params.name)
        .bind(&params.instructions)
        .bind(params.every_minutes)
        .bind(params.next_run_at)
        .execute(exec)
        .await
}

/// Params for `InsertRoutineRun` (Go `InsertRoutineRunParams`).
pub struct InsertRoutineRunParams {
    pub buddy_id: i64,
    pub business_id: i64,
    pub routine_id: Option<i64>,
}

/// `InsertRoutineRun`: `execresult`.
pub async fn insert_routine_run<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &InsertRoutineRunParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO caret_buddy_runs (buddy_id, business_id, routine_id, trigger_kind) VALUES (?, ?, ?, 'routine')")
        .bind(params.buddy_id)
        .bind(params.business_id)
        .bind(params.routine_id)
        .execute(exec)
        .await
}

/// Params for `InsertRun` (Go `InsertRunParams`).
pub struct InsertRunParams {
    pub buddy_id: i64,
    pub business_id: i64,
    pub trigger_kind: String,
}

/// `InsertRun`: `execresult`.
pub async fn insert_run<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &InsertRunParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query(
        "INSERT INTO caret_buddy_runs (buddy_id, business_id, trigger_kind) VALUES (?, ?, ?)",
    )
    .bind(params.buddy_id)
    .bind(params.business_id)
    .bind(&params.trigger_kind)
    .execute(exec)
    .await
}

/// Params for `ListActionsByBusiness` (Go `ListActionsByBusinessParams`).
pub struct ListActionsByBusinessParams {
    pub business_id: i64,
    pub status: String,
    pub limit: i32,
}

/// `ListActionsByBusiness`: `many`.
pub async fn list_actions_by_business<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListActionsByBusinessParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, run_id, business_id, tool_name, input, fingerprint, status, decided_by, decided_at, result, expires_at, created_at FROM caret_buddy_actions\nWHERE business_id = ? AND status = ?\nORDER BY id DESC\nLIMIT ?")
        .bind(params.business_id)
        .bind(&params.status)
        .bind(params.limit)
        .execute(exec)
        .await
}

/// Params for `ListAllActionsByBusiness` (Go `ListAllActionsByBusinessParams`).
pub struct ListAllActionsByBusinessParams {
    pub business_id: i64,
    pub limit: i32,
}

/// `ListAllActionsByBusiness`: `many`.
pub async fn list_all_actions_by_business<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListAllActionsByBusinessParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, run_id, business_id, tool_name, input, fingerprint, status, decided_by, decided_at, result, expires_at, created_at FROM caret_buddy_actions\nWHERE business_id = ?\nORDER BY id DESC\nLIMIT ?")
        .bind(params.business_id)
        .bind(params.limit)
        .execute(exec)
        .await
}

/// Params for `ListDueRoutinesForUpdate` (Go `ListDueRoutinesForUpdateParams`).
pub struct ListDueRoutinesForUpdateParams {
    pub limit: i32,
}

/// `ListDueRoutinesForUpdate`: `many`.
pub async fn list_due_routines_for_update<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListDueRoutinesForUpdateParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, buddy_id, business_id, user_id, name, instructions, every_minutes, enabled, next_run_at, last_run_at, created_at, updated_at FROM caret_buddy_routines\nWHERE enabled = 1 AND next_run_at IS NOT NULL AND next_run_at <= NOW()\nORDER BY next_run_at\nLIMIT ?\nFOR UPDATE")
        .bind(params.limit)
        .execute(exec)
        .await
}

/// Params for `ListRoutinesByOwner` (Go `ListRoutinesByOwnerParams`).
pub struct ListRoutinesByOwnerParams {
    pub business_id: i64,
    pub user_id: i64,
}

/// `ListRoutinesByOwner`: `many`.
pub async fn list_routines_by_owner<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListRoutinesByOwnerParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, buddy_id, business_id, user_id, name, instructions, every_minutes, enabled, next_run_at, last_run_at, created_at, updated_at FROM caret_buddy_routines\nWHERE business_id = ? AND user_id = ?\nORDER BY id")
        .bind(params.business_id)
        .bind(params.user_id)
        .execute(exec)
        .await
}

/// Params for `ListRunsByBusiness` (Go `ListRunsByBusinessParams`).
pub struct ListRunsByBusinessParams {
    pub business_id: i64,
    pub limit: i32,
}

/// `ListRunsByBusiness`: `many`.
pub async fn list_runs_by_business<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListRunsByBusinessParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, buddy_id, routine_id, business_id, trigger_kind, status, attempts, lease_until, report, error, tool_calls, tokens, started_at, finished_at, created_at FROM caret_buddy_runs\nWHERE business_id = ?\nORDER BY id DESC\nLIMIT ?")
        .bind(params.business_id)
        .bind(params.limit)
        .execute(exec)
        .await
}

/// Params for `MarkRunRunning` (Go `MarkRunRunningParams`).
pub struct MarkRunRunningParams {
    pub lease_seconds: i64,
    pub id: i64,
}

/// `MarkRunRunning`: `exec`.
pub async fn mark_run_running<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &MarkRunRunningParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE caret_buddy_runs\nSET status = 'running', attempts = attempts + 1, started_at = NOW(),\n    lease_until = DATE_ADD(NOW(), INTERVAL CAST(? AS SIGNED) SECOND)\nWHERE id = ?")
        .bind(params.lease_seconds)
        .bind(params.id)
        .execute(exec)
        .await
}

/// Params for `RecordBuddyEvent` (Go `RecordBuddyEventParams`).
pub struct RecordBuddyEventParams {
    pub event_id: u64,
    pub buddy_id: i64,
}

/// `RecordBuddyEvent`: `execresult`.
pub async fn record_buddy_event<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &RecordBuddyEventParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT IGNORE INTO caret_buddy_events (event_id, buddy_id) VALUES (?, ?)")
        .bind(params.event_id)
        .bind(params.buddy_id)
        .execute(exec)
        .await
}

/// Params for `RequeueExpiredRuns` (Go `RequeueExpiredRunsParams`).
pub struct RequeueExpiredRunsParams {
    pub attempts: u8,
}

/// `RequeueExpiredRuns`: `execrows`.
pub async fn requeue_expired_runs<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &RequeueExpiredRunsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE caret_buddy_runs\nSET status = 'queued', lease_until = NULL\nWHERE status = 'running' AND lease_until < NOW() AND attempts < ?")
        .bind(params.attempts)
        .execute(exec)
        .await
}

/// Params for `UpdateRoutine` (Go `UpdateRoutineParams`).
pub struct UpdateRoutineParams {
    pub name: String,
    pub instructions: String,
    pub every_minutes: i32,
    pub enabled: bool,
    pub next_run_at: Option<chrono::DateTime<chrono::Utc>>,
    pub id: i64,
    pub business_id: i64,
    pub user_id: i64,
}

/// `UpdateRoutine`: `execrows`.
pub async fn update_routine<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateRoutineParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE caret_buddy_routines\nSET name = ?, instructions = ?, every_minutes = ?, enabled = ?, next_run_at = ?\nWHERE id = ? AND business_id = ? AND user_id = ?")
        .bind(&params.name)
        .bind(&params.instructions)
        .bind(params.every_minutes)
        .bind(params.enabled)
        .bind(params.next_run_at)
        .bind(params.id)
        .bind(params.business_id)
        .bind(params.user_id)
        .execute(exec)
        .await
}

/// Params for `UpsertBuddy` (Go `UpsertBuddyParams`).
pub struct UpsertBuddyParams {
    pub business_id: i64,
    pub run_as_user_id: i64,
    pub enabled: bool,
    pub instructions: String,
    pub interval_minutes: i32,
    pub daily_run_cap: i32,
    pub enabled_flag: i64,
}

/// `UpsertBuddy`: `exec`.
pub async fn upsert_buddy<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpsertBuddyParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO caret_buddies (business_id, run_as_user_id, enabled, instructions, interval_minutes, daily_run_cap, next_run_at)\nVALUES (?, ?, ?, ?, ?, ?, CASE WHEN CAST(? AS SIGNED) = 1 THEN NOW() ELSE NULL END)\nON DUPLICATE KEY UPDATE\n  run_as_user_id = VALUES(run_as_user_id),\n  enabled = VALUES(enabled),\n  instructions = VALUES(instructions),\n  interval_minutes = VALUES(interval_minutes),\n  daily_run_cap = VALUES(daily_run_cap),\n  next_run_at = CASE\n    WHEN VALUES(enabled) = 0 THEN NULL\n    WHEN next_run_at IS NULL THEN NOW()\n    ELSE next_run_at\n  END")
        .bind(params.business_id)
        .bind(params.run_as_user_id)
        .bind(params.enabled)
        .bind(&params.instructions)
        .bind(params.interval_minutes)
        .bind(params.daily_run_cap)
        .bind(params.enabled_flag)
        .execute(exec)
        .await
}
