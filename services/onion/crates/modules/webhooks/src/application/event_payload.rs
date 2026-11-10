//! Envelope encoding. Mirrors `event_payload.go` and the envelope part of
//! `sender.go`.

use crate::domain::Event;
use serde::Serialize;

/// The delivered document. IDs are strings because an int64 crosses the
/// wire as a string throughout this product, and a subscriber parsing one
/// as a JSON number would lose precision at the top of the range.
#[derive(Debug, Clone, PartialEq, Serialize)]
pub struct Envelope {
    pub id: String,
    #[serde(rename = "type")]
    pub event_type: String,
    pub subject: EnvelopeSubject,
    pub occurred_at: String,
    pub data: serde_json::Value,
}

#[derive(Debug, Clone, PartialEq, Serialize)]
pub struct EnvelopeSubject {
    #[serde(rename = "type")]
    pub subject_type: String,
    pub id: String,
}

/// `eventPayloadBytes`: []byte / json.RawMessage pass through; anything
/// else is JSON-encoded; an encode failure yields `nil` (the sender then
/// dead-letters the delivery as undeliverable).
pub fn event_payload_bytes(payload: &serde_json::Value) -> Vec<u8> {
    match payload {
        serde_json::Value::String(text) => text.as_bytes().to_vec(),
        value => serde_json::to_vec(value).unwrap_or_default(),
    }
}

pub fn encode_envelope(event: &Event) -> Result<Vec<u8>, serde_json::Error> {
    let mut data: serde_json::Value =
        serde_json::from_slice(&event.payload).unwrap_or_else(|_| serde_json::json!({}));
    if data.is_null() {
        data = serde_json::json!({});
    }
    serde_json::to_vec(&Envelope {
        id: event.id.to_string(),
        event_type: event.event_type.clone(),
        subject: EnvelopeSubject {
            subject_type: event.subject_type.clone(),
            id: event.subject_id.to_string(),
        },
        occurred_at: event
            .occurred_at
            .to_rfc3339_opts(chrono::SecondsFormat::Secs, true),
        data,
    })
}

#[cfg(test)]
mod tests {
    use super::{encode_envelope, event_payload_bytes};
    use crate::domain::Event;

    #[test]
    fn envelope_shape() {
        let event = Event {
            id: 12,
            event_type: "order_created".into(),
            subject_type: "order".into(),
            subject_id: 99,
            payload: br#"{"a":1}"#.to_vec(),
            occurred_at: chrono::DateTime::from_timestamp(0, 0).unwrap(),
        };
        let bytes = encode_envelope(&event).unwrap();
        let parsed: serde_json::Value = serde_json::from_slice(&bytes).unwrap();
        assert_eq!(parsed["id"], "12");
        assert_eq!(parsed["subject"]["id"], "99");
        assert_eq!(parsed["data"]["a"], 1);
        assert_eq!(parsed["occurred_at"], "1970-01-01T00:00:00Z");
    }

    #[test]
    fn payload_passthrough() {
        assert_eq!(
            event_payload_bytes(&serde_json::json!({"a":1})),
            br#"{"a":1}"#
        );
    }
}
