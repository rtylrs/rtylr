//! Effective-access model, mirroring `internal/platform/authorization`.
//!
//! Persistence adapters translate DB rows into the same shapes; the
//! predicates in `roles.rs` operate on them.

use std::collections::HashMap;

use thiserror::Error;

#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash)]
pub enum AppId {
    Hub,
    Pos,
    Erp,
    Hr,
    Crm,
    Finance,
    Flow,
    Insights,
    Menu,
    Recruit,
}

impl AppId {
    pub fn as_str(&self) -> &'static str {
        match self {
            Self::Hub => "hub",
            Self::Pos => "pos",
            Self::Erp => "erp",
            Self::Hr => "hr",
            Self::Crm => "crm",
            Self::Finance => "finance",
            Self::Flow => "flow",
            Self::Insights => "insights",
            Self::Menu => "menu",
            Self::Recruit => "recruit",
        }
    }

    pub fn app_requires_scope(&self) -> bool {
        matches!(self, Self::Pos | Self::Flow)
    }
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct Scope {
    pub kind: String,
    pub id: String,
}

#[derive(Debug, Clone, PartialEq)]
pub struct EffectiveAppAccess {
    pub app_id: AppId,
    pub licensed: bool,
    pub has_access: bool,
    pub permissions: Vec<String>,
    pub permission_scopes: HashMap<String, Vec<Scope>>,
}

#[derive(Debug, Clone, PartialEq)]
pub struct EffectiveAccess {
    pub business_id: i64,
    pub dash_permissions: Vec<String>,
    pub apps: Vec<EffectiveAppAccess>,
    pub is_owner: bool,
    pub is_admin: bool,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum ProvisioningMode {
    Invitation,
    DirectCredentials,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum ProvisioningState {
    PendingInvitation,
    PendingActivation,
    Active,
    Suspended,
    Revoked,
}

#[derive(Debug, Clone, PartialEq)]
pub struct ProvisioningStatus {
    pub mode: ProvisioningMode,
    pub state: ProvisioningState,
}

#[derive(Debug, Error, PartialEq)]
pub enum ScopeError {
    #[error("invalid assignment scope")]
    InvalidScope,
    #[error("invalid assignment scope")]
    NotFound,
    #[error("assignment scope lookup failed: {0}")]
    LookupFailed(String),
    #[error("assignment scope reader is not configured")]
    ReaderNotConfigured,
}

impl ScopeError {
    pub fn invalid() -> Self {
        Self::InvalidScope
    }
}

/// Reads current committed references without joining the caller's tx.
pub trait ScopeReader: Send + Sync {
    fn scope_exists(
        &self,
        business_id: i64,
        scope_type: &str,
        scope_id: i64,
    ) -> Result<bool, ScopeError>;
}

/// `pos_location` / `flow_project` are the only permitted scope types.
pub fn validate_assignment_scope(
    reader: Option<&dyn ScopeReader>,
    business_id: i64,
    scope_type: &str,
    scope_id: &str,
) -> Result<(), ScopeError> {
    let id: i64 = scope_id.parse().map_err(|_| ScopeError::invalid())?;
    if id <= 0 || business_id <= 0 {
        return Err(ScopeError::invalid());
    }
    if scope_type != "pos_location" && scope_type != "flow_project" {
        return Err(ScopeError::invalid());
    }
    let reader = reader.ok_or(ScopeError::ReaderNotConfigured)?;
    if reader.scope_exists(business_id, scope_type, id)? {
        Ok(())
    } else {
        Err(ScopeError::invalid())
    }
}

#[cfg(test)]
mod tests {
    use super::{validate_assignment_scope, ScopeError, ScopeReader};

    struct StubReader {
        exists: bool,
    }

    impl ScopeReader for StubReader {
        fn scope_exists(
            &self,
            _business_id: i64,
            _scope_type: &str,
            _scope_id: i64,
        ) -> Result<bool, ScopeError> {
            Ok(self.exists)
        }
    }

    #[test]
    fn scope_validation_ported() {
        let reader = StubReader { exists: true };
        assert!(validate_assignment_scope(Some(&reader), 1, "pos_location", "5").is_ok());
        assert!(validate_assignment_scope(Some(&reader), 1, "flow_project", "5").is_ok());
        assert!(validate_assignment_scope(Some(&reader), 1, "other_type", "5").is_err());
        assert!(validate_assignment_scope(Some(&reader), 1, "pos_location", "0").is_err());
        assert!(validate_assignment_scope(Some(&reader), 1, "pos_location", "abc").is_err());
        assert!(validate_assignment_scope(Some(&reader), 0, "pos_location", "5").is_err());
        assert!(validate_assignment_scope(None, 1, "pos_location", "5").is_err());
        let absent = StubReader { exists: false };
        assert!(validate_assignment_scope(Some(&absent), 1, "pos_location", "5").is_err());
    }
}
