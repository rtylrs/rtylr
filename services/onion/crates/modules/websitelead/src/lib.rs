//! Website lead module: voxire.com contact/quote leads and newsletter
//! signups. Mirrors `internal/modules/websitelead`.

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
