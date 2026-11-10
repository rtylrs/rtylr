//! Domain type. Mirrors `internal/domain`.

use chrono::{DateTime, Utc};

/// One in-app notification for one user.
#[derive(Debug, Clone, PartialEq)]
pub struct Notification {
    pub id: i64,
    pub owner_id: i64,
    pub business_id: i64,
    pub user_id: i64,
    pub title: String,
    pub body: String,
    pub notification_type: String,
    pub is_read: bool,
    pub link: String,
    pub created_at: DateTime<Utc>,
}
