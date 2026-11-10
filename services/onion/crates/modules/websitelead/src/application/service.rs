//! Lead/subscriber service shell. Mirrors the validation and structure of
//! `internal/application/service.go` and `newsletter.go`; the outbound
//! email/sheets seams are runtime wiring.

use crate::domain::{
    is_allowed_language, is_allowed_newsletter_source, LeadInput, SubscriberInput,
};
use thiserror::Error;

/// Recipient lists for the two internal notifications, plus the Resend
/// audience id the worker mirrors subscribers into (empty disables it).
#[derive(Debug, Clone, Default, PartialEq)]
pub struct Notify {
    pub lead_emails: Vec<String>,
    pub newsletter_emails: Vec<String>,
    pub audience_id: String,
}

/// Detached notification timeout (`asyncLeadTaskTimeout`).
pub const ASYNC_LEAD_TASK_TIMEOUT_SECS: u64 = 10;

#[derive(Debug, Error, PartialEq)]
pub enum LeadError {
    #[error("name is required")]
    MissingName,
    #[error("email is required")]
    MissingEmail,
    #[error("invalid email address")]
    InvalidEmail,
    #[error("service is required")]
    MissingService,
    #[error("{0}")]
    Repo(String),
}

#[derive(Debug, Error, PartialEq)]
pub enum SubscriberError {
    #[error("email is required")]
    MissingEmail,
    #[error("invalid email address")]
    InvalidEmail,
    #[error("invalid newsletter source")]
    InvalidSource,
    #[error("invalid language")]
    InvalidLanguage,
    #[error("{0}")]
    Repo(String),
}

/// Port: lead + subscriber persistence.
pub trait Repository: Send + Sync {
    fn create_lead(&self, input: &LeadInput) -> Result<i64, LeadError>;
    fn create_subscriber(&self, input: &SubscriberInput) -> Result<i64, SubscriberError>;
}

/// The website-lead service.
pub struct Service {
    repo: std::sync::Arc<dyn Repository>,
    pub notify: Notify,
}

impl Service {
    pub fn new(repo: std::sync::Arc<dyn Repository>, notify: Notify) -> Self {
        Service { repo, notify }
    }

    /// `Submit`: validate, persist, then the caller fires the detached
    /// notifications (email/auto-reply/sheets) with a 10s budget.
    pub fn submit(&self, input: &LeadInput) -> Result<i64, LeadError> {
        validate_lead(input)?;
        self.repo.create_lead(input)
    }

    /// `Subscribe`: validate, persist, then the caller fires the
    /// subscriber notification detached.
    pub fn subscribe(&self, input: &SubscriberInput) -> Result<i64, SubscriberError> {
        validate_subscriber(input)?;
        self.repo.create_subscriber(input)
    }
}

fn validate_lead(input: &LeadInput) -> Result<(), LeadError> {
    if input.name.trim().is_empty() {
        return Err(LeadError::MissingName);
    }
    if input.email.trim().is_empty() {
        return Err(LeadError::MissingEmail);
    }
    if !is_plausible_email(&input.email) {
        return Err(LeadError::InvalidEmail);
    }
    if input.service.trim().is_empty() {
        return Err(LeadError::MissingService);
    }
    Ok(())
}

fn validate_subscriber(input: &SubscriberInput) -> Result<(), SubscriberError> {
    if input.email.trim().is_empty() {
        return Err(SubscriberError::MissingEmail);
    }
    if !is_plausible_email(&input.email) {
        return Err(SubscriberError::InvalidEmail);
    }
    if !is_allowed_newsletter_source(input.source.trim()) {
        return Err(SubscriberError::InvalidSource);
    }
    if !is_allowed_language(input.language.trim()) {
        return Err(SubscriberError::InvalidLanguage);
    }
    Ok(())
}

/// Minimal RFC-ish shape check: `local@domain.tld`, no spaces.
fn is_plausible_email(email: &str) -> bool {
    let email = email.trim();
    if email.contains(' ') || email.matches('@').count() != 1 {
        return false;
    }
    let Some((local, domain)) = email.split_once('@') else {
        return false;
    };
    if local.is_empty() || domain.is_empty() {
        return false;
    }
    let Some((host, tld)) = domain.rsplit_once('.') else {
        return false;
    };
    !host.is_empty() && !tld.is_empty()
}

#[cfg(test)]
mod tests {
    use super::{
        is_plausible_email, validate_lead, validate_subscriber, LeadError, LeadInput,
        SubscriberError, SubscriberInput, ASYNC_LEAD_TASK_TIMEOUT_SECS,
    };

    fn lead() -> LeadInput {
        LeadInput {
            name: "Ada".into(),
            email: "ada@example.com".into(),
            service: "pos".into(),
            ..Default::default()
        }
    }

    fn subscriber() -> SubscriberInput {
        SubscriberInput {
            email: "ada@example.com".into(),
            source: "intel_drop".into(),
            language: "en".into(),
            ..Default::default()
        }
    }

    #[test]
    fn lead_validation() {
        assert_eq!(
            validate_lead(&LeadInput::default()),
            Err(LeadError::MissingName)
        );
        let mut bad = lead();
        bad.name = "  ".to_string();
        assert_eq!(validate_lead(&bad), Err(LeadError::MissingName));
        let mut bad = lead();
        bad.email = String::new();
        assert_eq!(validate_lead(&bad), Err(LeadError::MissingEmail));
        let mut bad = lead();
        bad.email = "not-an-email".to_string();
        assert_eq!(validate_lead(&bad), Err(LeadError::InvalidEmail));
        let mut bad = lead();
        bad.service = String::new();
        assert_eq!(validate_lead(&bad), Err(LeadError::MissingService));
        assert!(validate_lead(&lead()).is_ok());
    }

    #[test]
    fn subscriber_validation() {
        assert!(validate_subscriber(&subscriber()).is_ok());
        let mut bad = subscriber();
        bad.source = "footer".to_string();
        assert_eq!(
            validate_subscriber(&bad),
            Err(SubscriberError::InvalidSource)
        );
        let mut bad = subscriber();
        bad.language = "fr".to_string();
        assert_eq!(
            validate_subscriber(&bad),
            Err(SubscriberError::InvalidLanguage)
        );
        let mut bad = subscriber();
        bad.email = "nope".to_string();
        assert_eq!(
            validate_subscriber(&bad),
            Err(SubscriberError::InvalidEmail)
        );
    }

    #[test]
    fn email_shape() {
        assert!(is_plausible_email("a@b.co"));
        assert!(is_plausible_email(" first.last+tag@sub.example.com "));
        assert!(!is_plausible_email("a@b"));
        assert!(!is_plausible_email("a b@example.com"));
        assert!(!is_plausible_email("a@@example.com"));
        assert_eq!(ASYNC_LEAD_TASK_TIMEOUT_SECS, 10);
    }
}
