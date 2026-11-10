//! Application layer: list/connect/disconnect and config redaction.
//! Mirrors `internal/application/service.go`.

pub mod redaction;
pub mod service;

pub use redaction::{is_secret_key, redact_config};
pub use service::{Service, INTEGRATIONS_TTL_SECS, MAX_CONFIG_BYTES};
