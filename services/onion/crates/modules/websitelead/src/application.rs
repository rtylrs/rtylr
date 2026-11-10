//! Application layer: budget tiering, validation, and the service shell.
//! Mirrors `internal/application/{budget,service,newsletter}.go`. The
//! outbound email/sheets seams are runtime wiring; the pure rules are
//! ported here.

pub mod budget;
pub mod service;

pub use budget::{budget_tier, TIER_HOT, TIER_NEW, TIER_WARM};
pub use service::{Notify, Service};
