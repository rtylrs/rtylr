//! Registration validation. Mirrors `registration.go`: reject what can
//! never be delivered so the tenant hears about it now rather than finding
//! a dead-lettered delivery later. Cheap URL checks only; the address
//! itself is judged at connect time by the transport's control hook.

use crate::application::transport::is_internal_host;
use thiserror::Error;

pub const MAX_WEBHOOK_URL_LENGTH: usize = 2048;

/// Events this module accepts. Currently the only supported event is
/// `order_created` (the queue is fed by the POS order-creation handler;
/// `eventbus.TypeOrderCreated`).
pub const SUPPORTED_EVENTS: [&str; 1] = ["order_created"];

#[derive(Debug, Error, PartialEq)]
pub enum RegistrationError {
    #[error("invalid webhook registration: {0}")]
    Invalid(String),
}

/// Normalise and check a tenant-supplied registration.
pub fn validate_registration(
    raw_url: &str,
    events: &[String],
) -> Result<(String, Vec<String>), RegistrationError> {
    let raw_url = raw_url.trim();
    if raw_url.is_empty() {
        return Err(RegistrationError::Invalid("url is required".into()));
    }
    if raw_url.len() > MAX_WEBHOOK_URL_LENGTH {
        return Err(RegistrationError::Invalid(format!(
            "url is longer than {MAX_WEBHOOK_URL_LENGTH} characters"
        )));
    }
    let parsed = url::Url::parse(raw_url)
        .map_err(|_| RegistrationError::Invalid("url is not a valid url".into()))?;
    let scheme = parsed.scheme();
    if scheme != "http" && scheme != "https" {
        return Err(RegistrationError::Invalid(
            "url must start with http:// or https://".into(),
        ));
    }
    let host = parsed
        .host_str()
        .ok_or_else(|| RegistrationError::Invalid("url has no host".into()))?;
    if parsed.username() != "" || parsed.password().is_some() {
        return Err(RegistrationError::Invalid(
            "url must not contain credentials".into(),
        ));
    }
    if is_internal_host(host) {
        return Err(RegistrationError::Invalid(
            "url must point to a public host".into(),
        ));
    }
    let mut cleaned: Vec<String> = Vec::with_capacity(events.len());
    for event in events {
        let event = event.trim();
        if !SUPPORTED_EVENTS.contains(&event) {
            return Err(RegistrationError::Invalid(format!(
                "unsupported event {event:?}"
            )));
        }
        if !cleaned.iter().any(|seen| seen == event) {
            cleaned.push(event.to_string());
        }
    }
    Ok((raw_url.to_string(), cleaned))
}

#[cfg(test)]
mod tests {
    use super::{validate_registration, RegistrationError};

    #[test]
    fn rejects_bad_urls() {
        assert!(matches!(
            validate_registration("", &[]),
            Err(RegistrationError::Invalid(_))
        ));
        assert!(matches!(
            validate_registration("ftp://example.com", &[]),
            Err(RegistrationError::Invalid(_))
        ));
        assert!(matches!(
            validate_registration("http://localhost:8080", &[]),
            Err(RegistrationError::Invalid(_))
        ));
        assert!(matches!(
            validate_registration("http://user:pass@example.com", &[]),
            Err(RegistrationError::Invalid(_))
        ));
        assert!(matches!(
            validate_registration("http://10.0.0.1/hook", &[]),
            Err(RegistrationError::Invalid(_))
        ));
        // IPv6 loopback / unique-local / link-local arrive bracketed.
        assert!(matches!(
            validate_registration("http://[::1]/hook", &[]),
            Err(RegistrationError::Invalid(_))
        ));
        assert!(matches!(
            validate_registration("http://[fd00::1]/hook", &[]),
            Err(RegistrationError::Invalid(_))
        ));
        assert!(matches!(
            validate_registration("http://[fe80::1]/hook", &[]),
            Err(RegistrationError::Invalid(_))
        ));
    }

    #[test]
    fn accepts_public_http_and_dedupes_events() {
        let (url, events) = validate_registration(
            "https://example.com/hook",
            &["order_created".to_string(), " order_created ".to_string()],
        )
        .unwrap();
        assert_eq!(url, "https://example.com/hook");
        assert_eq!(events, vec!["order_created".to_string()]);
        assert!(matches!(
            validate_registration("https://example.com", &["other.event".into()]),
            Err(RegistrationError::Invalid(_))
        ));
    }
}
