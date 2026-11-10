//! Paged notifications service. Mirrors `internal/application/service.go`.
//! The cache layer is the runtime's (the key builder is ported here so
//! the wiring matches Go's `notificationsKey`).

use crate::application::cursor::{decode_created_at_id, encode_created_at_id, CreatedAtId};
use crate::domain::Notification;
use chrono::{DateTime, Utc};
use thiserror::Error;

pub const DEFAULT_PAGE_SIZE: i32 = 100;
pub const MAX_PAGE_SIZE: i32 = 100;

/// Notifications are per-user and hot-read; piggyback on UserPrefix so
/// `/invalidate-cache?type=user` covers them too.
pub const NOTIFICATIONS_TTL_SECS: u64 = 30;

#[derive(Debug, Error, PartialEq)]
pub enum NotificationError {
    #[error("notification owner, business, and recipient are required")]
    MissingRecipients,
    #[error("notification title and type are required")]
    MissingTitleOrType,
    #[error("invalid cursor")]
    InvalidCursor,
    #[error("{0}")]
    Repo(String),
}

/// Port: full list (fallback when the repository has no page query).
pub trait Repository: Send + Sync {
    fn create(&self, notification: &Notification) -> Result<(), NotificationError>;
    fn list(&self, business_id: i64, user_id: i64) -> Result<Vec<Notification>, NotificationError>;
    fn mark_read(
        &self,
        business_id: i64,
        user_id: i64,
        notification_id: i64,
    ) -> Result<(), NotificationError>;
    fn mark_all_read(&self, business_id: i64, user_id: i64) -> Result<(), NotificationError>;
}

/// Port: keyset page query (`limit + 1` to detect a next page).
pub trait PagedRepository: Send + Sync {
    fn list_page(
        &self,
        business_id: i64,
        user_id: i64,
        limit: i32,
        cursor_created_at: Option<DateTime<Utc>>,
        cursor_id: Option<i64>,
    ) -> Result<Vec<Notification>, NotificationError>;
}

#[derive(Debug, Clone, Default, PartialEq)]
pub struct NotificationPage {
    pub notifications: Vec<Notification>,
    pub next_cursor: String,
}

/// The notifications service.
pub struct Service {
    // Trait objects: the runtime picks the paged repository when the
    // concrete type implements it, mirroring Go's type assertion.
    repo: std::sync::Arc<dyn Repository>,
    paged: Option<std::sync::Arc<dyn PagedRepository>>,
}

impl Service {
    pub fn new(
        repo: std::sync::Arc<dyn Repository>,
        paged: Option<std::sync::Arc<dyn PagedRepository>>,
    ) -> Self {
        Service { repo, paged }
    }

    /// `notificationsKey`: the cache key for one page.
    pub fn notifications_key(
        business_id: i64,
        user_id: i64,
        limit: i32,
        cursor_id: Option<i64>,
    ) -> String {
        let cursor = cursor_id.map(|id| id.to_string()).unwrap_or_default();
        format!("notifications:{business_id}:user:{user_id}|limit:{limit}|cursor:{cursor}")
    }

    /// `invalidateNotificationsScope`: the scope prefixes invalidated on
    /// every mutation.
    pub fn invalidate_scope(business_id: i64, user_id: i64) -> Vec<String> {
        vec![
            format!("user:{user_id}"),
            format!("notifications:{business_id}:user:{user_id}*"),
        ]
    }

    pub fn list(
        &self,
        business_id: i64,
        user_id: i64,
    ) -> Result<Vec<Notification>, NotificationError> {
        self.list_page(business_id, user_id, 0, "")
            .map(|page| page.notifications)
    }

    pub fn list_page(
        &self,
        business_id: i64,
        user_id: i64,
        limit: i32,
        token: &str,
    ) -> Result<NotificationPage, NotificationError> {
        let limit = if limit <= 0 {
            DEFAULT_PAGE_SIZE
        } else {
            limit.min(MAX_PAGE_SIZE)
        };
        let mut cursor_created_at = None;
        let mut cursor_id = None;
        if !token.is_empty() {
            let CreatedAtId { created_at, id } =
                decode_created_at_id(token).map_err(|_| NotificationError::InvalidCursor)?;
            cursor_created_at = Some(created_at);
            cursor_id = Some(id);
        }
        let rows = if let Some(paged) = self.paged.as_ref() {
            paged.list_page(
                business_id,
                user_id,
                limit + 1,
                cursor_created_at,
                cursor_id,
            )?
        } else {
            self.repo.list(business_id, user_id)?
        };
        let mut page = NotificationPage {
            notifications: rows,
            next_cursor: String::new(),
        };
        if page.notifications.len() > limit as usize {
            let last = page.notifications[limit as usize - 1].clone();
            page.notifications.truncate(limit as usize);
            page.next_cursor = encode_created_at_id(&CreatedAtId {
                created_at: last.created_at,
                id: last.id,
            });
        }
        Ok(page)
    }

    pub fn mark_read(
        &self,
        business_id: i64,
        user_id: i64,
        notification_id: i64,
    ) -> Result<(), NotificationError> {
        self.repo.mark_read(business_id, user_id, notification_id)
    }

    pub fn mark_all_read(&self, business_id: i64, user_id: i64) -> Result<(), NotificationError> {
        self.repo.mark_all_read(business_id, user_id)
    }

    pub fn create(&self, notification: &Notification) -> Result<(), NotificationError> {
        if notification.owner_id <= 0 || notification.business_id <= 0 || notification.user_id <= 0
        {
            return Err(NotificationError::MissingRecipients);
        }
        if notification.title.is_empty() || notification.notification_type.is_empty() {
            return Err(NotificationError::MissingTitleOrType);
        }
        self.repo.create(notification)
    }
}

#[cfg(test)]
mod tests {
    use super::{NotificationError, NotificationPage, Service, DEFAULT_PAGE_SIZE, MAX_PAGE_SIZE};
    use crate::domain::Notification;
    use chrono::{DateTime, Utc};
    use std::sync::Arc;

    struct FakeRepo {
        rows: Vec<Notification>,
    }

    impl super::Repository for FakeRepo {
        fn create(&self, _notification: &Notification) -> Result<(), NotificationError> {
            Ok(())
        }
        fn list(
            &self,
            _business_id: i64,
            _user_id: i64,
        ) -> Result<Vec<Notification>, NotificationError> {
            Ok(self.rows.clone())
        }
        fn mark_read(
            &self,
            _business_id: i64,
            _user_id: i64,
            _notification_id: i64,
        ) -> Result<(), NotificationError> {
            Ok(())
        }
        fn mark_all_read(&self, _business_id: i64, _user_id: i64) -> Result<(), NotificationError> {
            Ok(())
        }
    }

    struct FakePaged {
        rows: Vec<Notification>,
    }

    impl super::PagedRepository for FakePaged {
        fn list_page(
            &self,
            _business_id: i64,
            _user_id: i64,
            limit: i32,
            cursor_created_at: Option<chrono::DateTime<chrono::Utc>>,
            cursor_id: Option<i64>,
        ) -> Result<Vec<Notification>, NotificationError> {
            let start = match (cursor_created_at, cursor_id) {
                (Some(created_at), Some(id)) => self
                    .rows
                    .iter()
                    .position(|row| row.created_at == created_at && row.id == id)
                    .map(|index| index + 1)
                    .unwrap_or(0),
                _ => 0,
            };
            Ok(self
                .rows
                .iter()
                .skip(start)
                .take(limit as usize)
                .cloned()
                .collect())
        }
    }

    fn notification(id: i64) -> Notification {
        Notification {
            id,
            owner_id: 1,
            business_id: 1,
            user_id: 1,
            title: format!("t{id}"),
            body: String::new(),
            notification_type: "info".into(),
            is_read: false,
            link: String::new(),
            created_at: DateTime::<Utc>::from_timestamp(id, 0).unwrap(),
        }
    }

    #[test]
    fn page_limits_and_cursor() {
        let repo = Arc::new(FakeRepo {
            rows: (1..=5).map(notification).collect(),
        });
        let paged = Arc::new(FakePaged {
            rows: (1..=5).map(notification).collect(),
        });
        let service = Service::new(repo, Some(paged));
        assert_eq!(DEFAULT_PAGE_SIZE, 100);
        assert_eq!(MAX_PAGE_SIZE, 100);
        let page: NotificationPage = service.list_page(1, 1, 2, "").unwrap();
        assert_eq!(page.notifications.len(), 2);
        assert!(!page.next_cursor.is_empty());
        let next = service.list_page(1, 1, 2, &page.next_cursor).unwrap();
        assert_eq!(next.notifications.len(), 2);
        assert_eq!(next.notifications[0].id, 3);
    }

    #[test]
    fn create_validation() {
        let service = Service::new(Arc::new(FakeRepo { rows: Vec::new() }), None);
        let mut n = notification(1);
        n.owner_id = 0;
        assert_eq!(
            service.create(&n).unwrap_err(),
            NotificationError::MissingRecipients
        );
        let mut n = notification(1);
        n.title = String::new();
        assert_eq!(
            service.create(&n).unwrap_err(),
            NotificationError::MissingTitleOrType
        );
        assert!(service.create(&notification(1)).is_ok());
    }

    #[test]
    fn cache_key_and_scope() {
        assert_eq!(
            Service::notifications_key(1, 2, 10, Some(7)),
            "notifications:1:user:2|limit:10|cursor:7"
        );
        assert_eq!(
            Service::invalidate_scope(1, 2),
            vec!["user:2".to_string(), "notifications:1:user:2*".to_string()]
        );
    }
}
