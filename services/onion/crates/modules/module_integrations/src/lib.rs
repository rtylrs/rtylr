//! Integrations module: the per-business provider connections (currently
//! Stripe) with config redaction. Mirrors `internal/modules/integrations`.

pub mod application;
pub mod domain;
pub mod queries;

/// Workspace crate version, surfaced in CI and version checks.
pub const CRATE_VERSION: &str = env!("CARGO_PKG_VERSION");

#[cfg(test)]
mod tests {
    use super::CRATE_VERSION;

    #[test]
    fn crate_version_is_nonempty() {
        assert!(!CRATE_VERSION.is_empty());
    }
}
