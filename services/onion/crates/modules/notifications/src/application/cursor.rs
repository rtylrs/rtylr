//! `(created_at, id)` cursor codec. Mirrors
//! `internal/platform/cursor/created_at_id.go`: JSON payload
//! `{"created_at": RFC3339Nano, "id": int64}` base64url-nopad encoded.

use chrono::{DateTime, SecondsFormat, Utc};
use serde::{Deserialize, Serialize};
use thiserror::Error;

#[derive(Debug, Error, PartialEq)]
pub enum CursorError {
    #[error("invalid cursor")]
    Invalid,
}

#[derive(Debug, Clone, PartialEq)]
pub struct CreatedAtId {
    pub created_at: DateTime<Utc>,
    pub id: i64,
}

#[derive(Debug, Serialize, Deserialize)]
struct CreatedAtIdPayload {
    created_at: String,
    id: i64,
}

pub fn encode_created_at_id(value: &CreatedAtId) -> String {
    use base64::Engine as _;
    let payload = CreatedAtIdPayload {
        created_at: value.created_at.to_rfc3339_opts(SecondsFormat::Nanos, true),
        id: value.id,
    };
    let json = serde_json::to_vec(&payload).unwrap_or_default();
    base64::engine::general_purpose::URL_SAFE_NO_PAD.encode(json)
}

pub fn decode_created_at_id(token: &str) -> Result<CreatedAtId, CursorError> {
    use base64::Engine as _;
    if token.is_empty() {
        return Err(CursorError::Invalid);
    }
    let raw = base64::engine::general_purpose::URL_SAFE_NO_PAD
        .decode(token)
        .map_err(|_| CursorError::Invalid)?;
    let payload: CreatedAtIdPayload =
        serde_json::from_slice(&raw).map_err(|_| CursorError::Invalid)?;
    if payload.id <= 0 {
        return Err(CursorError::Invalid);
    }
    let created_at = DateTime::parse_from_rfc3339(&payload.created_at)
        .map_err(|_| CursorError::Invalid)?
        .with_timezone(&Utc);
    Ok(CreatedAtId {
        created_at,
        id: payload.id,
    })
}

#[cfg(test)]
mod tests {
    use super::{decode_created_at_id, encode_created_at_id, CreatedAtId, CursorError};
    use chrono::DateTime;

    #[test]
    fn round_trip() {
        let value = CreatedAtId {
            created_at: DateTime::from_timestamp(1_700_000_123, 123_456_789).unwrap(),
            id: 42,
        };
        let token = encode_created_at_id(&value);
        assert!(!token.contains('='));
        let decoded = decode_created_at_id(&token).unwrap();
        assert_eq!(decoded.id, 42);
        assert_eq!(
            decoded.created_at.timestamp_nanos_opt(),
            value.created_at.timestamp_nanos_opt()
        );
    }

    #[test]
    fn rejects_bad_tokens() {
        assert_eq!(
            decode_created_at_id("!!!not-base64!!!"),
            Err(CursorError::Invalid)
        );
        assert_eq!(decode_created_at_id(""), Err(CursorError::Invalid));
        // base64url of {"created_at":"...","id":0}
        assert_eq!(
            decode_created_at_id("eyJjcmVhdGVkX2F0IjoiMTk3MC0wMS0wMVQwMDowMDowMFoiLCJpZCI6MH0"),
            Err(CursorError::Invalid)
        );
    }
}
