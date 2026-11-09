//! Event bus outbox, mirroring `internal/platform/eventbus`.

use std::future::Future;
use std::pin::Pin;

use serde_json::{json, Value};
use thiserror::Error;

pub const JOB_NAME: &str = "eventbus.dispatch_events";
pub const DEFAULT_SCHEDULE: &str = "* * * * *";
pub const DEFAULT_BATCH_SIZE: i32 = 100;
pub const TYPE_ORDER_CREATED: &str = "order_created";
pub const SUBJECT_ORDER: &str = "order";

#[derive(Debug, Error, PartialEq)]
pub enum EventError {
    #[error("eventbus: incomplete fact {0:?}")]
    Incomplete(String),
    #[error("eventbus: encode {0:?}: {1}")]
    Encode(String, String),
    #[error("eventbus: dedupe key exceeds 191 bytes")]
    DedupeTooLong,
    #[error("eventbus: invalid fact JSON")]
    InvalidJson,
    #[error("eventbus: {0}")]
    Other(String),
}

#[derive(Debug, Clone, PartialEq)]
pub struct Event {
    pub id: u64,
    pub business_id: i64,
    pub event_type: String,
    pub subject_type: String,
    pub subject_id: i64,
    pub occurred_at: chrono::DateTime<chrono::Utc>,
    pub payload: Value,
    pub dedupe_key: String,
}

/// Validated envelope ready to insert (Go `EncodedEvent`).
#[derive(Debug, Clone, PartialEq)]
pub struct EncodedEvent {
    pub business_id: i64,
    pub event_type: String,
    pub subject_type: String,
    pub subject_id: i64,
    pub occurred_at: chrono::DateTime<chrono::Utc>,
    pub payload: Value,
    pub dedupe_key: String,
}

pub fn payload_json(value: Option<Value>) -> Result<Value, EventError> {
    match value {
        None => Ok(json!({})),
        Some(text) if text.is_string() => {
            let text = text.as_str().unwrap_or_default();
            let parsed: Value = serde_json::from_str(text).map_err(|_| EventError::InvalidJson)?;
            Ok(parsed)
        }
        Some(value) => {
            if value.is_null() {
                Ok(json!({}))
            } else {
                Ok(value)
            }
        }
    }
}

/// Validate the envelope and default the dedupe key
/// (`{type}:{subject_type}:{subject_id}`).
///
/// Source JSON lives in `internal/platform/eventbus/sql/queries/`; the
/// inline constants below are the same statements for sqlx call sites.
pub fn encode(
    business_id: i64,
    event_type: &str,
    subject_type: &str,
    subject_id: i64,
    occurred_at: chrono::DateTime<chrono::Utc>,
    payload: Option<Value>,
    dedupe_key: &str,
) -> Result<EncodedEvent, EventError> {
    if business_id <= 0 || event_type.is_empty() || subject_type.is_empty() || subject_id <= 0 {
        return Err(EventError::Incomplete(event_type.to_string()));
    }
    let payload = payload_json(payload).map_err(|error| match error {
        EventError::InvalidJson => {
            EventError::Encode(event_type.to_string(), "invalid fact JSON".to_string())
        }
        other => other,
    })?;
    let key = if dedupe_key.is_empty() {
        format!("{event_type}:{subject_type}:{subject_id}")
    } else {
        dedupe_key.to_string()
    };
    if key.len() > 191 {
        return Err(EventError::DedupeTooLong);
    }
    Ok(EncodedEvent {
        business_id,
        event_type: event_type.to_string(),
        subject_type: subject_type.to_string(),
        subject_id,
        occurred_at,
        payload,
        dedupe_key: key,
    })
}

pub const INSERT_EVENT_SQL: &str = r#"INSERT INTO platform_events
    (business_id, event_type, subject_type, subject_id, payload, dedupe_key, occurred_at)
VALUES (?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE id = id"#;

pub const LIST_UNDISPATCHED_SQL: &str = r#"SELECT id, business_id, event_type, subject_type, subject_id, payload, occurred_at
FROM platform_events
WHERE dispatched_at IS NULL
ORDER BY id ASC
LIMIT ?"#;

pub const MARK_DISPATCHED_SQL: &str = r#"UPDATE platform_events
SET dispatched_at = NOW()
WHERE id = ? AND dispatched_at IS NULL"#;

/// A fact consumer; writes are idempotent so partial fan-out can retry.
pub trait EventSubscriber: Send + Sync {
    fn dispatch_event<'a>(
        &'a self,
        event: &'a Event,
    ) -> Pin<Box<dyn Future<Output = Result<(), EventError>> + Send + 'a>>;
}

/// Committed reads + local acknowledgement for one producer outbox.
pub trait FactSource: Send + Sync {
    fn pending<'a>(
        &'a self,
        limit: i32,
    ) -> Pin<Box<dyn Future<Output = Result<Vec<Event>, EventError>> + Send + 'a>>;
    fn mark_relayed<'a>(
        &'a self,
        id: u64,
    ) -> Pin<Box<dyn Future<Output = Result<(), EventError>> + Send + 'a>>;
}

/// The outbox store the dispatcher fans out from.
pub trait EventStore: Send + Sync {
    fn list_undispatched<'a>(
        &'a self,
        limit: i32,
    ) -> Pin<Box<dyn Future<Output = Result<Vec<Event>, EventError>> + Send + 'a>>;
    fn mark_dispatched<'a>(
        &'a self,
        id: u64,
    ) -> Pin<Box<dyn Future<Output = Result<(), EventError>> + Send + 'a>>;
    fn insert_encoded<'a>(
        &'a self,
        event: &'a EncodedEvent,
    ) -> Pin<Box<dyn Future<Output = Result<(), EventError>> + Send + 'a>>;
}

/// Publish/archive one fact (Go `Publisher.Record` equivalent).
pub struct RecordFacts<'a> {
    pub business_id: i64,
    pub event_type: &'a str,
    pub subject_type: &'a str,
    pub subject_id: i64,
    pub occurred_at: chrono::DateTime<chrono::Utc>,
    pub payload: Option<Value>,
    pub dedupe_key: &'a str,
}

pub async fn record_event(
    store: &dyn EventStore,
    facts: RecordFacts<'_>,
) -> Result<(), EventError> {
    let encoded = encode(
        facts.business_id,
        facts.event_type,
        facts.subject_type,
        facts.subject_id,
        facts.occurred_at,
        facts.payload,
        facts.dedupe_key,
    )?;
    store.insert_encoded(&encoded).await
}

/// Fan out undispatched facts to subscribers in order.
///
/// A subscriber failure logs and skips the event (it stays undispatched and
/// retries next pass); it never aborts the batch. Relay errors are
/// collected and returned alongside the dispatched count.
pub struct Dispatcher<S, T> {
    store: S,
    sources: Vec<T>,
    batch_size: i32,
    subscribers: Vec<Box<dyn EventSubscriber>>,
}

impl<S, T> Dispatcher<S, T>
where
    S: EventStore,
    T: FactSource,
{
    pub fn new(store: S, subscribers: Vec<Box<dyn EventSubscriber>>) -> Self {
        Self {
            store,
            sources: Vec::new(),
            batch_size: DEFAULT_BATCH_SIZE,
            subscribers,
        }
    }

    pub fn with_batch_size(mut self, batch_size: i32) -> Self {
        self.batch_size = batch_size;
        self
    }

    pub fn with_sources(mut self, sources: Vec<T>) -> Self {
        self.sources = sources;
        self
    }

    pub async fn relay(&self) -> Result<(), EventError> {
        let mut failures: Vec<String> = Vec::new();
        for source in &self.sources {
            let facts = match source.pending(self.batch_size).await {
                Ok(facts) => facts,
                Err(error) => {
                    failures.push(format!("read producer facts: {error}"));
                    continue;
                }
            };
            for fact in &facts {
                let encoded = match encode(
                    fact.business_id,
                    &fact.event_type,
                    &fact.subject_type,
                    fact.subject_id,
                    fact.occurred_at,
                    Some(fact.payload.clone()),
                    &fact.dedupe_key,
                ) {
                    Ok(encoded) => encoded,
                    Err(error) => {
                        failures.push(format!("archive producer fact: {error}"));
                        break;
                    }
                };
                if let Err(error) = self.store.insert_encoded(&encoded).await {
                    failures.push(format!("archive producer fact: {error}"));
                    break;
                }
                if let Err(error) = source.mark_relayed(fact.id).await {
                    failures.push(format!("acknowledge producer fact: {error}"));
                    break;
                }
            }
        }
        if failures.is_empty() {
            Ok(())
        } else {
            Err(EventError::Other(failures.join("; ")))
        }
    }

    pub async fn dispatch_batch(&self) -> Result<i64, EventError> {
        let relay_err = self.relay().await.err();
        let events = self.store.list_undispatched(self.batch_size).await?;
        let mut dispatched = 0_i64;
        for event in &events {
            if self.dispatch_one(event).await.is_ok() {
                dispatched += 1;
            }
        }
        // Mirror Go: relay failures surface as the batch error (joined
        // with any listing error) while dispatched events still count.
        if let Some(error) = relay_err {
            return Err(error);
        }
        Ok(dispatched)
    }

    async fn dispatch_one(&self, event: &Event) -> Result<(), EventError> {
        for subscriber in &self.subscribers {
            subscriber.dispatch_event(event).await?;
        }
        self.store.mark_dispatched(event.id).await
    }
}

#[cfg(test)]
mod tests {
    use super::{
        encode, record_event, Dispatcher, Event, EventError, EventStore, EventSubscriber,
        FactSource,
    };
    use chrono::TimeZone as _;
    use std::future::Future;
    use std::pin::Pin;
    use std::sync::Mutex;

    fn now() -> chrono::DateTime<chrono::Utc> {
        chrono::Utc.timestamp_opt(1_700_000_000, 0).unwrap()
    }

    #[test]
    fn encode_defaults_and_validates() {
        let event = encode(19, "order_created", "order", 202, now(), None, "").unwrap();
        assert_eq!(event.dedupe_key, "order_created:order:202");
        assert_eq!(event.payload, serde_json::json!({}));
        assert!(encode(0, "t", "s", 1, now(), None, "").is_err());
        assert!(encode(1, "", "s", 1, now(), None, "").is_err());
        assert!(encode(1, "t", "s", 0, now(), None, "").is_err());
        assert!(encode(1, "t", "s", 1, now(), None, &"k".repeat(192)).is_err());
        assert!(encode(
            1,
            "t",
            "s",
            1,
            now(),
            Some(serde_json::json!("not json{")),
            ""
        )
        .is_err());
        let ok = encode(
            1,
            "t",
            "s",
            1,
            now(),
            Some(serde_json::json!("{\"k\":1}")),
            "",
        )
        .unwrap();
        assert_eq!(ok.payload, serde_json::json!({"k": 1}));
    }

    struct MemoryStore {
        pending: Vec<Event>,
        marks: Mutex<Vec<u64>>,
    }

    impl EventStore for MemoryStore {
        fn list_undispatched<'a>(
            &'a self,
            _limit: i32,
        ) -> Pin<Box<dyn Future<Output = Result<Vec<Event>, EventError>> + Send + 'a>> {
            Box::pin(async move { Ok(self.pending.clone()) })
        }

        fn mark_dispatched<'a>(
            &'a self,
            id: u64,
        ) -> Pin<Box<dyn Future<Output = Result<(), EventError>> + Send + 'a>> {
            Box::pin(async move {
                self.marks.lock().unwrap().push(id);
                Ok(())
            })
        }

        fn insert_encoded<'a>(
            &'a self,
            _event: &'a super::EncodedEvent,
        ) -> Pin<Box<dyn Future<Output = Result<(), EventError>> + Send + 'a>> {
            Box::pin(async move { Ok(()) })
        }
    }

    struct EmptySource;

    impl FactSource for EmptySource {
        fn pending<'a>(
            &'a self,
            _limit: i32,
        ) -> Pin<Box<dyn Future<Output = Result<Vec<Event>, EventError>> + Send + 'a>> {
            Box::pin(async move { Ok(vec![]) })
        }

        fn mark_relayed<'a>(
            &'a self,
            _id: u64,
        ) -> Pin<Box<dyn Future<Output = Result<(), EventError>> + Send + 'a>> {
            Box::pin(async move { Ok(()) })
        }
    }

    struct Recorder {
        events: Mutex<Vec<Event>>,
        fail: bool,
    }

    impl EventSubscriber for Recorder {
        fn dispatch_event<'a>(
            &'a self,
            event: &'a Event,
        ) -> Pin<Box<dyn Future<Output = Result<(), EventError>> + Send + 'a>> {
            Box::pin(async move {
                self.events.lock().unwrap().push(event.clone());
                if self.fail {
                    Err(EventError::Other("consumer unavailable".to_string()))
                } else {
                    Ok(())
                }
            })
        }
    }

    fn fact() -> Event {
        Event {
            id: 71,
            business_id: 19,
            event_type: "order_created".to_string(),
            subject_type: "order".to_string(),
            subject_id: 202,
            occurred_at: now(),
            payload: serde_json::json!({"order_id": 202}),
            dedupe_key: String::new(),
        }
    }

    #[tokio::test]
    async fn distributes_before_marking() {
        let store = MemoryStore {
            pending: vec![fact()],
            marks: Mutex::new(vec![]),
        };
        let dispatcher = Dispatcher::new(
            store,
            vec![Box::new(Recorder {
                events: Mutex::new(vec![]),
                fail: false,
            })],
        )
        .with_sources(vec![EmptySource]);
        assert_eq!(dispatcher.dispatch_batch().await.unwrap(), 1);
        assert_eq!(dispatcher.store.marks.lock().unwrap().len(), 1);
    }

    #[tokio::test]
    async fn failure_leaves_undispatched() {
        let store = MemoryStore {
            pending: vec![fact()],
            marks: Mutex::new(vec![]),
        };
        let dispatcher = Dispatcher::new(
            store,
            vec![Box::new(Recorder {
                events: Mutex::new(vec![]),
                fail: true,
            })],
        )
        .with_sources(vec![EmptySource]);
        assert_eq!(dispatcher.dispatch_batch().await.unwrap(), 0);
        assert!(dispatcher.store.marks.lock().unwrap().is_empty());
    }

    #[tokio::test]
    async fn record_event_inserts() {
        struct InsertStore;
        impl EventStore for InsertStore {
            fn list_undispatched<'a>(
                &'a self,
                _limit: i32,
            ) -> Pin<Box<dyn Future<Output = Result<Vec<Event>, EventError>> + Send + 'a>>
            {
                Box::pin(async move { Ok(vec![]) })
            }
            fn mark_dispatched<'a>(
                &'a self,
                _id: u64,
            ) -> Pin<Box<dyn Future<Output = Result<(), EventError>> + Send + 'a>> {
                Box::pin(async move { Ok(()) })
            }
            fn insert_encoded<'a>(
                &'a self,
                event: &'a super::EncodedEvent,
            ) -> Pin<Box<dyn Future<Output = Result<(), EventError>> + Send + 'a>> {
                Box::pin(async move {
                    assert_eq!(event.dedupe_key, "order_created:order:202");
                    Ok(())
                })
            }
        }
        record_event(
            &InsertStore,
            super::RecordFacts {
                business_id: 19,
                event_type: "order_created",
                subject_type: "order",
                subject_id: 202,
                occurred_at: now(),
                payload: None,
                dedupe_key: "",
            },
        )
        .await
        .unwrap();
    }
}
