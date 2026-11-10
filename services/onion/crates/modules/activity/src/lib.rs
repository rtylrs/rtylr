//! Activity module: the audit-activity projection and its list view.
//! Mirrors `internal/modules/activity`.

pub mod application;
pub mod domain;
pub mod facts;
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
