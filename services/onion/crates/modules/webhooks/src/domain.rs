//! Domain types and sentinel errors. Mirrors `internal/domain`.

use chrono::{DateTime, Utc};
use thiserror::Error;

#[derive(Debug, Error, PartialEq)]
pub enum WebhookError {
    #[error("webhook not found")]
    NotFound,
    #[error("webhook delivery already claimed")]
    DeliveryNotClaimed,
    #[error("webhook registration not found")]
    DeliveryRegistrationNotFound,
    #[error("webhook notification recipient not found")]
    NotificationRecipientNotFound,
}

/// One registered webhook subscription.
#[derive(Debug, Clone, PartialEq)]
pub struct Webhook {
    pub id: i64,
    pub owner_id: i64,
    pub business_id: i64,
    pub url: String,
    pub events: Vec<String>,
    pub secret: String,
    pub is_active: bool,
    pub last_triggered_at: Option<DateTime<Utc>>,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
}

/// One queued attempt to send one event to one registration. The
/// dispatcher (`crates/events`) creates these; the sender drains them.
#[derive(Debug, Clone, PartialEq)]
pub struct Delivery {
    pub id: u64,
    pub business_id: i64,
    pub webhook_id: i64,
    pub event_id: u64,
    /// Count after the claim, so a delivery on its last permitted try
    /// already reads `attempts == max_attempts`.
    pub attempts: i32,
    pub max_attempts: i32,
}

/// The subscriber a delivery is addressed to.
#[derive(Debug, Clone, PartialEq)]
pub struct Registration {
    pub id: i64,
    pub url: String,
    pub secret: String,
    pub is_active: bool,
}

/// The outbox row a delivery was created from.
#[derive(Debug, Clone, PartialEq)]
pub struct Event {
    pub id: u64,
    pub event_type: String,
    pub subject_type: String,
    pub subject_id: i64,
    pub payload: Vec<u8>,
    pub occurred_at: DateTime<Utc>,
}

/// Delivery listing filter. A zero webhook_id and an empty status mean
/// no filter rather than match nothing.
#[derive(Debug, Clone, Default, PartialEq)]
pub struct DeliveryFilter {
    pub webhook_id: i64,
    pub status: String,
    pub limit: i32,
    pub offset: i64,
}

/// A delivery attempt as the registration owner sees it, with the event
/// type resolved so a row means something on its own.
#[derive(Debug, Clone, PartialEq)]
pub struct DeliveryRecord {
    /// `id`/`event_id` are u64 because both queue tables key on
    /// `BIGINT UNSIGNED`; they narrow to i64 at the contract boundary.
    pub id: u64,
    pub webhook_id: i64,
    pub event_id: u64,
    pub event_type: String,
    pub status: String,
    pub attempts: i32,
    pub max_attempts: i32,
    pub response_status: i32,
    pub last_error: String,
    pub next_attempt_at: Option<DateTime<Utc>>,
    pub delivered_at: Option<DateTime<Utc>>,
    pub created_at: DateTime<Utc>,
}

/// A registration that has dead-lettered deliveries recently, summarised
/// for one notification rather than one per lost event.
#[derive(Debug, Clone, PartialEq)]
pub struct FailingRegistration {
    pub business_id: i64,
    pub webhook_id: i64,
    pub url: String,
    pub dead_count: i64,
    pub last_dead_at: DateTime<Utc>,
}

/// Who to tell about a failing registration: the user who created it,
/// confirmed still an active member of its business.
#[derive(Debug, Clone, PartialEq)]
pub struct NotificationRecipient {
    pub user_id: i64,
    pub email: String,
    pub name: String,
}

#[cfg(test)]
mod tests {
    use super::WebhookError;

    #[test]
    fn sentinel_errors() {
        assert_eq!(
            WebhookError::DeliveryNotClaimed.to_string(),
            "webhook delivery already claimed"
        );
    }
}
