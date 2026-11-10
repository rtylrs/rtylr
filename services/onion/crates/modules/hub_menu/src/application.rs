//! Application layer: banner validation and menu-item update rules.
//! Mirrors `internal/application/service.go`.

pub mod service;

pub use service::{
    normalize_banner_input, Service, MAX_BANNER_BODY_LENGTH, MAX_BANNER_TITLE_LENGTH,
    MAX_BANNER_URL_LENGTH,
};
