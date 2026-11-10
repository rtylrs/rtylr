//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `ClaimWebhookDelivery` (Go `ClaimWebhookDeliveryParams`).
pub struct ClaimWebhookDeliveryParams {
    pub id: u64,
}

/// `ClaimWebhookDelivery`: `execresult`.
pub async fn claim_webhook_delivery<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ClaimWebhookDeliveryParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE webhook_deliveries\nSET status = 'delivering', attempts = attempts + 1, updated_at = NOW()\nWHERE id = ? AND status IN ('pending', 'failed')")
        .bind(params.id)
        .execute(exec)
        .await
}

/// Params for `CountWebhookDeliveriesForBusiness` (Go `CountWebhookDeliveriesForBusinessParams`).
pub struct CountWebhookDeliveriesForBusinessParams {
    pub business_id: i64,
    pub webhook_id: i64,
    pub status: String,
}

/// `CountWebhookDeliveriesForBusiness`: `one`.
pub async fn count_webhook_deliveries_for_business<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountWebhookDeliveriesForBusinessParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*)\nFROM webhook_deliveries d\nWHERE d.business_id = ?\n  AND (? = 0 OR d.webhook_id = ?)\n  AND (? = '' OR d.status = ?)")
        .bind(params.business_id)
        .bind(params.webhook_id)
        .bind(params.webhook_id)
        .bind(&params.status)
        .bind(&params.status)
        .fetch_optional(exec)
        .await
}

/// Params for `CreateWebhook` (Go `CreateWebhookParams`).
pub struct CreateWebhookParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub url: String,
    pub events: String,
    pub secret: String,
}

/// `CreateWebhook`: `execresult`.
pub async fn create_webhook<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateWebhookParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO hub_webhooks (owner_id, business_id, url, events, secret)\nVALUES (?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(&params.url)
        .bind(&params.events)
        .bind(&params.secret)
        .execute(exec)
        .await
}

/// Params for `DeleteWebhook` (Go `DeleteWebhookParams`).
pub struct DeleteWebhookParams {
    pub id: i64,
    pub business_id: i64,
}

/// `DeleteWebhook`: `exec`.
pub async fn delete_webhook<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DeleteWebhookParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("DELETE FROM hub_webhooks WHERE id = ? AND business_id = ?")
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `EnqueueWebhookDelivery` (Go `EnqueueWebhookDeliveryParams`).
pub struct EnqueueWebhookDeliveryParams {
    pub business_id: i64,
    pub webhook_id: i64,
    pub event_id: u64,
}

/// `EnqueueWebhookDelivery`: `exec`.
pub async fn enqueue_webhook_delivery<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &EnqueueWebhookDeliveryParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO webhook_deliveries (business_id, webhook_id, event_id)\nVALUES (?, ?, ?)\nON DUPLICATE KEY UPDATE id = id")
        .bind(params.business_id)
        .bind(params.webhook_id)
        .bind(params.event_id)
        .execute(exec)
        .await
}

/// Params for `GetWebhook` (Go `GetWebhookParams`).
pub struct GetWebhookParams {
    pub id: i64,
    pub business_id: i64,
}

/// `GetWebhook`: `one`.
pub async fn get_webhook<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetWebhookParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, url, events, secret, is_active, last_triggered_at, created_at, updated_at\nFROM hub_webhooks\nWHERE id = ? AND business_id = ?")
        .bind(params.id)
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetWebhookForDelivery` (Go `GetWebhookForDeliveryParams`).
pub struct GetWebhookForDeliveryParams {
    pub id: i64,
    pub business_id: i64,
}

/// `GetWebhookForDelivery`: `one`.
pub async fn get_webhook_for_delivery<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetWebhookForDeliveryParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query(
        "SELECT id, url, secret, is_active\nFROM hub_webhooks\nWHERE id = ? AND business_id = ?",
    )
    .bind(params.id)
    .bind(params.business_id)
    .fetch_optional(exec)
    .await
}

/// Params for `GetWebhookNotificationOwner` (Go `GetWebhookNotificationOwnerParams`).
pub struct GetWebhookNotificationOwnerParams {
    pub id: i64,
    pub business_id: i64,
}

/// `GetWebhookNotificationOwner`: `one`.
pub async fn get_webhook_notification_owner<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetWebhookNotificationOwnerParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT owner_id\nFROM hub_webhooks\nWHERE id = ? AND business_id = ?")
        .bind(params.id)
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `ListDueWebhookDeliveries` (Go `ListDueWebhookDeliveriesParams`).
pub struct ListDueWebhookDeliveriesParams {
    pub limit: i32,
}

/// `ListDueWebhookDeliveries`: `many`.
pub async fn list_due_webhook_deliveries<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListDueWebhookDeliveriesParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, business_id, webhook_id, event_id, attempts, max_attempts\nFROM webhook_deliveries\nWHERE status IN ('pending', 'failed')\n  AND next_attempt_at <= NOW()\nORDER BY next_attempt_at ASC, id ASC\nLIMIT ?")
        .bind(params.limit)
        .execute(exec)
        .await
}

/// Params for `ListRegistrationsWithRecentDeadDeliveries` (Go `ListRegistrationsWithRecentDeadDeliveriesParams`).
pub struct ListRegistrationsWithRecentDeadDeliveriesParams {
    pub within_hours: String,
    pub notified_on: chrono::NaiveDate,
    pub limit: i32,
}

/// `ListRegistrationsWithRecentDeadDeliveries`: `many`.
pub async fn list_registrations_with_recent_dead_deliveries<
    'e,
    E: sqlx::Executor<'e, Database = MySql>,
>(
    exec: E,
    params: &ListRegistrationsWithRecentDeadDeliveriesParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT d.business_id, d.webhook_id, w.url,\n       COUNT(*) AS dead_count,\n       MAX(d.updated_at) AS last_dead_at\nFROM webhook_deliveries d\nJOIN hub_webhooks w ON w.id = d.webhook_id AND w.business_id = d.business_id\nWHERE d.status = 'dead'\n  AND d.updated_at >= DATE_SUB(NOW(), INTERVAL ? HOUR)\n  AND NOT EXISTS (\n        SELECT 1\n        FROM webhook_dead_delivery_notices n\n        WHERE n.business_id = d.business_id\n          AND n.webhook_id = d.webhook_id\n          AND n.notice_date = ?\n      )\nGROUP BY d.business_id, d.webhook_id, w.url\nORDER BY d.webhook_id ASC\nLIMIT ?")
        .bind(&params.within_hours)
        .bind(params.notified_on)
        .bind(params.limit)
        .execute(exec)
        .await
}

/// Params for `ListWebhookDeliveriesForBusiness` (Go `ListWebhookDeliveriesForBusinessParams`).
pub struct ListWebhookDeliveriesForBusinessParams {
    pub business_id: i64,
    pub webhook_id: i64,
    pub status: String,
    pub limit: i32,
    pub offset: i32,
}

/// `ListWebhookDeliveriesForBusiness`: `many`.
pub async fn list_webhook_deliveries_for_business<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListWebhookDeliveriesForBusinessParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT d.id, d.webhook_id, d.event_id, d.status,\n       d.attempts, d.max_attempts, d.response_status, d.last_error,\n       d.next_attempt_at, d.delivered_at, d.created_at\nFROM webhook_deliveries d\nWHERE d.business_id = ?\n  AND (? = 0 OR d.webhook_id = ?)\n  AND (? = '' OR d.status = ?)\nORDER BY d.id DESC\nLIMIT ? OFFSET ?")
        .bind(params.business_id)
        .bind(params.webhook_id)
        .bind(params.webhook_id)
        .bind(&params.status)
        .bind(&params.status)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `ListWebhooks` (Go `ListWebhooksParams`).
pub struct ListWebhooksParams {
    pub business_id: i64,
}

/// `ListWebhooks`: `many`.
pub async fn list_webhooks<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListWebhooksParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, url, events, secret, is_active, last_triggered_at, created_at, updated_at\nFROM hub_webhooks\nWHERE business_id = ?\nORDER BY created_at DESC")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `ListWebhooksForEvent` (Go `ListWebhooksForEventParams`).
pub struct ListWebhooksForEventParams {
    pub business_id: i64,
    pub event_type: String,
}

/// `ListWebhooksForEvent`: `many`.
pub async fn list_webhooks_for_event<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListWebhooksForEventParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id\nFROM hub_webhooks\nWHERE business_id = ?\n  AND is_active = TRUE\n  AND JSON_CONTAINS(events, JSON_QUOTE(?))")
        .bind(params.business_id)
        .bind(&params.event_type)
        .execute(exec)
        .await
}

/// Params for `MarkWebhookDelivered` (Go `MarkWebhookDeliveredParams`).
pub struct MarkWebhookDeliveredParams {
    pub response_status: Option<i16>,
    pub id: u64,
}

/// `MarkWebhookDelivered`: `exec`.
pub async fn mark_webhook_delivered<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &MarkWebhookDeliveredParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE webhook_deliveries\nSET status = 'delivered',\n    delivered_at = NOW(),\n    response_status = ?,\n    last_error = NULL,\n    updated_at = NOW()\nWHERE id = ?")
        .bind(params.response_status)
        .bind(params.id)
        .execute(exec)
        .await
}

/// Params for `MarkWebhookDeliveryDead` (Go `MarkWebhookDeliveryDeadParams`).
pub struct MarkWebhookDeliveryDeadParams {
    pub last_error: Option<String>,
    pub id: u64,
}

/// `MarkWebhookDeliveryDead`: `exec`.
pub async fn mark_webhook_delivery_dead<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &MarkWebhookDeliveryDeadParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE webhook_deliveries\nSET status = 'dead', last_error = ?, updated_at = NOW()\nWHERE id = ?")
        .bind(&params.last_error)
        .bind(params.id)
        .execute(exec)
        .await
}

/// Params for `MarkWebhookDeliveryFailed` (Go `MarkWebhookDeliveryFailedParams`).
pub struct MarkWebhookDeliveryFailedParams {
    pub retry_after_seconds: String,
    pub response_status: Option<i16>,
    pub last_error: Option<String>,
    pub id: u64,
}

/// `MarkWebhookDeliveryFailed`: `exec`.
pub async fn mark_webhook_delivery_failed<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &MarkWebhookDeliveryFailedParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE webhook_deliveries\nSET status = IF(attempts >= max_attempts, 'dead', 'failed'),\n    next_attempt_at = DATE_ADD(NOW(), INTERVAL ? SECOND),\n    response_status = ?,\n    last_error = ?,\n    updated_at = NOW()\nWHERE id = ?")
        .bind(&params.retry_after_seconds)
        .bind(params.response_status)
        .bind(&params.last_error)
        .bind(params.id)
        .execute(exec)
        .await
}

/// Params for `RecordDeadDeliveryNotice` (Go `RecordDeadDeliveryNoticeParams`).
pub struct RecordDeadDeliveryNoticeParams {
    pub business_id: i64,
    pub webhook_id: i64,
    pub notice_date: chrono::NaiveDate,
}

/// `RecordDeadDeliveryNotice`: `exec`.
pub async fn record_dead_delivery_notice<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &RecordDeadDeliveryNoticeParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO webhook_dead_delivery_notices (business_id, webhook_id, notice_date)\nVALUES (?, ?, ?)\nON DUPLICATE KEY UPDATE notice_date = VALUES(notice_date)")
        .bind(params.business_id)
        .bind(params.webhook_id)
        .bind(params.notice_date)
        .execute(exec)
        .await
}

/// Params for `RecordWebhookTriggered` (Go `RecordWebhookTriggeredParams`).
pub struct RecordWebhookTriggeredParams {
    pub id: i64,
    pub business_id: i64,
}

/// `RecordWebhookTriggered`: `exec`.
pub async fn record_webhook_triggered<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &RecordWebhookTriggeredParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query(
        "UPDATE hub_webhooks\nSET last_triggered_at = NOW()\nWHERE id = ? AND business_id = ?",
    )
    .bind(params.id)
    .bind(params.business_id)
    .execute(exec)
    .await
}

/// Params for `RecoverAbandonedWebhookDeliveries` (Go `RecoverAbandonedWebhookDeliveriesParams`).
pub struct RecoverAbandonedWebhookDeliveriesParams {
    pub stale_after_seconds: String,
}

/// `RecoverAbandonedWebhookDeliveries`: `execresult`.
pub async fn recover_abandoned_webhook_deliveries<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &RecoverAbandonedWebhookDeliveriesParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE webhook_deliveries\nSET status = IF(attempts >= max_attempts, 'dead', 'failed'),\n    next_attempt_at = NOW(),\n    last_error = 'delivery was abandoned in flight and has been requeued',\n    updated_at = NOW()\nWHERE status = 'delivering'\n  AND updated_at < DATE_SUB(NOW(), INTERVAL ? SECOND)")
        .bind(&params.stale_after_seconds)
        .execute(exec)
        .await
}

/// Params for `RotateWebhookSecret` (Go `RotateWebhookSecretParams`).
pub struct RotateWebhookSecretParams {
    pub secret: String,
    pub id: i64,
    pub business_id: i64,
}

/// `RotateWebhookSecret`: `execresult`.
pub async fn rotate_webhook_secret<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &RotateWebhookSecretParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE hub_webhooks SET secret = ? WHERE id = ? AND business_id = ?")
        .bind(&params.secret)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `UpdateWebhook` (Go `UpdateWebhookParams`).
pub struct UpdateWebhookParams {
    pub url: String,
    pub events: String,
    pub is_active: bool,
    pub id: i64,
    pub business_id: i64,
}

/// `UpdateWebhook`: `exec`.
pub async fn update_webhook<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateWebhookParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE hub_webhooks\nSET url = ?, events = ?, is_active = ?\nWHERE id = ? AND business_id = ?")
        .bind(&params.url)
        .bind(&params.events)
        .bind(params.is_active)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}
