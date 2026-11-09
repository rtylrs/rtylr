//! HTTP server platform: middleware-chain order, rate limiting, Partials
//! X-API-KEY bypass, health/version payloads, and the error writer.
//! Mirrors `internal/platform/server/http` (the unit-1.8 scope in
//! `docs/rust-migration/STATE.md`). The axum wiring lives in `bins/api`
//! (unit 1.9); this crate ports the platform rules as pure decisions.

pub mod chain;
pub mod cors;
pub mod errorwriter;
pub mod headers;
pub mod limits;
pub mod partials;
pub mod ratelimit;
pub mod version;

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
