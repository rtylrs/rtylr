//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `CountScheduledReports` (Go `CountScheduledReportsParams`).
pub struct CountScheduledReportsParams {
    pub business_id: i64,
}

/// `CountScheduledReports`: `one`.
pub async fn count_scheduled_reports<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountScheduledReportsParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) AS count\nFROM insights_scheduled_reports\nWHERE business_id = ?")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `CreateScheduledReport` (Go `CreateScheduledReportParams`).
pub struct CreateScheduledReportParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub template_id: String,
    pub cadence: String,
    pub channel: String,
    pub recipient: String,
    pub is_active: String,
}

/// `CreateScheduledReport`: `execresult`.
pub async fn create_scheduled_report<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateScheduledReportParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO insights_scheduled_reports\n  (owner_id, business_id, template_id, cadence, channel, recipient, is_active)\nVALUES (?, ?, ?, ?, ?, ?, COALESCE(?, TRUE))")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(&params.template_id)
        .bind(&params.cadence)
        .bind(&params.channel)
        .bind(&params.recipient)
        .bind(&params.is_active)
        .execute(exec)
        .await
}

/// Params for `DeleteScheduledReport` (Go `DeleteScheduledReportParams`).
pub struct DeleteScheduledReportParams {
    pub id: i64,
    pub business_id: i64,
}

/// `DeleteScheduledReport`: `exec`.
pub async fn delete_scheduled_report<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DeleteScheduledReportParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("DELETE FROM insights_scheduled_reports WHERE id = ? AND business_id = ?")
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `GetScheduledReport` (Go `GetScheduledReportParams`).
pub struct GetScheduledReportParams {
    pub id: i64,
    pub business_id: i64,
}

/// `GetScheduledReport`: `one`.
pub async fn get_scheduled_report<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetScheduledReportParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, template_id, cadence, channel, recipient, is_active, created_at, updated_at\nFROM insights_scheduled_reports\nWHERE id = ? AND business_id = ?")
        .bind(params.id)
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetSettings` (Go `GetSettingsParams`).
pub struct GetSettingsParams {
    pub business_id: i64,
}

/// `GetSettings`: `one`.
pub async fn get_settings<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetSettingsParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT owner_id, business_id, default_currency, fiscal_year_start, timezone,\n       default_period, compare_previous, dashboard_widgets, favorite_reports\nFROM insights_settings\nWHERE business_id = ?")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `ListScheduledReports` (Go `ListScheduledReportsParams`).
pub struct ListScheduledReportsParams {
    pub business_id: i64,
    pub sort_by: String,
    pub sort_dir: String,
    pub limit: i32,
    pub offset: i32,
}

/// `ListScheduledReports`: `many`.
pub async fn list_scheduled_reports<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListScheduledReportsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, template_id, cadence, channel, recipient, is_active, created_at, updated_at\nFROM insights_scheduled_reports\nWHERE business_id = ?\nORDER BY\n  CASE WHEN ? = 'created_at' AND ? = 'asc' THEN created_at END ASC,\n  CASE WHEN ? = 'created_at' AND ? = 'desc' THEN created_at END DESC,\n  CASE WHEN ? = 'updated_at' AND ? = 'asc' THEN updated_at END ASC,\n  CASE WHEN ? = 'updated_at' AND ? = 'desc' THEN updated_at END DESC,\n  CASE WHEN ? = 'recipient' AND ? = 'asc' THEN LOWER(recipient) END ASC,\n  CASE WHEN ? = 'recipient' AND ? = 'desc' THEN LOWER(recipient) END DESC,\n  CASE WHEN ? = 'cadence' AND ? = 'asc' THEN cadence END ASC,\n  CASE WHEN ? = 'cadence' AND ? = 'desc' THEN cadence END DESC,\n  CASE WHEN ? = 'channel' AND ? = 'asc' THEN channel END ASC,\n  CASE WHEN ? = 'channel' AND ? = 'desc' THEN channel END DESC,\n  CASE WHEN ? = 'is_active' AND ? = 'asc' THEN is_active END ASC,\n  CASE WHEN ? = 'is_active' AND ? = 'desc' THEN is_active END DESC,\n  created_at DESC,\n  id DESC\nLIMIT ? OFFSET ?")
        .bind(params.business_id)
        .bind(&params.sort_by)
        .bind(&params.sort_dir)
        .bind(&params.sort_by)
        .bind(&params.sort_dir)
        .bind(&params.sort_by)
        .bind(&params.sort_dir)
        .bind(&params.sort_by)
        .bind(&params.sort_dir)
        .bind(&params.sort_by)
        .bind(&params.sort_dir)
        .bind(&params.sort_by)
        .bind(&params.sort_dir)
        .bind(&params.sort_by)
        .bind(&params.sort_dir)
        .bind(&params.sort_by)
        .bind(&params.sort_dir)
        .bind(&params.sort_by)
        .bind(&params.sort_dir)
        .bind(&params.sort_by)
        .bind(&params.sort_dir)
        .bind(&params.sort_by)
        .bind(&params.sort_dir)
        .bind(&params.sort_by)
        .bind(&params.sort_dir)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `UpdateScheduledReport` (Go `UpdateScheduledReportParams`).
pub struct UpdateScheduledReportParams {
    pub template_id: String,
    pub cadence: String,
    pub channel: String,
    pub recipient: String,
    pub is_active: Option<bool>,
    pub id: i64,
    pub business_id: i64,
}

/// `UpdateScheduledReport`: `exec`.
pub async fn update_scheduled_report<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateScheduledReportParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE insights_scheduled_reports\nSET template_id = ?, cadence = ?, channel = ?, recipient = ?,\n    is_active = COALESCE(?, is_active)\nWHERE id = ? AND business_id = ?")
        .bind(&params.template_id)
        .bind(&params.cadence)
        .bind(&params.channel)
        .bind(&params.recipient)
        .bind(params.is_active)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `UpsertSettings` (Go `UpsertSettingsParams`).
pub struct UpsertSettingsParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub default_currency: String,
    pub fiscal_year_start: String,
    pub timezone: String,
    pub default_period: String,
    pub compare_previous: bool,
    pub dashboard_widgets: String,
    pub favorite_reports: String,
}

/// `UpsertSettings`: `exec`.
pub async fn upsert_settings<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpsertSettingsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO insights_settings (\n  owner_id, business_id, default_currency, fiscal_year_start, timezone,\n  default_period, compare_previous, dashboard_widgets, favorite_reports\n)\nVALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)\nON DUPLICATE KEY UPDATE\n  default_currency = VALUES(default_currency),\n  fiscal_year_start = VALUES(fiscal_year_start),\n  timezone = VALUES(timezone),\n  default_period = VALUES(default_period),\n  compare_previous = VALUES(compare_previous),\n  dashboard_widgets = VALUES(dashboard_widgets),\n  favorite_reports = VALUES(favorite_reports),\n  updated_at = CURRENT_TIMESTAMP")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(&params.default_currency)
        .bind(&params.fiscal_year_start)
        .bind(&params.timezone)
        .bind(&params.default_period)
        .bind(params.compare_previous)
        .bind(&params.dashboard_widgets)
        .bind(&params.favorite_reports)
        .execute(exec)
        .await
}
