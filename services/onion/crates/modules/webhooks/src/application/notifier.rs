//! Dead-delivery notifications. Mirrors `notifier.go`.

use crate::application::delivery_repository::{
    NotificationContactResolver, NotificationEmail, NotificationQueue, NotificationRepository,
};
use crate::domain::FailingRegistration;
use chrono::{DateTime, Timelike, Utc};
use std::sync::Arc;

/// Bounds which failures are worth mentioning. A registration that broke
/// months ago and was never fixed still has dead rows; without this it
/// would generate a notification every day forever.
pub const DEAD_NOTIFY_WITHIN_HOURS: i64 = 24;
/// Bounds one pass. One notification per failing registration, so this is
/// a ceiling on how many businesses are told in a single run.
pub const DEAD_NOTIFY_BATCH: i32 = 100;

/// Tells the owner of a registration that it has stopped receiving events.
pub struct Notifier<R: NotificationRepository, Q: NotificationQueue, C: NotificationContactResolver>
{
    repo: Arc<R>,
    queue: Arc<Q>,
    contacts: Option<Arc<C>>,
    within_hours: i64,
    batch_size: i32,
}

impl<R: NotificationRepository, Q: NotificationQueue, C: NotificationContactResolver>
    Notifier<R, Q, C>
{
    pub fn new(repo: Arc<R>, queue: Arc<Q>, contacts: Option<Arc<C>>) -> Self {
        Notifier {
            repo,
            queue,
            contacts,
            within_hours: DEAD_NOTIFY_WITHIN_HOURS,
            batch_size: DEAD_NOTIFY_BATCH,
        }
    }

    /// Send at most one message per failing registration. A registration
    /// that cannot be notified is skipped rather than aborting the pass.
    pub fn notify_failing(&self) -> Result<(), String> {
        let now = Utc::now();
        let notified_on = now
            .with_hour(0)
            .and_then(|value| value.with_minute(0))
            .and_then(|value| value.with_second(0))
            .and_then(|value| value.with_nanosecond(0))
            .unwrap_or(now);
        let failing = self
            .repo
            .list_failing_registrations(self.within_hours, self.batch_size, notified_on)
            .map_err(|error| format!("listing failing registrations: {error}"))?;
        for registration in failing {
            self.notify_one(&registration, notified_on);
        }
        Ok(())
    }

    /// Tell one registration's owner, once a day at most.
    pub fn notify_one(&self, registration: &FailingRegistration, notified_on: DateTime<Utc>) {
        let owner_id = match self
            .repo
            .get_notification_owner(registration.webhook_id, registration.business_id)
        {
            Ok(owner_id) => owner_id,
            Err(_) => return,
        };
        let Some(contacts) = self.contacts.as_ref() else {
            return;
        };
        let resolved = contacts.resolve(registration.business_id, owner_id);
        let recipient = match resolved {
            Ok(recipient) => recipient,
            // Lookup failure: do not record (Go returns without
            // recording).
            Err(_) => return,
        };
        let Some((email, name, active)) = recipient else {
            // `None` = creator left/inactive: record so the day's window
            // closes.
            self.record_notice(registration, notified_on);
            return;
        };
        if !active || email.is_empty() {
            self.record_notice(registration, notified_on);
            return;
        }
        // One message per registration per day. The key is what makes this
        // safe to run hourly over a query that keeps returning the same
        // rows: later enqueues in a day are no-ops.
        let dedupe_key = format!(
            "webhook_dead:{}:{}",
            registration.webhook_id,
            notified_on.format("%Y-%m-%d")
        );
        let email = NotificationEmail {
            to: email,
            subject: "A webhook endpoint has stopped receiving events".into(),
            html: dead_delivery_html(registration, &name),
            text: dead_delivery_text(registration, &name),
            dedupe_key,
        };
        if self.queue.enqueue(&email).is_ok() {
            self.record_notice(registration, notified_on);
        }
    }

    fn record_notice(&self, registration: &FailingRegistration, notified_on: DateTime<Utc>) {
        let _ = self.repo.record_dead_delivery_notice(
            registration.business_id,
            registration.webhook_id,
            notified_on,
        );
    }
}

/// Renders the message. The URL is the tenant's own and is escaped because
/// it lands in an HTML body, not because it is untrusted.
pub fn dead_delivery_html(registration: &FailingRegistration, name: &str) -> String {
    let greeting = if name.is_empty() {
        String::new()
    } else {
        format!("<p>Hi {},</p>", html_escape(name))
    };
    format!(
        "{greeting}<p>{} to your webhook endpoint <code>{}</code> could not be delivered and have been given up on.</p>\
<p>Events are retried several times before this happens, so the endpoint has been unreachable \
or returning errors for a while. Anything already given up on will not be sent again.</p>\
<p>The delivery attempts, including what your endpoint answered, are available from the \
developer settings for this business.</p>",
        html_escape(&delivery_count_phrase(registration.dead_count)),
        html_escape(&registration.url),
    )
}

pub fn dead_delivery_text(registration: &FailingRegistration, name: &str) -> String {
    let greeting = if name.is_empty() {
        String::new()
    } else {
        format!("Hi {name},\n\n")
    };
    format!(
        "{greeting}{} to your webhook endpoint {} could not be delivered and have been given up on.\n\n\
Events are retried several times before this happens, so the endpoint has been unreachable \
or returning errors for a while. Anything already given up on will not be sent again.\n\n\
The delivery attempts, including what your endpoint answered, are available from the \
developer settings for this business.",
        delivery_count_phrase(registration.dead_count),
        registration.url,
    )
}

/// Keeps the message readable at one, which is the most common case.
pub fn delivery_count_phrase(count: i64) -> String {
    if count == 1 {
        "One event".into()
    } else {
        format!("{count} events")
    }
}

fn html_escape(text: &str) -> String {
    text.replace('&', "&amp;")
        .replace('<', "&lt;")
        .replace('>', "&gt;")
        .replace('"', "&quot;")
        .replace('\'', "&#39;")
}

#[cfg(test)]
mod tests {
    use super::{dead_delivery_html, dead_delivery_text, delivery_count_phrase, Notifier};
    use crate::domain::FailingRegistration;
    use chrono::{DateTime, Utc};

    type ContactResult = Result<Option<(String, String, bool)>, crate::application::AppError>;
    type ContactFn = dyn Fn(i64, i64) -> ContactResult + Send + Sync;

    fn registration() -> FailingRegistration {
        FailingRegistration {
            business_id: 1,
            webhook_id: 2,
            url: "https://example.com/hook".into(),
            dead_count: 1,
            last_dead_at: DateTime::<Utc>::from_timestamp(0, 0).unwrap(),
        }
    }

    #[test]
    fn phrases() {
        assert_eq!(delivery_count_phrase(1), "One event");
        assert_eq!(delivery_count_phrase(5), "5 events");
    }

    #[test]
    fn messages_escape_the_url() {
        let mut reg = registration();
        reg.url = "https://example.com/<script>".into();
        let html = dead_delivery_html(&reg, "Ada");
        assert!(html.contains("&lt;script&gt;"));
        assert!(html.contains("Hi Ada,"));
        let text = dead_delivery_text(&reg, "");
        assert!(text.contains("https://example.com/<script>"));
    }

    #[test]
    fn escapes_the_recipient_name_too() {
        let reg = registration();
        let html = dead_delivery_html(&reg, "<b>Ada</b>");
        assert!(html.contains("Hi &lt;b&gt;Ada&lt;/b&gt;,"));
        assert!(!html.contains("<b>Ada</b>"));
    }

    struct FakeRepo {
        notices: std::sync::Arc<std::sync::Mutex<Vec<(i64, i64)>>>,
    }

    impl crate::application::delivery_repository::NotificationRepository for FakeRepo {
        fn list_failing_registrations(
            &self,
            _within_hours: i64,
            _limit: i32,
            _since: chrono::DateTime<chrono::Utc>,
        ) -> Result<Vec<FailingRegistration>, crate::application::AppError> {
            Ok(vec![registration()])
        }
        fn get_notification_owner(
            &self,
            _webhook_id: i64,
            _business_id: i64,
        ) -> Result<i64, crate::application::AppError> {
            Ok(5)
        }
        fn record_dead_delivery_notice(
            &self,
            business_id: i64,
            webhook_id: i64,
            _at: chrono::DateTime<chrono::Utc>,
        ) -> Result<(), crate::application::AppError> {
            self.notices.lock().unwrap().push((business_id, webhook_id));
            Ok(())
        }
    }

    struct FakeQueue;

    impl crate::application::delivery_repository::NotificationQueue for FakeQueue {
        fn enqueue(
            &self,
            _email: &crate::application::delivery_repository::NotificationEmail,
        ) -> Result<(), crate::application::AppError> {
            Ok(())
        }
    }

    struct FakeContacts(std::sync::Arc<ContactFn>);

    impl crate::application::delivery_repository::NotificationContactResolver for FakeContacts {
        fn resolve(
            &self,
            business_id: i64,
            user_id: i64,
        ) -> Result<Option<(String, String, bool)>, crate::application::AppError> {
            (self.0)(business_id, user_id)
        }
    }

    fn notifier_with(
        notices: std::sync::Arc<std::sync::Mutex<Vec<(i64, i64)>>>,
        contacts: std::sync::Arc<ContactFn>,
    ) -> Notifier<FakeRepo, FakeQueue, FakeContacts> {
        let repo = std::sync::Arc::new(FakeRepo {
            notices: notices.clone(),
        });
        let queue = std::sync::Arc::new(FakeQueue);
        let resolver = std::sync::Arc::new(FakeContacts(contacts));
        Notifier::new(repo, queue, Some(resolver))
    }

    #[test]
    fn resolver_error_does_not_record_a_notice() {
        let notices = std::sync::Arc::new(std::sync::Mutex::new(Vec::new()));
        let notifier = notifier_with(
            notices.clone(),
            std::sync::Arc::new(|_, _| Err(crate::application::AppError::Repo("boom".into()))),
        );
        notifier.notify_failing().unwrap();
        assert!(notices.lock().unwrap().is_empty());
    }

    #[test]
    fn inactive_member_records_but_does_not_email() {
        let notices = std::sync::Arc::new(std::sync::Mutex::new(Vec::new()));
        let notifier = notifier_with(
            notices.clone(),
            std::sync::Arc::new(|_, _| Ok(Some(("ada@example.com".into(), "Ada".into(), false)))),
        );
        notifier.notify_failing().unwrap();
        assert_eq!(*notices.lock().unwrap(), vec![(1, 2)]);
    }
}
