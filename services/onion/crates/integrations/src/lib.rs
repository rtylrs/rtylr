//! External clients: S3/CDN, Resend, Slack, Paddle, Trino, Sheets,
//! OpenAI/Gemini. Mirrors
//! `internal/platform/{storage,cdn,email,paddle,slack,sheets,snips}`.
//!
//! All HTTP goes through `reqwest`; failures surface as typed errors
//! (never panics). Network calls are never made by unit tests.

pub mod ai;
pub mod cdn;
pub mod email;
pub mod paddle;
pub mod sheets;
pub mod slack;
pub mod storage;
pub mod trino;

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
