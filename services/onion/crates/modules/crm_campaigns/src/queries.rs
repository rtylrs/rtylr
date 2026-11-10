//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `CountCampaigns` (Go `CountCampaignsParams`).
pub struct CountCampaignsParams {
    pub business_id: i64,
}

/// `CountCampaigns`: `one`.
pub async fn count_campaigns<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountCampaignsParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) FROM crm_campaigns WHERE business_id = ?")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `CreateCampaign` (Go `CreateCampaignParams`).
pub struct CreateCampaignParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub name: String,
    pub description: Option<String>,
    pub channel: String,
    pub segment_id: Option<i64>,
    pub subject: Option<String>,
    pub content: String,
    pub scheduled_at: Option<chrono::DateTime<chrono::Utc>>,
}

/// `CreateCampaign`: `execresult`.
pub async fn create_campaign<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateCampaignParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO crm_campaigns (owner_id, business_id, name, description, channel, segment_id, subject, content, scheduled_at)\nVALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(&params.name)
        .bind(&params.description)
        .bind(&params.channel)
        .bind(params.segment_id)
        .bind(&params.subject)
        .bind(&params.content)
        .bind(params.scheduled_at)
        .execute(exec)
        .await
}

/// Params for `GetCampaignByID` (Go `GetCampaignByIDParams`).
pub struct GetCampaignByIDParams {
    pub id: i64,
    pub business_id: i64,
}

/// `GetCampaignByID`: `one`.
pub async fn get_campaign_by_id<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetCampaignByIDParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, name, description, channel, status, segment_id, subject,\n       content, scheduled_at, sent_at, audience_count, sent_count, delivered_count,\n       opened_count, clicked_count, created_at, updated_at\nFROM crm_campaigns WHERE id = ? AND business_id = ?")
        .bind(params.id)
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `ListCampaigns` (Go `ListCampaignsParams`).
pub struct ListCampaignsParams {
    pub business_id: i64,
}

/// `ListCampaigns`: `many`.
pub async fn list_campaigns<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListCampaignsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, name, description, channel, status, segment_id, subject,\n       content, scheduled_at, sent_at, audience_count, sent_count, delivered_count,\n       opened_count, clicked_count, created_at, updated_at\nFROM crm_campaigns WHERE business_id = ? ORDER BY created_at DESC LIMIT 100")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `ListCampaignsByStatus` (Go `ListCampaignsByStatusParams`).
pub struct ListCampaignsByStatusParams {
    pub business_id: i64,
    pub status: String,
}

/// `ListCampaignsByStatus`: `many`.
pub async fn list_campaigns_by_status<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListCampaignsByStatusParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, name, description, channel, status, segment_id, subject,\n       content, scheduled_at, sent_at, audience_count, sent_count, delivered_count,\n       opened_count, clicked_count, created_at, updated_at\nFROM crm_campaigns WHERE business_id = ? AND status = ? ORDER BY created_at DESC LIMIT 100")
        .bind(params.business_id)
        .bind(&params.status)
        .execute(exec)
        .await
}

/// Params for `ListCampaignsPaginated` (Go `ListCampaignsPaginatedParams`).
pub struct ListCampaignsPaginatedParams {
    pub business_id: i64,
    pub limit: i32,
    pub offset: i32,
}

/// `ListCampaignsPaginated`: `many`.
pub async fn list_campaigns_paginated<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListCampaignsPaginatedParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, name, description, channel, status, segment_id, subject,\n       content, scheduled_at, sent_at, audience_count, sent_count, delivered_count,\n       opened_count, clicked_count, created_at, updated_at\nFROM crm_campaigns WHERE business_id = ? ORDER BY created_at DESC\nLIMIT ? OFFSET ?")
        .bind(params.business_id)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `UpdateCampaign` (Go `UpdateCampaignParams`).
pub struct UpdateCampaignParams {
    pub name: String,
    pub description: Option<String>,
    pub channel: String,
    pub segment_id: Option<i64>,
    pub subject: Option<String>,
    pub content: String,
    pub scheduled_at: Option<chrono::DateTime<chrono::Utc>>,
    pub id: i64,
    pub business_id: i64,
}

/// `UpdateCampaign`: `execresult`.
pub async fn update_campaign<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateCampaignParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE crm_campaigns\nSET name = ?, description = ?, channel = ?, segment_id = ?, subject = ?,\n    content = ?, scheduled_at = ?\nWHERE id = ? AND business_id = ?")
        .bind(&params.name)
        .bind(&params.description)
        .bind(&params.channel)
        .bind(params.segment_id)
        .bind(&params.subject)
        .bind(&params.content)
        .bind(params.scheduled_at)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `UpdateCampaignStatus` (Go `UpdateCampaignStatusParams`).
pub struct UpdateCampaignStatusParams {
    pub status: String,
    pub id: i64,
    pub business_id: i64,
}

/// `UpdateCampaignStatus`: `execresult`.
pub async fn update_campaign_status<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateCampaignStatusParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE crm_campaigns SET status = ? WHERE id = ? AND business_id = ?")
        .bind(&params.status)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}
