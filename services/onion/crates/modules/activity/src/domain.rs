//! Domain types. Mirrors `internal/domain/activity.go`.

use chrono::{DateTime, Utc};

/// One activity-log entry.
#[derive(Debug, Clone, PartialEq)]
pub struct ActivityLog {
    pub id: i64,
    pub owner_id: i64,
    pub business_id: i64,
    pub app: String,
    pub entity_type: String,
    pub entity_id: i64,
    pub action: String,
    pub description: String,
    pub user_id: i64,
    pub user_name: String,
    pub created_at: DateTime<Utc>,
}

/// A page of activity plus the total count for the same owner/app
/// filter, so the client can render pagination controls.
#[derive(Debug, Clone, Default, PartialEq)]
pub struct ActivityPage {
    pub entries: Vec<ActivityLog>,
    pub total: i64,
}
