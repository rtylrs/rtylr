//! Application layer: cursor codec, ports, and the paged service.
//! Mirrors `internal/application/service.go` and
//! `internal/platform/cursor/created_at_id.go`.

pub mod cursor;
pub mod service;

pub use service::{NotificationPage, Service, DEFAULT_PAGE_SIZE, MAX_PAGE_SIZE};
