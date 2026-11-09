//! Events foundation: event bus outbox, job registry/scheduler with MySQL
//! locks, and the email queue. Mirrors
//! `internal/platform/{eventbus,jobs,emailqueue}`.
//!
//! Lock keys, job names, schedules and outbox semantics must match Go:
//! during an ECS rolling deploy the old Go task and the new Rust task run
//! together, so both expand the same outbox idempotently.

pub mod emailqueue;
pub mod eventbus;
pub mod jobs;
pub mod locker;

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
