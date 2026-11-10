//! Integrations service. Mirrors `internal/application/service.go`.
//! The cache layer is the runtime's; scope/key formats are ported here.

use crate::application::redaction::redact_config;
use crate::domain::{Integration, IntegrationError, IntegrationStatus, PROVIDERS};
use serde_json::Value;
use thiserror::Error;
/// `integrationsTTL`: five minutes.
pub const INTEGRATIONS_TTL_SECS: u64 = 5 * 60;
/// `maxConfigBytes`: 8 KiB.
pub const MAX_CONFIG_BYTES: usize = 8 << 10;

#[derive(Debug, Error, PartialEq)]
pub enum ServiceError {
    #[error("integration id is required")]
    MissingId,
    #[error("{0}")]
    Domain(#[from] IntegrationError),
}

/// Port: list/connect/disconnect.
pub trait Repository: Send + Sync {
    fn list(&self, business_id: i64) -> Result<Vec<Integration>, ServiceError>;
    fn connect(
        &self,
        business_id: i64,
        provider: &str,
        config: &Value,
    ) -> Result<i64, ServiceError>;
    fn disconnect(&self, business_id: i64, integration_id: i64) -> Result<(), ServiceError>;
}

/// The integrations service.
pub struct Service {
    repo: std::sync::Arc<dyn Repository>,
}

impl Service {
    pub fn new(repo: std::sync::Arc<dyn Repository>) -> Self {
        Service { repo }
    }

    /// Cache key for the list view (Go appends `|domain=v2`).
    pub fn list_cache_key(business_id: i64) -> String {
        format!("integrations:{business_id}:list")
    }

    /// Scope prefix invalidated on every mutation.
    pub fn invalidate_scope(business_id: i64) -> String {
        format!("integrations:{business_id}")
    }

    pub fn list(&self, business_id: i64) -> Result<Vec<Integration>, ServiceError> {
        self.repo.list(business_id)
    }

    /// `Connect`: provider must be in `PROVIDERS`, config must be a JSON
    /// object within `MAX_CONFIG_BYTES` (empty becomes `{}`).
    pub fn connect(
        &self,
        business_id: i64,
        provider: &str,
        config: &Value,
    ) -> Result<i64, ServiceError> {
        let provider = provider.trim().to_ascii_lowercase();
        if !PROVIDERS.contains(&provider.as_str()) {
            return Err(IntegrationError::InvalidProvider.into());
        }
        let mut config = config.clone();
        if config.is_null() {
            config = serde_json::json!({});
        }
        let encoded = serde_json::to_vec(&config).unwrap_or_default();
        if encoded.len() > MAX_CONFIG_BYTES {
            return Err(IntegrationError::InvalidConfig.into());
        }
        if !config.is_object() {
            return Err(IntegrationError::InvalidConfig.into());
        }
        self.repo.connect(business_id, &provider, &config)
    }

    pub fn disconnect(&self, business_id: i64, integration_id: i64) -> Result<(), ServiceError> {
        if integration_id == 0 {
            return Err(ServiceError::MissingId);
        }
        self.repo.disconnect(business_id, integration_id)
    }
}

/// The list view the API returns: configs are redacted before they leave
/// the module.
pub fn redacted_list(integrations: &[Integration]) -> Vec<Value> {
    integrations
        .iter()
        .map(|integration| {
            serde_json::json!({
                "id": integration.id,
                "owner_id": integration.owner_id,
                "business_id": integration.business_id,
                "provider": integration.provider,
                "status": integration.status.as_str(),
                "config": redact_config(&integration.config),
                "connected_at": integration.connected_at,
                "created_at": integration.created_at,
            })
        })
        .collect()
}

/// Parse a stored status string.
pub fn parse_status(text: &str) -> IntegrationStatus {
    IntegrationStatus::parse(text)
}

#[cfg(test)]
mod tests {
    use super::{redacted_list, Service, ServiceError, INTEGRATIONS_TTL_SECS, MAX_CONFIG_BYTES};
    use crate::domain::{Integration, IntegrationStatus};
    use chrono::{DateTime, Utc};
    use serde_json::Value;
    use std::sync::Arc;

    struct FakeRepo;

    impl super::Repository for FakeRepo {
        fn list(&self, _business_id: i64) -> Result<Vec<Integration>, ServiceError> {
            Ok(vec![Integration {
                id: 1,
                owner_id: 1,
                business_id: 1,
                provider: "stripe".into(),
                status: IntegrationStatus::Connected,
                config: serde_json::json!({"secret_key": "sk", "business_id": 1}),
                connected_at: Some(DateTime::<Utc>::from_timestamp(0, 0).unwrap()),
                created_at: DateTime::<Utc>::from_timestamp(0, 0).unwrap(),
            }])
        }
        fn connect(
            &self,
            _business_id: i64,
            provider: &str,
            _config: &Value,
        ) -> Result<i64, ServiceError> {
            assert_eq!(provider, "stripe");
            Ok(9)
        }
        fn disconnect(&self, _business_id: i64, _integration_id: i64) -> Result<(), ServiceError> {
            Ok(())
        }
    }

    #[test]
    fn connect_validation() {
        let service = Service::new(Arc::new(FakeRepo));
        assert!(matches!(
            service.connect(1, "paypal", &serde_json::json!({})),
            Err(ServiceError::Domain(
                crate::domain::IntegrationError::InvalidProvider
            ))
        ));
        assert!(matches!(
            service.connect(1, "stripe", &serde_json::json!("nope")),
            Err(ServiceError::Domain(
                crate::domain::IntegrationError::InvalidConfig
            ))
        ));
        // Empty config becomes {}.
        assert_eq!(service.connect(1, " stripe ", &Value::Null).unwrap(), 9);
        // Oversized config rejected.
        let big = serde_json::json!({"blob": "x".repeat(MAX_CONFIG_BYTES)});
        assert!(matches!(
            service.connect(1, "stripe", &big),
            Err(ServiceError::Domain(
                crate::domain::IntegrationError::InvalidConfig
            ))
        ));
    }

    #[test]
    fn disconnect_requires_id() {
        let service = Service::new(Arc::new(FakeRepo));
        assert_eq!(
            service.disconnect(1, 0).unwrap_err(),
            ServiceError::MissingId
        );
    }

    #[test]
    fn list_redacts_secrets() {
        let service = Service::new(Arc::new(FakeRepo));
        let list = service.list(1).unwrap();
        let redacted = redacted_list(&list);
        assert_eq!(redacted[0]["config"]["secret_key"], "[redacted]");
        assert_eq!(redacted[0]["config"]["business_id"], 1);
        assert_eq!(INTEGRATIONS_TTL_SECS, 300);
    }
}
