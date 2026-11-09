//! AI quota module. Mirrors `internal/modules/aiquota`:
//! per-tenant monthly token allowances, pool/per-person modes,
//! atomic reservations, and post-stream reconciliation.

pub mod application;
pub mod domain;
pub mod infrastructure;
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
