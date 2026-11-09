//! Realtime fan-out, mirroring `internal/platform/wsbus`.
//!
//! Topics are `"<scope>:<id>"` (e.g. `user:42`, `workspace:12`).
//! Authorization runs at subscribe time AND again before every event fans
//! out; results are cached per `(user, topic)` for `AUTH_TTL`.
//! The default hub is plain data passed via `AppState`, never a
//! process-global.

use std::collections::{HashMap, HashSet};
use std::sync::{Arc, Mutex, RwLock};

use serde::{Deserialize, Serialize};
use thiserror::Error;

pub const AUTH_TTL_SECS: u64 = 45;
pub const MAX_TOPICS_PER_SUBSCRIBE: usize = 64;
pub const MAX_TOPICS_PER_CLIENT: usize = 256;
pub const MAX_CONNECTIONS_PER_USER: usize = 8;
pub const MAX_INBOUND_OPS_PER_WINDOW: usize = 100;
pub const PUBLISH_CHANNEL_CAPACITY: usize = 256;
pub const BACKPLANE_CHANNEL: &str = "wsbus-events";

pub const ROUTE_DASH_WS: &str = "/dash/ws";
pub const ROUTE_FLOW_WS: &str = "/flow/ws";
pub const ROUTE_POS_WS: &str = "/pos/ws";

pub const AUTHORIZED_SCOPES: [&str; 5] = ["user", "business", "workspace", "project", "task"];

#[derive(Debug, Error, PartialEq)]
pub enum WsError {
    #[error("unknown topic scope: {0}")]
    UnknownScope(String),
    #[error("bad topic: {0}")]
    BadTopic(String),
    #[error("already subscribed to {0}")]
    DuplicateTopic(String),
    #[error("topic limit exceeded: {0}")]
    TopicLimit(String),
    #[error("per-user connection cap reached for user {0}")]
    ConnectionCap(i64),
}

#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
pub struct Event {
    pub topic: String,
    #[serde(rename = "type")]
    pub event_type: String,
    #[serde(default, rename = "entity_id", skip_serializing_if = "is_zero_i64")]
    pub entity_id: i64,
    #[serde(default, rename = "actor_id", skip_serializing_if = "is_zero_i64")]
    pub actor_id: i64,
}

fn is_zero_i64(value: &i64) -> bool {
    *value == 0
}

#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
pub struct ClientMessage {
    pub op: String,
    #[serde(default)]
    pub topics: Vec<String>,
}

#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
pub struct ServerMessage {
    pub op: String,
    #[serde(default, skip_serializing_if = "Vec::is_empty")]
    pub topics: Vec<String>,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub event: Option<Event>,
    #[serde(default, skip_serializing_if = "String::is_empty")]
    pub error: String,
}

pub fn parse_topic(topic: &str) -> Option<(String, i64)> {
    let colon = topic.find(':')?;
    if colon == 0 || colon == topic.len() - 1 {
        return None;
    }
    let id_text = &topic[colon + 1..];
    if !id_text.bytes().all(|byte| byte.is_ascii_digit()) {
        return None;
    }
    let id: i64 = id_text.parse().ok()?;
    Some((topic[..colon].to_string(), id))
}

/// The main_routes.go authorizer shape check: unknown scopes deny.
pub fn authorize_topic_scope(topic: &str) -> Result<String, WsError> {
    let (scope, _) = parse_topic(topic).ok_or_else(|| WsError::BadTopic(topic.to_string()))?;
    if AUTHORIZED_SCOPES.contains(&scope.as_str()) {
        Ok(scope)
    } else {
        Err(WsError::UnknownScope(scope))
    }
}

pub type AuthorizeFn = Arc<dyn Fn(i64, &str) -> bool + Send + Sync>;

#[derive(Debug, Clone)]
struct AuthEntry {
    allowed: bool,
    expires_secs: u64,
}

#[derive(Debug, Default)]
struct Inner {
    topics: HashMap<String, HashSet<usize>>,
    clients: HashMap<usize, Client>,
    user_cap: HashMap<i64, usize>,
    next_client: usize,
    auth_cache: HashMap<String, AuthEntry>,
}

#[derive(Debug, Clone)]
pub struct Client {
    pub id: usize,
    pub user_id: i64,
    pub topics: HashSet<String>,
    pub ops: usize,
    pub window_start_secs: u64,
}

pub struct Hub {
    authorizer: AuthorizeFn,
    auth_ttl_secs: u64,
    inner: RwLock<Inner>,
    pub publish_tx: std::sync::mpsc::SyncSender<Event>,
    publish_rx: Mutex<Option<std::sync::mpsc::Receiver<Event>>>,
}

impl Hub {
    pub fn new(authorizer: AuthorizeFn) -> Self {
        let (publish_tx, publish_rx) = std::sync::mpsc::sync_channel(PUBLISH_CHANNEL_CAPACITY);
        Self {
            authorizer,
            auth_ttl_secs: AUTH_TTL_SECS,
            inner: RwLock::new(Inner::default()),
            publish_tx,
            publish_rx: Mutex::new(Some(publish_rx)),
        }
    }

    pub fn with_auth_ttl(mut self, ttl_secs: u64) -> Self {
        self.auth_ttl_secs = ttl_secs;
        self
    }

    fn now_secs() -> u64 {
        std::time::SystemTime::now()
            .duration_since(std::time::UNIX_EPOCH)
            .map(|d| d.as_secs())
            .unwrap_or(0)
    }

    pub fn authorize(&self, user_id: i64, topic: &str) -> bool {
        let key = format!("{user_id}|{topic}");
        {
            let inner = self.inner.read().unwrap_or_else(|e| e.into_inner());
            if let Some(entry) = inner.auth_cache.get(&key) {
                if Self::now_secs() < entry.expires_secs {
                    return entry.allowed;
                }
            }
        }
        let allowed = (self.authorizer)(user_id, topic);
        let mut inner = self.inner.write().unwrap_or_else(|e| e.into_inner());
        inner.auth_cache.insert(
            key,
            AuthEntry {
                allowed,
                expires_secs: Self::now_secs() + self.auth_ttl_secs,
            },
        );
        allowed
    }

    pub fn invalidate_auth(&self, user_id: i64, topic: &str) {
        let key = format!("{user_id}|{topic}");
        self.inner
            .write()
            .unwrap_or_else(|e| e.into_inner())
            .auth_cache
            .remove(&key);
    }

    pub fn prune_auth_cache(&self) {
        let now = Self::now_secs();
        let mut inner = self.inner.write().unwrap_or_else(|e| e.into_inner());
        inner.auth_cache.retain(|_, entry| now < entry.expires_secs);
    }

    /// Non-blocking publish; drops when full (matches Go).
    pub fn publish(&self, event: Event) {
        let _ = self.publish_tx.try_send(event);
    }

    /// Drain the publish queue until empty. Returns fan-out rows.
    pub fn drain(&self) -> Vec<Event> {
        let mut out = Vec::new();
        let guard = self.publish_rx.lock().unwrap_or_else(|e| e.into_inner());
        let Some(rx) = guard.as_ref() else {
            return out;
        };
        while let Ok(event) = rx.try_recv() {
            out.push(event);
        }
        out
    }

    pub fn register(&self, user_id: i64) -> Result<usize, WsError> {
        let mut inner = self.inner.write().unwrap_or_else(|e| e.into_inner());
        let count = inner.user_cap.entry(user_id).or_default();
        if *count >= self.user_cap_count() {
            return Err(WsError::ConnectionCap(user_id));
        }
        *count += 1;
        let id = inner.next_client;
        inner.next_client += 1;
        inner.clients.insert(
            id,
            Client {
                id,
                user_id,
                topics: HashSet::new(),
                ops: 0,
                window_start_secs: Self::now_secs(),
            },
        );
        Ok(id)
    }

    fn user_cap_count(&self) -> usize {
        MAX_CONNECTIONS_PER_USER
    }

    pub fn release(&self, user_id: i64) {
        let mut inner = self.inner.write().unwrap_or_else(|e| e.into_inner());
        if let Some(count) = inner.user_cap.get_mut(&user_id) {
            *count = count.saturating_sub(1);
        }
    }

    pub fn subscribe(&self, client_id: usize, topic: &str) -> Result<(), WsError> {
        authorize_topic_scope(topic)?;
        let mut inner = self.inner.write().unwrap_or_else(|e| e.into_inner());
        let Some(client) = inner.clients.get_mut(&client_id) else {
            return Err(WsError::BadTopic(format!(
                "client {client_id} not registered"
            )));
        };
        if client.topics.contains(topic) {
            return Err(WsError::DuplicateTopic(topic.to_string()));
        }
        if client.topics.len() >= MAX_TOPICS_PER_CLIENT {
            return Err(WsError::TopicLimit(topic.to_string()));
        }
        client.topics.insert(topic.to_string());
        inner
            .topics
            .entry(topic.to_string())
            .or_default()
            .insert(client_id);
        Ok(())
    }

    pub fn unsubscribe(&self, client_id: usize, topic: &str) {
        let mut inner = self.inner.write().unwrap_or_else(|e| e.into_inner());
        if let Some(set) = inner.topics.get_mut(topic) {
            set.remove(&client_id);
            if set.is_empty() {
                inner.topics.remove(topic);
            }
        }
        if let Some(client) = inner.clients.get_mut(&client_id) {
            client.topics.remove(topic);
        }
    }

    /// Remove a client from every topic and the cap table.
    pub fn unregister(&self, client_id: usize, user_id: i64) {
        self.unsubscribe_all(client_id);
        let mut inner = self.inner.write().unwrap_or_else(|e| e.into_inner());
        inner.clients.remove(&client_id);
        if let Some(count) = inner.user_cap.get_mut(&user_id) {
            *count = count.saturating_sub(1);
        }
    }

    pub fn unsubscribe_all(&self, client_id: usize) {
        let mut inner = self.inner.write().unwrap_or_else(|e| e.into_inner());
        let topics: Vec<String> = inner
            .clients
            .get(&client_id)
            .map(|client| client.topics.iter().cloned().collect())
            .unwrap_or_default();
        for topic in topics {
            if let Some(set) = inner.topics.get_mut(&topic) {
                set.remove(&client_id);
                if set.is_empty() {
                    inner.topics.remove(&topic);
                }
            }
        }
        if let Some(client) = inner.clients.get_mut(&client_id) {
            client.topics.clear();
        }
    }

    /// Authorize + snapshot subscribers for one event. Revoked members are
    /// removed lazily during fan-out (Go parity).
    pub fn fanout(&self, event: &Event) -> Vec<(usize, String)> {
        let raw = event_json(event);
        let target: Vec<usize> = {
            let inner = self.inner.read().unwrap_or_else(|e| e.into_inner());
            inner
                .topics
                .get(&event.topic)
                .cloned()
                .unwrap_or_default()
                .into_iter()
                .collect()
        };
        let mut delivered = Vec::new();
        for client_id in target {
            let user_id = {
                let inner = self.inner.read().unwrap_or_else(|e| e.into_inner());
                inner.clients.get(&client_id).map(|client| client.user_id)
            };
            let Some(user_id) = user_id else {
                continue;
            };
            if !self.authorize(user_id, &event.topic) {
                self.unsubscribe(client_id, &event.topic);
                continue;
            }
            delivered.push((client_id, raw.clone()));
        }
        delivered
    }

    pub fn client_count(&self) -> usize {
        self.inner
            .read()
            .unwrap_or_else(|e| e.into_inner())
            .clients
            .len()
    }

    pub fn allow_inbound_op(&self, client_id: usize) -> bool {
        let mut inner = self.inner.write().unwrap_or_else(|e| e.into_inner());
        let Some(client) = inner.clients.get_mut(&client_id) else {
            return false;
        };
        let now = Self::now_secs();
        if now.saturating_sub(client.window_start_secs) >= 1 {
            client.window_start_secs = now;
            client.ops = 0;
        }
        if client.ops >= MAX_INBOUND_OPS_PER_WINDOW {
            return false;
        }
        client.ops += 1;
        true
    }

    pub fn handle_message(
        &self,
        client_id: usize,
        message: &ClientMessage,
    ) -> Result<ServerMessage, String> {
        match message.op.as_str() {
            "subscribe" => {
                if !self.allow_inbound_op(client_id) {
                    return Err("rate limited".to_string());
                }
                let mut allowed = Vec::new();
                let topics: Vec<String> = message
                    .topics
                    .iter()
                    .take(MAX_TOPICS_PER_SUBSCRIBE)
                    .cloned()
                    .collect();
                let user_id = self
                    .inner
                    .read()
                    .unwrap_or_else(|e| e.into_inner())
                    .clients
                    .get(&client_id)
                    .map(|client| client.user_id)
                    .unwrap_or(0);
                for topic in topics {
                    if self.authorize(user_id, &topic) && self.subscribe(client_id, &topic).is_ok()
                    {
                        allowed.push(topic);
                    }
                }
                Ok(ServerMessage {
                    op: "subscribed".to_string(),
                    topics: allowed,
                    event: None,
                    error: String::new(),
                })
            }
            "unsubscribe" => {
                for topic in &message.topics {
                    self.unsubscribe(client_id, topic);
                }
                Ok(ServerMessage {
                    op: "unsubscribed".to_string(),
                    topics: message.topics.clone(),
                    event: None,
                    error: String::new(),
                })
            }
            "ping" => Ok(ServerMessage {
                op: "pong".to_string(),
                topics: Vec::new(),
                event: None,
                error: String::new(),
            }),
            other => Err(format!("unknown op: {other}")),
        }
    }
}

pub fn event_json(event: &Event) -> String {
    serde_json::to_string(&ServerMessage {
        op: "event".to_string(),
        topics: Vec::new(),
        event: Some(event.clone()),
        error: String::new(),
    })
    .unwrap_or_default()
}

/// Redis backplane envelope; every replica fans out locally exactly once.
/// Go embeds `Event` in `backplaneEvent`, so the wire shape is the flat
/// `{"topic":...,"type":...}` object — never nested under an `event` key.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct BackplaneEvent {
    pub topic: String,
    #[serde(rename = "type")]
    pub event_type: String,
    #[serde(default, rename = "entity_id", skip_serializing_if = "is_zero_i64")]
    pub entity_id: i64,
    #[serde(default, rename = "actor_id", skip_serializing_if = "is_zero_i64")]
    pub actor_id: i64,
}

impl From<&Event> for BackplaneEvent {
    fn from(event: &Event) -> Self {
        Self {
            topic: event.topic.clone(),
            event_type: event.event_type.clone(),
            entity_id: event.entity_id,
            actor_id: event.actor_id,
        }
    }
}

impl From<&BackplaneEvent> for Event {
    fn from(event: &BackplaneEvent) -> Self {
        Self {
            topic: event.topic.clone(),
            event_type: event.event_type.clone(),
            entity_id: event.entity_id,
            actor_id: event.actor_id,
        }
    }
}

pub fn backplane_channel() -> &'static str {
    BACKPLANE_CHANNEL
}

pub async fn publish_backplane(
    client: &mut redis::aio::MultiplexedConnection,
    event: &Event,
) -> Result<(), WsError> {
    let payload = serde_json::to_string(&BackplaneEvent::from(event))
        .map_err(|_| WsError::BadTopic("backplane encode".to_string()))?;
    let _: () = redis::cmd("PUBLISH")
        .arg(backplane_channel())
        .arg(payload)
        .query_async(client)
        .await
        .map_err(|_| WsError::BadTopic("backplane publish".to_string()))?;
    Ok(())
}

#[cfg(test)]
mod tests {
    use super::{
        authorize_topic_scope, event_json, BackplaneEvent, Event, Hub, WsError, AUTHORIZED_SCOPES,
    };
    use std::sync::Arc;

    fn hub() -> Hub {
        Hub::new(Arc::new(|user_id, topic| {
            // Parity with main_routes.go: user:<id> is that user.
            if let Some((scope, id)) = topic.split_once(':') {
                return (scope == "user" && id == user_id.to_string()) || scope == "business";
            }
            false
        }))
    }

    fn event(topic: &str) -> Event {
        Event {
            topic: topic.to_string(),
            event_type: "thing.updated".to_string(),
            entity_id: 7,
            actor_id: 0,
        }
    }

    #[test]
    fn topic_parsing() {
        assert_eq!(
            super::parse_topic("user:42"),
            Some(("user".to_string(), 42))
        );
        assert_eq!(super::parse_topic("user:"), None);
        assert_eq!(super::parse_topic(":42"), None);
        assert_eq!(super::parse_topic("task:abc"), None);
        assert_eq!(super::parse_topic("task:9:6"), None);
        assert_eq!(authorize_topic_scope("user:42"), Ok("user".to_string()));
        assert_eq!(
            authorize_topic_scope("room:1"),
            Err(WsError::UnknownScope("room".to_string()))
        );
    }

    #[test]
    fn topic_scope_front_door() {
        // authorize_topic_scope is the cheap main_routes.go pre-check that
        // runs before membership checks (which need DB and run in Phase 3).
        assert_eq!(
            authorize_topic_scope("business:5"),
            Ok("business".to_string())
        );
        for scope in AUTHORIZED_SCOPES {
            assert!(authorize_topic_scope(&format!("{scope}:1")).is_ok());
        }
    }

    #[test]
    fn event_json_shape() {
        let raw = event_json(&event("task:88"));
        let value: serde_json::Value = serde_json::from_str(&raw).unwrap();
        assert_eq!(value["op"], "event");
        assert_eq!(value["event"]["topic"], "task:88");
        assert_eq!(value["event"]["entity_id"], 7);
    }

    #[test]
    fn register_and_subscribe_fanout() {
        let hub = hub();
        let client = hub.register(42).unwrap();
        hub.subscribe(client, "user:42").unwrap();
        let delivered = hub.fanout(&event("user:42"));
        assert_eq!(delivered.len(), 1);
        assert!(delivered[0].1.contains("user:42"));
        assert_eq!(hub.client_count(), 1);
        // Publishing routes through the queue; draining preserves order.
        hub.publish(event("user:42"));
        hub.publish(event("user:42"));
        assert_eq!(hub.drain().len(), 2);
    }

    #[test]
    fn duplicate_and_limits() {
        let hub = hub();
        let client = hub.register(1).unwrap();
        hub.subscribe(client, "business:1").unwrap();
        assert_eq!(
            hub.subscribe(client, "business:1"),
            Err(WsError::DuplicateTopic("business:1".to_string()))
        );
        assert!(hub.subscribe(client, "room:3").is_err());
    }

    #[test]
    fn backplane_envelope_shape() {
        let envelope = BackplaneEvent::from(&event("project:7"));
        let raw = serde_json::to_string(&envelope).unwrap();
        assert!(raw.contains("\"topic\":\"project:7\""));
        assert!(!raw.contains("\"event\":"));
        let back: BackplaneEvent = serde_json::from_str(&raw).unwrap();
        assert_eq!(Event::from(&back), event("project:7"));
        assert_eq!(super::backplane_channel(), "wsbus-events");
    }
}
