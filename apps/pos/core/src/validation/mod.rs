//! Input-validation primitives shared across Tauri commands.
//!
//! Every Tauri command should funnel untrusted input through this module
//! before reaching the hardware or transport layers. Keep validators pure
//! and easily property-testable.

pub mod drawer;
pub mod kds;
pub mod money;
pub mod net;
pub mod receipt;
pub mod text;
