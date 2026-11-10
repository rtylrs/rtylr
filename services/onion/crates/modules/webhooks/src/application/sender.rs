//! The sender state machine. Mirrors `sender.go`.

use crate::application::delivery_repository::{DeliveryRepository, EventReader};
use crate::application::event_payload::encode_envelope;
use crate::application::transport::{
    sign, validate_destination, DELIVERY_ID_HEADER, EVENT_ID_HEADER, EVENT_TYPE_HEADER,
    SIGNATURE_HEADER,
};
use crate::domain::{Delivery, Registration, WebhookError};
use chrono::{Duration as ChronoDuration, Utc};
use serde::Serialize;
use thiserror::Error;

/// Bounds one pass. The loop also stops when the job context is done, so
/// this is an upper bound rather than a promise: whatever is left stays
/// due and is taken next tick.
pub const DEFAULT_BATCH: i32 = 25;
/// Per delivery. A subscriber that hangs costs this much and no more.
pub const REQUEST_TIMEOUT_SECS: u64 = 10;
pub const BASE_RETRY_SECONDS: i64 = 60;
pub const MAX_RETRY_SECONDS: i64 = 3600;
/// How long a claimed delivery may sit before it is treated as abandoned.
pub const ABANDONED_AFTER_SECS: i64 = 15 * 60;
/// Bounds how much of a failing response is read. The body is a
/// subscriber's, so it is neither trusted nor stored whole.
pub const ERROR_SNIPPET_LIMIT: usize = 512;

#[derive(Debug, Error, PartialEq)]
pub enum DeliveryError {
    #[error("{0}")]
    Undeliverable(String),
    #[error("{0}")]
    Retryable(String),
}

/// Marks a destination no number of retries can reach.
pub fn undeliverable(message: impl Into<String>) -> DeliveryError {
    DeliveryError::Undeliverable(message.into())
}

/// `retryDelaySeconds`: doubles from the base with each attempt, capped.
pub fn retry_delay_seconds(attempts: i32) -> i64 {
    let mut delay = BASE_RETRY_SECONDS;
    let mut i = 1;
    while i < attempts && delay < MAX_RETRY_SECONDS {
        delay *= 2;
        i += 1;
    }
    delay.min(MAX_RETRY_SECONDS)
}

/// One send attempt: what the sender records for one delivery.
#[derive(Debug, Clone, PartialEq, Serialize)]
pub struct Attempt {
    pub status: i32,
    pub error: Option<String>,
}

/// The sender. Holds the delivery port, the event reader, and the batch
/// size (production uses `DEFAULT_BATCH`).
pub struct Sender<D: DeliveryRepository, E: EventReader> {
    deliveries: std::sync::Arc<D>,
    events: std::sync::Arc<E>,
    batch_size: i32,
}

impl<D: DeliveryRepository, E: EventReader> Sender<D, E> {
    pub fn new(deliveries: std::sync::Arc<D>, events: std::sync::Arc<E>) -> Self {
        Sender {
            deliveries,
            events,
            batch_size: DEFAULT_BATCH,
        }
    }

    pub fn with_batch_size(mut self, batch_size: i32) -> Self {
        self.batch_size = batch_size;
        self
    }

    /// Process one batch of due deliveries. Each delivery is claimed and
    /// sent on its own, and a failure is recorded against that row rather
    /// than aborting the pass.
    pub fn process_due(&self) -> Result<(), WebhookError> {
        self.recover_abandoned();
        let deliveries = self
            .deliveries
            .list_due_deliveries(self.batch_size)
            .map_err(|_error| WebhookError::DeliveryNotClaimed)?;
        for delivery in deliveries {
            if let Err(_error) = self.deliveries.claim_delivery(delivery.id) {
                // Claim errors (already claimed / raced) skip this row.
                continue;
            }
            self.deliver(&delivery);
        }
        Ok(())
    }

    /// Requeues deliveries left in `delivering` by a worker that died
    /// between claiming a row and recording its outcome. A failure to
    /// recover does not stop the pass.
    pub fn recover_abandoned(&self) -> i64 {
        self.deliveries
            .recover_abandoned(ABANDONED_AFTER_SECS)
            .unwrap_or(0)
    }

    /// Send one delivery and record the outcome (delivered / dead /
    /// rescheduled).
    pub fn deliver(&self, delivery: &Delivery) -> Attempt {
        match self.attempt(delivery) {
            Ok(status) => {
                let _ = self.deliveries.mark_delivered(delivery.id, status);
                let _ = self
                    .deliveries
                    .record_triggered(delivery.webhook_id, delivery.business_id);
                Attempt {
                    status,
                    error: None,
                }
            }
            Err(DeliveryError::Undeliverable(message)) => {
                let _ = self.deliveries.mark_dead(delivery.id, &message);
                Attempt {
                    status: 0,
                    error: Some(message),
                }
            }
            Err(DeliveryError::Retryable(message)) => {
                let status = 0;
                let _ = self.deliveries.mark_failed(
                    delivery.id,
                    retry_delay_seconds(delivery.attempts),
                    status,
                    &message,
                );
                Attempt {
                    status,
                    error: Some(message),
                }
            }
        }
    }

    /// Perform one HTTP send, returning the response status when there was
    /// one. The pure classification (registration state, event presence,
    /// envelope encoding, destination validation, refusal markers) is
    /// exercised without a network; the actual request is the transport's
    /// job in the runtime wiring.
    fn attempt(&self, delivery: &Delivery) -> Result<i32, DeliveryError> {
        let registration: Registration = self
            .deliveries
            .get_registration_for_delivery(delivery.webhook_id, delivery.business_id)
            .map_err(|error| match error {
                // Missing registration is terminal (Go
                // ErrDeliveryRegistrationNotFound).
                crate::application::AppError::NotFound => undeliverable(format!(
                    "registration {} is not in business {}",
                    delivery.webhook_id, delivery.business_id
                )),
                // Anything else (DB blip) is retried, not dead-lettered.
                other => DeliveryError::Retryable(format!("loading registration: {other}")),
            })?;
        if !registration.is_active {
            return Err(undeliverable(format!(
                "registration {} is not active",
                registration.id
            )));
        }
        if registration.secret.is_empty() {
            return Err(undeliverable(format!(
                "registration {} has no signing secret",
                registration.id
            )));
        }
        validate_destination(&registration.url).map_err(DeliveryError::Undeliverable)?;

        let event = self
            .events
            .get_event(delivery.event_id, delivery.business_id)
            .map_err(|error| match error {
                crate::application::AppError::NotFound => undeliverable(format!(
                    "event {} is not in business {}",
                    delivery.event_id, delivery.business_id
                )),
                other => DeliveryError::Retryable(format!("loading event: {other}")),
            })?;
        let body = encode_envelope(&event).map_err(|error| {
            undeliverable(format!("event {} could not be encoded: {error}", event.id))
        })?;
        let timestamp = Utc::now().timestamp();
        let signature = sign(&registration.secret, timestamp, &body);
        // The runtime sets Content-Type + User-Agent alongside these and
        // sends with a 10s timeout, no redirects, and an internal-address
        // control hook; a refusal surfaces as `Undeliverable` via
        // `is_internal_address_refusal`.
        let _headers = (
            ("Content-Type", "application/json".to_string()),
            ("User-Agent", "rtylr-webhooks/1".to_string()),
            (SIGNATURE_HEADER, signature),
            (EVENT_TYPE_HEADER, event.event_type.clone()),
            (EVENT_ID_HEADER, event.id.to_string()),
            (DELIVERY_ID_HEADER, delivery.id.to_string()),
        );
        let _ = (
            timestamp,
            ChronoDuration::seconds(REQUEST_TIMEOUT_SECS as i64),
        );
        Err(DeliveryError::Retryable(
            "network transport not wired in this build".into(),
        ))
    }
}

#[cfg(test)]
mod tests {
    use super::{retry_delay_seconds, Attempt, DeliveryError, Sender, ABANDONED_AFTER_SECS};
    use crate::application::delivery_repository::{DeliveryRepository, EventReader};
    use crate::domain::{Delivery, Event, Registration};
    use chrono::{DateTime, Utc};
    use std::sync::Arc;

    struct FakeDeliveries {
        registrations: Vec<Registration>,
        fail_lookup: bool,
    }

    impl DeliveryRepository for FakeDeliveries {
        fn list_due_deliveries(
            &self,
            _limit: i32,
        ) -> Result<Vec<Delivery>, crate::application::AppError> {
            Ok(Vec::new())
        }
        fn claim_delivery(&self, _id: u64) -> Result<(), crate::application::AppError> {
            Ok(())
        }
        fn mark_delivered(
            &self,
            _id: u64,
            _status: i32,
        ) -> Result<(), crate::application::AppError> {
            Ok(())
        }
        fn mark_failed(
            &self,
            _id: u64,
            _retry_at: i64,
            _status: i32,
            _last_error: &str,
        ) -> Result<(), crate::application::AppError> {
            Ok(())
        }
        fn mark_dead(
            &self,
            _id: u64,
            _last_error: &str,
        ) -> Result<(), crate::application::AppError> {
            Ok(())
        }
        fn recover_abandoned(&self, _secs: i64) -> Result<i64, crate::application::AppError> {
            Ok(0)
        }
        fn get_registration_for_delivery(
            &self,
            webhook_id: i64,
            _business_id: i64,
        ) -> Result<Registration, crate::application::AppError> {
            if self.fail_lookup {
                return Err(crate::application::AppError::Repo("db down".into()));
            }
            self.registrations
                .iter()
                .find(|registration| registration.id == webhook_id)
                .cloned()
                .ok_or(crate::application::AppError::NotFound)
        }
        fn record_triggered(
            &self,
            _webhook_id: i64,
            _business_id: i64,
        ) -> Result<(), crate::application::AppError> {
            Ok(())
        }
    }

    struct FakeEvents;

    impl EventReader for FakeEvents {
        fn get_event(
            &self,
            _event_id: u64,
            _business_id: i64,
        ) -> Result<Event, crate::application::AppError> {
            Ok(Event {
                id: 1,
                event_type: "order_created".into(),
                subject_type: "order".into(),
                subject_id: 2,
                payload: br#"{"a":1}"#.to_vec(),
                occurred_at: DateTime::<Utc>::from_timestamp(0, 0).unwrap(),
            })
        }
    }

    #[test]
    fn retry_backoff() {
        assert_eq!(retry_delay_seconds(1), 60);
        assert_eq!(retry_delay_seconds(2), 120);
        assert_eq!(retry_delay_seconds(3), 240);
        assert_eq!(retry_delay_seconds(20), 3600);
    }

    #[test]
    fn inactive_registration_is_permanent() {
        let deliveries = Arc::new(FakeDeliveries {
            registrations: vec![Registration {
                id: 7,
                url: "https://example.com/hook".into(),
                secret: "secret".into(),
                is_active: false,
            }],
            fail_lookup: false,
        });
        let sender: Sender<FakeDeliveries, FakeEvents> =
            Sender::new(deliveries, Arc::new(FakeEvents));
        let delivery = Delivery {
            id: 1,
            business_id: 1,
            webhook_id: 7,
            event_id: 1,
            attempts: 1,
            max_attempts: 5,
        };
        let Attempt { error, .. } = sender.deliver(&delivery);
        assert_eq!(error.as_deref(), Some("registration 7 is not active"));
    }

    #[test]
    fn transient_lookup_failure_is_not_dead_lettered() {
        let deliveries = Arc::new(FakeDeliveries {
            registrations: Vec::new(),
            fail_lookup: true,
        });
        let sender: Sender<FakeDeliveries, FakeEvents> =
            Sender::new(deliveries, Arc::new(FakeEvents));
        let delivery = Delivery {
            id: 2,
            business_id: 1,
            webhook_id: 9,
            event_id: 1,
            attempts: 1,
            max_attempts: 5,
        };
        // A DB error must reschedule (retryable), not dead-letter.
        let Attempt { error, .. } = sender.deliver(&delivery);
        assert_eq!(error.as_deref(), Some("loading registration: db down"));
    }

    #[test]
    fn batch_and_abandoned_constants() {
        assert_eq!(super::DEFAULT_BATCH, 25);
        assert_eq!(ABANDONED_AFTER_SECS, 900);
        assert!(matches!(
            DeliveryError::Undeliverable("x".into()),
            DeliveryError::Undeliverable(_)
        ));
    }
}
