//! Tenant re-export: the canonical `TenantScope`/`ScopeResolver` live in
//! `db::tenant` (Unit 1.2); this module re-exports them under the `auth`
//! crate so `AppState` wires one scope type.

pub use db::tenant::{ScopeError, ScopeResolver, TenantScope};
