//! Ports (traits) the application services need. Mirrors the interfaces
//! in `internal/application`.

use crate::domain::{Delivery, DeliveryRecord, Event, FailingRegistration, Registration, Webhook};
use chrono::{DateTime, Utc};
use thiserror::Error;

#[derive(Debug, Error, PartialEq)]
pub enum AppError {
    #[error("not found")]
    NotFound,
    #[error("{0}")]
    Repo(String),
}

/// Registration + delivery listing port.
pub trait Repository: Send + Sync {
    fn list(&self, business_id: i64) -> Result<Vec<Webhook>, AppError>;
    fn create(
        &self,
        owner_id: i64,
        business_id: i64,
        url: &str,
        secret: &str,
        events: &[String],
    ) -> Result<i64, AppError>;
    fn get(&self, business_id: i64, webhook_id: i64) -> Result<Webhook, AppError>;
    fn update(
        &self,
        business_id: i64,
        webhook_id: i64,
        url: &str,
        events: &[String],
        is_active: bool,
    ) -> Result<(), AppError>;
    fn delete(&self, business_id: i64, webhook_id: i64) -> Result<(), AppError>;
    fn rotate_secret(
        &self,
        business_id: i64,
        webhook_id: i64,
        secret: &str,
    ) -> Result<(), AppError>;
    fn list_deliveries_for_business(
        &self,
        business_id: i64,
        filter: &crate::domain::DeliveryFilter,
    ) -> Result<(Vec<DeliveryRecord>, i64), AppError>;
}

/// Delivery queue port used by the sender.
pub trait DeliveryRepository: Send + Sync {
    fn list_due_deliveries(&self, limit: i32) -> Result<Vec<Delivery>, AppError>;
    fn claim_delivery(&self, id: u64) -> Result<(), AppError>;
    fn mark_delivered(&self, id: u64, status: i32) -> Result<(), AppError>;
    fn mark_failed(
        &self,
        id: u64,
        retry_at: i64,
        status: i32,
        last_error: &str,
    ) -> Result<(), AppError>;
    fn mark_dead(&self, id: u64, last_error: &str) -> Result<(), AppError>;
    fn recover_abandoned(&self, abandoned_after_secs: i64) -> Result<i64, AppError>;
    fn get_registration_for_delivery(
        &self,
        webhook_id: i64,
        business_id: i64,
    ) -> Result<Registration, AppError>;
    fn record_triggered(&self, webhook_id: i64, business_id: i64) -> Result<(), AppError>;
}

/// Event outbox reader (cross-schema read through the events crate).
pub trait EventReader: Send + Sync {
    fn get_event(&self, event_id: u64, business_id: i64) -> Result<Event, AppError>;
}

/// Notification port for the dead-delivery notices.
pub trait NotificationRepository: Send + Sync {
    /// `within_hours` bounds which failures are worth mentioning.
    fn list_failing_registrations(
        &self,
        within_hours: i64,
        limit: i32,
        since: DateTime<Utc>,
    ) -> Result<Vec<FailingRegistration>, AppError>;
    fn get_notification_owner(&self, webhook_id: i64, business_id: i64) -> Result<i64, AppError>;
    fn record_dead_delivery_notice(
        &self,
        business_id: i64,
        webhook_id: i64,
        at: DateTime<Utc>,
    ) -> Result<(), AppError>;
}

/// Email queue sink.
pub trait NotificationQueue: Send + Sync {
    fn enqueue(&self, email: &NotificationEmail) -> Result<(), AppError>;
}

/// Who to tell about a failing registration: (email, name, still an
/// active member). `Ok(None)` means no recipient (creator left or is
/// inactive); an `Err` is a lookup failure and must NOT be recorded as a
/// notice.
pub trait NotificationContactResolver: Send + Sync {
    fn resolve(
        &self,
        business_id: i64,
        user_id: i64,
    ) -> Result<Option<(String, String, bool)>, AppError>;
}

/// One notification email, queued rather than sent inline.
#[derive(Debug, Clone, PartialEq)]
pub struct NotificationEmail {
    pub to: String,
    pub subject: String,
    pub html: String,
    pub text: String,
    pub dedupe_key: String,
}
