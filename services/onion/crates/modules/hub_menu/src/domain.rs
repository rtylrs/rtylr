//! Domain types. Mirrors `internal/domain/banner.go`.

use chrono::{DateTime, Utc};
use thiserror::Error;

#[derive(Debug, Error, PartialEq)]
pub enum BannerError {
    #[error("invalid menu banner input")]
    InvalidInput,
    #[error("menu banner not found")]
    NotFound,
}

/// One navigation-menu banner.
#[derive(Debug, Clone, PartialEq)]
pub struct Banner {
    pub id: i64,
    pub business_id: i64,
    pub title: String,
    pub body: String,
    pub image_url: String,
    pub link_url: String,
    pub position: i32,
    pub is_active: bool,
    pub starts_at: DateTime<Utc>,
    pub ends_at: DateTime<Utc>,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
}

/// Banner create/update payload.
#[derive(Debug, Clone, Default, PartialEq)]
pub struct BannerInput {
    pub title: String,
    pub body: String,
    pub image_url: String,
    pub link_url: String,
    pub position: i32,
    pub is_active: bool,
    pub starts_at: DateTime<Utc>,
    pub ends_at: DateTime<Utc>,
}
