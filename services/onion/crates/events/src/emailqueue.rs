//! Email queue, mirroring `internal/platform/emailqueue`.
//!
//! The SQL statements below are the same ones the Go sqlc layer runs on
//! `email_jobs`; `job_id` is the claim token that fences every ack/nack.
//! Delivery is at-least-once: claims expire after `CLAIM_LEASE`, so a
//! slow worker's stale write is fenced and discarded (see
//! `log_lost_claim` parity in `queue.go`).

use std::future::Future;
use std::pin::Pin;
use std::sync::atomic::AtomicI64;
use std::time::Duration;

use sqlx::Row;
use thiserror::Error;

pub const PRIORITY_CRITICAL: i32 = 10;
pub const PRIORITY_HIGH: i32 = 5;
pub const PRIORITY_NORMAL: i32 = 1;
pub const MAX_WORKERS: usize = 50;
pub const CLAIM_LEASE: Duration = Duration::from_secs(5 * 60);
pub const SEND_TIMEOUT: Duration = Duration::from_secs(2 * 60);
pub const RETENTION_INTERVAL: Duration = Duration::from_secs(3600);
pub const DEFAULT_RETENTION: Duration = Duration::from_secs(7 * 24 * 3600);

#[derive(Debug, Error)]
pub enum QueueError {
    #[error("email queue: {0}")]
    Other(String),
    #[error("sqlx: {0}")]
    Sqlx(#[from] sqlx::Error),
}

/// A unit of email to send, as rows in `email_jobs`.
#[derive(Debug, Clone, PartialEq)]
pub struct Job {
    pub priority: i32,
    pub to: String,
    pub subject: String,
    pub html: String,
    pub text: String,
    pub from: String,
    pub reply_to: String,
    pub dedupe_key: String,
}

pub fn normalize_priority(priority: i32) -> i32 {
    if priority <= 0 {
        PRIORITY_NORMAL
    } else {
        priority
    }
}

pub fn retry_delay(attempt: i32) -> i64 {
    match attempt {
        1 => 30,
        2 => 120,
        _ => 300,
    }
}

/// Go durations (`30m`, `15m`, `90s`, `24h`) for scheduler/queue env knobs.
#[derive(Debug, Error, PartialEq)]
pub enum DurationError {
    #[error("duration is empty")]
    Empty,
    #[error("unknown unit in duration: {0}")]
    UnknownUnit(String),
    #[error("invalid duration: {0}")]
    Invalid(String),
}

pub fn parse_go_duration(raw: &str) -> Result<Duration, DurationError> {
    let text = raw.trim();
    if text.is_empty() {
        return Err(DurationError::Empty);
    }
    if text == "0" {
        return Ok(Duration::ZERO);
    }
    let mut rest = text;
    let mut total_ns: f64 = 0.0;
    let mut saw_unit = false;
    while !rest.is_empty() {
        let number_end = rest
            .find(|c: char| !(c.is_ascii_digit() || c == '.' || c == '-' || c == '+'))
            .unwrap_or(rest.len());
        if number_end == 0 {
            return Err(DurationError::Invalid(text.to_string()));
        }
        let number: f64 = rest[..number_end]
            .parse()
            .map_err(|_| DurationError::Invalid(text.to_string()))?;
        let unit_start = number_end;
        let unit_end = rest[unit_start..]
            .find(|c: char| c.is_ascii_digit() || c == '.' || c == '-' || c == '+')
            .map(|i| unit_start + i)
            .unwrap_or(rest.len());
        let unit = &rest[unit_start..unit_end];
        let factor = match unit {
            "ns" => 1.0,
            "us" | "µs" | "μs" => 1e3,
            "ms" => 1e6,
            "s" => 1e9,
            "m" => 60e9,
            "h" => 3600e9,
            _ => return Err(DurationError::UnknownUnit(unit.to_string())),
        };
        total_ns += number * factor;
        saw_unit = true;
        rest = &rest[unit_end..];
    }
    if !saw_unit || !total_ns.is_finite() {
        return Err(DurationError::Invalid(text.to_string()));
    }
    Ok(Duration::from_nanos(total_ns.max(0.0) as u64))
}

pub type Sender = std::sync::Arc<
    dyn Fn(Job) -> Pin<Box<dyn Future<Output = Result<(), String>> + Send>> + Send + Sync,
>;

pub type EnvFn = Box<dyn Fn(&str) -> Option<String> + Send + Sync>;

pub struct Pool {
    pub pool: sqlx::MySqlPool,
    pub sender: Sender,
    pub env: EnvFn,
    pub active: AtomicI64,
}

fn process_env(key: &str) -> Option<String> {
    std::env::var(key).ok()
}

/// Enqueue statement, kept as a const so the Go/Rust SQL stays visibly
/// identical in review.
pub const ENQUEUE_SQL: &str = r#"INSERT INTO email_jobs (priority, to_email, subject, html, text_body, from_email, reply_to)
VALUES (?, ?, ?, ?, ?, ?, ?)"#;

pub const ENQUEUE_DEDUPE_SQL: &str = r#"INSERT INTO email_jobs (priority, to_email, subject, html, text_body, from_email, reply_to, dedupe_key)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE id = id"#;

pub const CLAIM_SQL: &str = r#"UPDATE email_jobs
SET job_id = ?, status = 'processing'
WHERE status = 'pending' AND job_id IS NULL AND scheduled_at <= NOW()
ORDER BY (priority + LEAST(TIMESTAMPDIFF(MINUTE, created_at, NOW()) DIV 60, 10)) DESC,
         scheduled_at ASC
LIMIT 1"#;

pub const GET_CLAIMED_SQL: &str = r#"SELECT id, to_email, subject, html, text_body, from_email, reply_to, attempts, max_attempts
FROM email_jobs WHERE job_id = ? AND status = 'processing'"#;

pub const RELEASE_CLAIM_SQL: &str = r#"UPDATE email_jobs
SET status = 'pending', job_id = NULL, scheduled_at = NOW()
WHERE status = 'processing' AND job_id = ?"#;

pub const MARK_SENT_SQL: &str = r#"UPDATE email_jobs
SET status = 'sent', sent_at = NOW(), attempts = ?, job_id = NULL
WHERE id = ? AND status = 'processing' AND job_id = ?"#;

pub const RESCHEDULE_SQL: &str = r#"UPDATE email_jobs
SET status = ?, attempts = ?, last_error = ?, job_id = NULL,
    scheduled_at = DATE_ADD(NOW(), INTERVAL ? SECOND)
WHERE id = ? AND status = 'processing' AND job_id = ?"#;

pub const RECOVER_STUCK_SQL: &str = r#"UPDATE email_jobs
SET status = CASE WHEN attempts + 1 >= max_attempts THEN 'failed' ELSE 'pending' END,
    attempts = attempts + 1,
    job_id = NULL,
    last_error = COALESCE(last_error, 'worker claim expired during delivery')
WHERE status = 'processing' AND updated_at < DATE_SUB(NOW(), INTERVAL ? SECOND)"#;

pub const DELETE_FINISHED_SQL: &str = r#"DELETE FROM email_jobs
WHERE status IN ('sent', 'failed') AND created_at < ?"#;

#[derive(Debug, Clone)]
pub struct ClaimedJob {
    pub id: u64,
    pub to_email: String,
    pub subject: String,
    pub html: String,
    pub text_body: Option<String>,
    pub from_email: Option<String>,
    pub reply_to: Option<String>,
    pub attempts: u8,
    pub max_attempts: u8,
}

impl Pool {
    pub fn new(pool: sqlx::MySqlPool, sender: Sender) -> Self {
        Self::with_env(pool, sender, Box::new(process_env))
    }

    pub fn with_env(pool: sqlx::MySqlPool, sender: Sender, env: EnvFn) -> Self {
        Self {
            pool,
            sender,
            env,
            active: AtomicI64::new(0),
        }
    }

    pub async fn enqueue(&self, job: &Job) -> Result<(), QueueError> {
        let sealed_html = auth::emailcrypto::seal_html(&job.html, &*self.env);
        let sealed_text = auth::emailcrypto::seal_html(&job.text, &*self.env);
        let priority = normalize_priority(job.priority);
        if job.dedupe_key.is_empty() {
            sqlx::query(ENQUEUE_SQL)
                .bind(priority as i8)
                .bind(&job.to)
                .bind(&job.subject)
                .bind(&sealed_html)
                .bind(optional(&sealed_text))
                .bind(optional(&job.from))
                .bind(optional(&job.reply_to))
                .execute(&self.pool)
                .await?;
            Ok(())
        } else {
            sqlx::query(ENQUEUE_DEDUPE_SQL)
                .bind(priority as i8)
                .bind(&job.to)
                .bind(&job.subject)
                .bind(&sealed_html)
                .bind(optional(&sealed_text))
                .bind(optional(&job.from))
                .bind(optional(&job.reply_to))
                .bind(optional(&job.dedupe_key))
                .execute(&self.pool)
                .await?;
            Ok(())
        }
    }

    /// Claim one job with a fresh token; rows affected is the win count.
    pub async fn claim(&self) -> Result<Option<String>, QueueError> {
        let token = uuid::Uuid::new_v4().to_string();
        let result = sqlx::query(CLAIM_SQL)
            .bind(&token)
            .execute(&self.pool)
            .await?;
        if result.rows_affected() == 0 {
            Ok(None)
        } else {
            Ok(Some(token))
        }
    }

    pub async fn get_claimed(&self, token: &str) -> Result<Option<ClaimedJob>, QueueError> {
        let row = sqlx::query(GET_CLAIMED_SQL)
            .bind(token)
            .fetch_optional(&self.pool)
            .await?;
        Ok(row.map(|r| {
            let id: u64 = r.try_get("id").unwrap_or(0);
            let attempts: u8 = r.try_get("attempts").unwrap_or(0);
            let max_attempts: u8 = r.try_get("max_attempts").unwrap_or(0);
            ClaimedJob {
                id,
                to_email: r.try_get("to_email").unwrap_or_default(),
                subject: r.try_get("subject").unwrap_or_default(),
                html: r.try_get("html").unwrap_or_default(),
                text_body: r.try_get("text_body").unwrap_or(None),
                from_email: r.try_get("from_email").unwrap_or(None),
                reply_to: r.try_get("reply_to").unwrap_or(None),
                attempts,
                max_attempts,
            }
        }))
    }

    pub async fn release_claim(&self, token: &str) -> Result<u64, QueueError> {
        let result = sqlx::query(RELEASE_CLAIM_SQL)
            .bind(token)
            .execute(&self.pool)
            .await?;
        Ok(result.rows_affected())
    }

    pub async fn mark_sent(&self, id: u64, token: &str, attempts: u8) -> Result<u64, QueueError> {
        let result = sqlx::query(MARK_SENT_SQL)
            .bind(attempts as i8)
            .bind(id)
            .bind(token)
            .execute(&self.pool)
            .await?;
        Ok(result.rows_affected())
    }

    pub async fn reschedule(
        &self,
        id: u64,
        token: &str,
        status: &str,
        attempts: u8,
        last_error: &str,
        delay_secs: i64,
    ) -> Result<u64, QueueError> {
        let result = sqlx::query(RESCHEDULE_SQL)
            .bind(status)
            .bind(attempts)
            .bind(last_error)
            .bind(delay_secs)
            .bind(id)
            .bind(token)
            .execute(&self.pool)
            .await?;
        Ok(result.rows_affected())
    }

    /// Recover claims whose worker died before the ack.
    pub async fn recover_stuck(&self) -> Result<u64, QueueError> {
        let stale_secs = CLAIM_LEASE.as_secs() as i64;
        let result = sqlx::query(RECOVER_STUCK_SQL)
            .bind(stale_secs)
            .execute(&self.pool)
            .await?;
        Ok(result.rows_affected())
    }

    /// Retention purge for sent/failed history older than the cutoff.
    pub async fn cleanup_old(&self) -> Result<u64, QueueError> {
        let cutoff = chrono::Utc::now()
            - chrono::TimeDelta::from_std(DEFAULT_RETENTION)
                .map_err(|_| QueueError::Other("retention window".to_string()))?;
        let result = sqlx::query(DELETE_FINISHED_SQL)
            .bind(cutoff.naive_utc())
            .execute(&self.pool)
            .await?;
        Ok(result.rows_affected())
    }

    /// One claim-and-send iteration. Returns whether the loop should keep
    /// draining (i.e. a job was actually claimed). Sends are bounded by
    /// `SEND_TIMEOUT` so a hung provider cannot outlive the claim lease.
    pub async fn run_once(&self) -> Result<bool, QueueError> {
        let Some(token) = self.claim().await? else {
            return Ok(false);
        };
        let Some(job) = self.get_claimed(&token).await? else {
            self.release_claim(&token).await?;
            return Ok(false);
        };
        let html = auth::emailcrypto::open_html(&job.html, &*self.env);
        if !job.html.is_empty() && html.is_empty() {
            // Undecryptable bodies fail permanently (no infinite loop).
            self.reschedule_send_error(
                job.id,
                &token,
                job.max_attempts,
                job.max_attempts,
                "body undecryptable (EMAIL_ENCRYPTION_KEY missing or changed)",
            )
            .await?;
            return Ok(true);
        }
        let message = Job {
            priority: PRIORITY_NORMAL,
            to: job.to_email.clone(),
            subject: job.subject.clone(),
            html,
            text: job
                .text_body
                .clone()
                .map(|text| auth::emailcrypto::open_html(&text, &*self.env))
                .unwrap_or_default(),
            from: job.from_email.clone().unwrap_or_default(),
            reply_to: job.reply_to.clone().unwrap_or_default(),
            dedupe_key: String::new(),
        };
        let send = (self.sender)(message);
        let outcome = tokio::time::timeout(SEND_TIMEOUT, send).await;
        match outcome {
            Ok(Ok(())) => {
                self.mark_sent(job.id, &token, job.attempts.saturating_add(1))
                    .await?;
            }
            Ok(Err(error)) => {
                self.reschedule_send_error(job.id, &token, job.attempts, job.max_attempts, &error)
                    .await?;
            }
            Err(_) => {
                self.reschedule_send_error(
                    job.id,
                    &token,
                    job.attempts,
                    job.max_attempts,
                    "send timed out",
                )
                .await?;
            }
        }
        Ok(true)
    }

    async fn reschedule_send_error(
        &self,
        id: u64,
        token: &str,
        attempts: u8,
        max_attempts: u8,
        error: &str,
    ) -> Result<u64, QueueError> {
        let attempts = attempts.saturating_add(1);
        let status = if attempts as i64 >= max_attempts as i64 {
            "failed"
        } else {
            "pending"
        };
        self.reschedule(
            id,
            token,
            status,
            attempts,
            error,
            retry_delay(i32::from(attempts)),
        )
        .await
    }
}

fn optional(value: &str) -> Option<&str> {
    if value.is_empty() {
        None
    } else {
        Some(value)
    }
}

pub fn dedupe_key_for(box_: &str, extra: &str) -> String {
    if extra.is_empty() {
        box_.to_string()
    } else {
        format!("{box_}:{extra}")
    }
}

#[cfg(test)]
mod tests {
    use super::{normalize_priority, parse_go_duration, retry_delay, PRIORITY_CRITICAL};
    use std::time::Duration;

    #[test]
    fn priorities_and_delays() {
        assert_eq!(normalize_priority(0), 1);
        assert_eq!(normalize_priority(-3), 1);
        assert_eq!(normalize_priority(10), 10);
        assert_eq!(retry_delay(1), 30);
        assert_eq!(retry_delay(2), 120);
        assert_eq!(retry_delay(5), 300);
        assert_eq!(PRIORITY_CRITICAL, 10);
    }

    #[test]
    fn durations() {
        assert_eq!(parse_go_duration("30m").unwrap(), Duration::from_secs(1800));
        assert_eq!(parse_go_duration("15m").unwrap(), Duration::from_secs(900));
        assert_eq!(parse_go_duration("90s").unwrap(), Duration::from_secs(90));
        assert!(parse_go_duration("").is_err());
        assert!(parse_go_duration("0").is_ok());
        assert!(parse_go_duration("banana").is_err());
    }

    #[test]
    fn sql_consts_cover_claim_lifecycle() {
        for sql in [
            super::CLAIM_SQL,
            super::GET_CLAIMED_SQL,
            super::MARK_SENT_SQL,
            super::RESCHEDULE_SQL,
            super::RECOVER_STUCK_SQL,
            super::ENQUEUE_DEDUPE_SQL,
        ] {
            assert!(
                sql.contains("email_jobs") || sql.contains("job_id"),
                "{}",
                sql
            );
        }
        assert!(super::ENQUEUE_DEDUPE_SQL.contains("ON DUPLICATE KEY UPDATE"));
        assert!(super::CLAIM_SQL.contains("ORDER BY") || super::CLAIM_SQL.contains("order"));
    }

    #[test]
    fn dedupe_key_shape() {
        assert_eq!(
            super::dedupe_key_for("newsletter", "list"),
            "newsletter:list"
        );
    }
}
