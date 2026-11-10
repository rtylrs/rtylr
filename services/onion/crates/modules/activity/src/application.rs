//! Application layer. Mirrors `internal/application/service.go`.

pub mod service;

pub use service::{Service, ACTIVITY_CACHE_TTL_SECS};
