//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `AddMonthlyTokenUsage` (Go `AddMonthlyTokenUsageParams`).
pub struct AddMonthlyTokenUsageParams {
    pub business_id: i64,
    pub user_id: i64,
    pub billing_month: chrono::NaiveDate,
    pub tokens_used: i64,
}

/// `AddMonthlyTokenUsage`: `exec`.
pub async fn add_monthly_token_usage<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &AddMonthlyTokenUsageParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO ai_usage_monthly (business_id, user_id, billing_month, tokens_used, requests)\nVALUES (?, ?, ?, ?, 1)\nON DUPLICATE KEY UPDATE\n  tokens_used = tokens_used + VALUES(tokens_used),\n  requests = requests + 1")
        .bind(params.business_id)
        .bind(params.user_id)
        .bind(params.billing_month)
        .bind(params.tokens_used)
        .execute(exec)
        .await
}

/// Params for `AdjustMonthlyTokenUsage` (Go `AdjustMonthlyTokenUsageParams`).
pub struct AdjustMonthlyTokenUsageParams {
    pub delta: i64,
    pub business_id: i64,
    pub user_id: i64,
    pub billing_month: chrono::NaiveDate,
}

/// `AdjustMonthlyTokenUsage`: `exec`.
pub async fn adjust_monthly_token_usage<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &AdjustMonthlyTokenUsageParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE ai_usage_monthly\nSET tokens_used = GREATEST(0, CAST(tokens_used AS SIGNED) + ?)\nWHERE business_id = ?\n  AND user_id = ?\n  AND billing_month = ?")
        .bind(params.delta)
        .bind(params.business_id)
        .bind(params.user_id)
        .bind(params.billing_month)
        .execute(exec)
        .await
}

/// Params for `EnsureUsageRow` (Go `EnsureUsageRowParams`).
pub struct EnsureUsageRowParams {
    pub business_id: i64,
    pub user_id: i64,
    pub billing_month: chrono::NaiveDate,
}

/// `EnsureUsageRow`: `exec`.
pub async fn ensure_usage_row<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &EnsureUsageRowParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT IGNORE INTO ai_usage_monthly (business_id, user_id, billing_month, tokens_used, requests)\nVALUES (?, ?, ?, 0, 0)")
        .bind(params.business_id)
        .bind(params.user_id)
        .bind(params.billing_month)
        .execute(exec)
        .await
}

/// Params for `GetMonthlyTokenUsage` (Go `GetMonthlyTokenUsageParams`).
pub struct GetMonthlyTokenUsageParams {
    pub business_id: i64,
    pub user_id: i64,
    pub billing_month: chrono::NaiveDate,
}

/// `GetMonthlyTokenUsage`: `one`.
pub async fn get_monthly_token_usage<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetMonthlyTokenUsageParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT tokens_used FROM ai_usage_monthly\nWHERE business_id = ? AND user_id = ? AND billing_month = ?")
        .bind(params.business_id)
        .bind(params.user_id)
        .bind(params.billing_month)
        .fetch_optional(exec)
        .await
}

/// Params for `GetQuotaMode` (Go `GetQuotaModeParams`).
pub struct GetQuotaModeParams {
    pub business_id: i64,
}

/// `GetQuotaMode`: `one`.
pub async fn get_quota_mode<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetQuotaModeParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT mode FROM ai_quota_settings WHERE business_id = ?")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `ReserveMonthlyTokens` (Go `ReserveMonthlyTokensParams`).
pub struct ReserveMonthlyTokensParams {
    pub estimate: i64,
    pub business_id: i64,
    pub user_id: i64,
    pub billing_month: chrono::NaiveDate,
    pub allow_overage: bool,
    pub allowance: i64,
}

/// `ReserveMonthlyTokens`: `execresult`.
pub async fn reserve_monthly_tokens<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ReserveMonthlyTokensParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE ai_usage_monthly\nSET tokens_used = tokens_used + ?,\n    requests = requests + 1\nWHERE business_id = ?\n  AND user_id = ?\n  AND billing_month = ?\n  AND (? = TRUE OR tokens_used + ? <= ?)")
        .bind(params.estimate)
        .bind(params.business_id)
        .bind(params.user_id)
        .bind(params.billing_month)
        .bind(params.allow_overage)
        .bind(params.estimate)
        .bind(params.allowance)
        .execute(exec)
        .await
}

/// Params for `UpsertQuotaMode` (Go `UpsertQuotaModeParams`).
pub struct UpsertQuotaModeParams {
    pub business_id: i64,
    pub mode: String,
}

/// `UpsertQuotaMode`: `exec`.
pub async fn upsert_quota_mode<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpsertQuotaModeParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO ai_quota_settings (business_id, mode) VALUES (?, ?)\nON DUPLICATE KEY UPDATE mode = VALUES(mode)")
        .bind(params.business_id)
        .bind(&params.mode)
        .execute(exec)
        .await
}
