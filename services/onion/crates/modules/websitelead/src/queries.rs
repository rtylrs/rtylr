//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `CreateNewsletterSubscriber` (Go `CreateNewsletterSubscriberParams`).
pub struct CreateNewsletterSubscriberParams {
    pub email: String,
    pub source: String,
    pub post_slug: Option<String>,
    pub post_title: Option<String>,
    pub post_category: Option<String>,
    pub language: String,
}

/// `CreateNewsletterSubscriber`: `execresult`.
pub async fn create_newsletter_subscriber<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateNewsletterSubscriberParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO voxire_newsletter_subscribers (email, source, post_slug, post_title, post_category, language)\nVALUES (?, ?, ?, ?, ?, ?)")
        .bind(&params.email)
        .bind(&params.source)
        .bind(&params.post_slug)
        .bind(&params.post_title)
        .bind(&params.post_category)
        .bind(&params.language)
        .execute(exec)
        .await
}

/// Params for `CreateWebsiteLead` (Go `CreateWebsiteLeadParams`).
pub struct CreateWebsiteLeadParams {
    pub name: String,
    pub email: String,
    pub phone: Option<String>,
    pub company: Option<String>,
    pub service: Option<String>,
    pub project: Option<String>,
    pub budget: Option<String>,
    pub source: Option<String>,
}

/// `CreateWebsiteLead`: `execresult`.
pub async fn create_website_lead<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateWebsiteLeadParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO voxire_website_leads (name, email, phone, company, service, project, budget, source)\nVALUES (?, ?, ?, ?, ?, ?, ?, ?)")
        .bind(&params.name)
        .bind(&params.email)
        .bind(&params.phone)
        .bind(&params.company)
        .bind(&params.service)
        .bind(&params.project)
        .bind(&params.budget)
        .bind(&params.source)
        .execute(exec)
        .await
}

/// Params for `ListUnsyncedNewsletterSubscribers` (Go `ListUnsyncedNewsletterSubscribersParams`).
pub struct ListUnsyncedNewsletterSubscribersParams {
    pub resend_sync_attempts: i32,
    pub limit: i32,
}

/// `ListUnsyncedNewsletterSubscribers`: `many`.
pub async fn list_unsynced_newsletter_subscribers<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListUnsyncedNewsletterSubscribersParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, email\nFROM voxire_newsletter_subscribers\nWHERE resend_synced_at IS NULL AND resend_sync_attempts < ?\nORDER BY id\nLIMIT ?")
        .bind(params.resend_sync_attempts)
        .bind(params.limit)
        .execute(exec)
        .await
}

/// Params for `MarkNewsletterSubscriberSyncFailed` (Go `MarkNewsletterSubscriberSyncFailedParams`).
pub struct MarkNewsletterSubscriberSyncFailedParams {
    pub resend_sync_error: Option<String>,
    pub id: i64,
}

/// `MarkNewsletterSubscriberSyncFailed`: `exec`.
pub async fn mark_newsletter_subscriber_sync_failed<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &MarkNewsletterSubscriberSyncFailedParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE voxire_newsletter_subscribers\nSET resend_sync_attempts = resend_sync_attempts + 1, resend_sync_error = ?\nWHERE id = ?")
        .bind(&params.resend_sync_error)
        .bind(params.id)
        .execute(exec)
        .await
}

/// Params for `MarkNewsletterSubscriberSynced` (Go `MarkNewsletterSubscriberSyncedParams`).
pub struct MarkNewsletterSubscriberSyncedParams {
    pub resend_contact_id: Option<String>,
    pub id: i64,
}

/// `MarkNewsletterSubscriberSynced`: `exec`.
pub async fn mark_newsletter_subscriber_synced<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &MarkNewsletterSubscriberSyncedParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE voxire_newsletter_subscribers\nSET resend_contact_id = ?, resend_synced_at = CURRENT_TIMESTAMP, resend_sync_error = NULL\nWHERE id = ?")
        .bind(&params.resend_contact_id)
        .bind(params.id)
        .execute(exec)
        .await
}
