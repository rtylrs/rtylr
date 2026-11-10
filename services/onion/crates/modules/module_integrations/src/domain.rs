//! Domain types and sentinel errors. Mirrors `internal/domain`.

use chrono::{DateTime, Utc};
use serde_json::Value;
use thiserror::Error;

#[derive(Debug, Error, PartialEq)]
pub enum IntegrationError {
    #[error("integration not found")]
    NotFound,
    #[error("unsupported integration provider")]
    InvalidProvider,
    #[error("integration config must be a JSON object")]
    InvalidConfig,
}

/// Providers a business can connect.
pub const PROVIDERS: [&str; 1] = ["stripe"];

/// `IntegrationStatus`.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum IntegrationStatus {
    Connected,
    Disconnected,
    Error,
}

impl IntegrationStatus {
    pub fn as_str(self) -> &'static str {
        match self {
            IntegrationStatus::Connected => "connected",
            IntegrationStatus::Disconnected => "disconnected",
            IntegrationStatus::Error => "error",
        }
    }

    pub fn parse(text: &str) -> Self {
        match text {
            "connected" => IntegrationStatus::Connected,
            "error" => IntegrationStatus::Error,
            _ => IntegrationStatus::Disconnected,
        }
    }
}

/// One provider connection.
#[derive(Debug, Clone, PartialEq)]
pub struct Integration {
    pub id: i64,
    pub owner_id: i64,
    pub business_id: i64,
    pub provider: String,
    pub status: IntegrationStatus,
    pub config: Value,
    pub connected_at: Option<DateTime<Utc>>,
    pub created_at: DateTime<Utc>,
}

#[cfg(test)]
mod tests {
    use super::{IntegrationStatus, PROVIDERS};

    #[test]
    fn provider_list_and_status() {
        assert_eq!(PROVIDERS, ["stripe"]);
        assert_eq!(IntegrationStatus::Connected.as_str(), "connected");
        assert_eq!(IntegrationStatus::parse("error"), IntegrationStatus::Error);
        assert_eq!(
            IntegrationStatus::parse("nonsense"),
            IntegrationStatus::Disconnected
        );
    }
}
