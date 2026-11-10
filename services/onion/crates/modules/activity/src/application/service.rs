//! Activity list service. Mirrors `internal/application/service.go`.
//! The cache layer is the runtime's; the key format is ported here.

use crate::domain::{ActivityLog, ActivityPage};
use thiserror::Error;

/// `activityCacheTTL`: one minute.
pub const ACTIVITY_CACHE_TTL_SECS: u64 = 60;

#[derive(Debug, Error, PartialEq)]
pub enum ActivityError {
    #[error("{0}")]
    Repo(String),
}

/// Port: one page plus the total for the same filter.
pub trait Repository: Send + Sync {
    fn list_activity(
        &self,
        business_id: i64,
        app: &str,
        limit: i32,
        offset: i32,
    ) -> Result<(Vec<ActivityLog>, i64), ActivityError>;
}

/// The activity service.
pub struct Service {
    repo: std::sync::Arc<dyn Repository>,
}

impl Service {
    pub fn new(repo: std::sync::Arc<dyn Repository>) -> Self {
        Service { repo }
    }

    /// `activity:<business>:<app>:<limit>:<offset>` — the cache key Go
    /// builds before appending the `|domain=v2` version suffix.
    pub fn cache_key(business_id: i64, app: &str, limit: i32, offset: i32) -> String {
        format!("activity:{business_id}:{app}:{limit}:{offset}")
    }

    /// A single page of activity for the business, optionally filtered to
    /// one app (empty = all), with the total count for that filter.
    pub fn list_activity(
        &self,
        business_id: i64,
        app: &str,
        limit: i32,
        offset: i32,
    ) -> Result<ActivityPage, ActivityError> {
        let (entries, total) = self.repo.list_activity(business_id, app, limit, offset)?;
        Ok(ActivityPage { entries, total })
    }
}

#[cfg(test)]
mod tests {
    use super::{ActivityError, Service, ACTIVITY_CACHE_TTL_SECS};
    use crate::domain::{ActivityLog, ActivityPage};
    use chrono::{DateTime, Utc};
    use std::sync::Arc;

    struct FakeRepo;

    impl super::Repository for FakeRepo {
        fn list_activity(
            &self,
            _business_id: i64,
            app: &str,
            _limit: i32,
            _offset: i32,
        ) -> Result<(Vec<ActivityLog>, i64), ActivityError> {
            if app == "boom" {
                return Err(ActivityError::Repo("db down".into()));
            }
            Ok((
                vec![ActivityLog {
                    id: 1,
                    owner_id: 1,
                    business_id: 1,
                    app: app.to_string(),
                    entity_type: "order".into(),
                    entity_id: 2,
                    action: "created".into(),
                    description: String::new(),
                    user_id: 3,
                    user_name: "Ada".into(),
                    created_at: DateTime::<Utc>::from_timestamp(0, 0).unwrap(),
                }],
                1,
            ))
        }
    }

    #[test]
    fn lists_with_total() {
        let service = Service::new(Arc::new(FakeRepo));
        let page: ActivityPage = service.list_activity(1, "pos", 20, 0).unwrap();
        assert_eq!(page.total, 1);
        assert_eq!(page.entries[0].user_name, "Ada");
        assert!(service.list_activity(1, "boom", 20, 0).is_err());
    }

    #[test]
    fn cache_key_and_ttl() {
        assert_eq!(Service::cache_key(7, "pos", 20, 40), "activity:7:pos:20:40");
        assert_eq!(ACTIVITY_CACHE_TTL_SECS, 60);
    }
}
