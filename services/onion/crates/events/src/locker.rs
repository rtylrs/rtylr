//! MySQL distributed lock, mirroring `internal/platform/jobs/mysql_locker.go`.
//!
//! Lock names are `onion.job.<name>` capped at MySQL's 64-byte limit with a
//! SHA-256 hex suffix. `GET_LOCK(name, 0)` is non-blocking: a busy worker
//! skips instead of queueing (503-on-collision in the scheduler).

use std::future::Future;
use std::pin::Pin;

use crate::jobs::{JobError, Locker};
use sha2::{Digest, Sha256};

pub const DEFAULT_LOCK_PREFIX: &str = "onion.job.";
pub const MYSQL_LOCK_NAME_LIMIT: usize = 64;

pub fn lock_name(prefix: &str, name: &str) -> String {
    let prefix = prefix.trim();
    let prefix = if prefix.is_empty() {
        DEFAULT_LOCK_PREFIX
    } else {
        prefix
    };
    let full = format!("{prefix}{name}");
    if full.len() <= MYSQL_LOCK_NAME_LIMIT {
        return full;
    }
    let hex_sum: String = Sha256::digest(full.as_bytes())
        .iter()
        .map(|byte| format!("{byte:02x}"))
        .collect();
    let space = MYSQL_LOCK_NAME_LIMIT.saturating_sub(prefix.len());
    if space == 0 {
        return hex_sum[..MYSQL_LOCK_NAME_LIMIT].to_string();
    }
    format!("{prefix}{}", &hex_sum[..space])
}

pub fn default_lock_name(name: &str) -> String {
    lock_name(DEFAULT_LOCK_PREFIX, name)
}

/// A held MySQL user lock; the connection stays held until `release`.
pub struct MysqlLockHandle {
    pub conn: Option<sqlx::pool::PoolConnection<sqlx::MySql>>,
    pub lock_name: String,
}

impl MysqlLockHandle {
    /// Release the lock on the same connection that took it and close the
    /// session (mirrors Go's `conn.Close()` in the release path).
    pub async fn release(&mut self) -> Result<(), JobError> {
        let Some(conn) = self.conn.as_mut() else {
            return Ok(());
        };
        let released: Option<i64> = sqlx::query_scalar("SELECT RELEASE_LOCK(?)")
            .bind(self.lock_name.clone())
            .fetch_optional(&mut **conn)
            .await
            .map_err(|error| JobError::Failed("RELEASE_LOCK".to_string(), error.to_string()))?;
        self.conn = None;
        match released {
            Some(0) => Err(JobError::Failed(
                format!(
                    "mysql lock {:?} was not held by this connection",
                    self.lock_name
                ),
                String::new(),
            )),
            _ => Ok(()),
        }
    }
}

/// Per-job locks via MySQL user locks on a dedicated connection.
/// Callers pass the pool from the composition root (`db::Database`).
#[derive(Clone)]
pub struct MySqlLocker {
    pool: sqlx::MySqlPool,
    prefix: String,
}

impl MySqlLocker {
    pub fn new(pool: sqlx::MySqlPool) -> Self {
        Self {
            pool,
            prefix: DEFAULT_LOCK_PREFIX.to_string(),
        }
    }

    pub fn with_prefix(pool: sqlx::MySqlPool, prefix: String) -> Self {
        Self { pool, prefix }
    }
}

impl Locker for MySqlLocker {
    fn try_lock<'a>(
        &'a self,
        name: &'a str,
    ) -> Pin<Box<dyn Future<Output = Result<Option<MysqlLockHandle>, JobError>> + Send + 'a>> {
        let lock_name = lock_name(&self.prefix, name);
        let pool = self.pool.clone();
        Box::pin(async move {
            let mut conn = pool
                .acquire()
                .await
                .map_err(|error| JobError::Failed("lock acquire".to_string(), error.to_string()))?;
            let acquired: Option<i64> = sqlx::query_scalar("SELECT GET_LOCK(?, 0)")
                .bind(lock_name.clone())
                .fetch_optional(&mut *conn)
                .await
                .map_err(|error| JobError::Failed("GET_LOCK".to_string(), error.to_string()))?;
            let Some(value) = acquired else {
                return Ok(None);
            };
            if value != 1 {
                return Ok(None);
            }
            Ok(Some(MysqlLockHandle {
                conn: Some(conn),
                lock_name,
            }))
        })
    }
}

#[cfg(test)]
mod tests {
    use super::{default_lock_name, lock_name};

    #[test]
    fn lock_name_fits_limit() {
        let name = default_lock_name(&"very-long-job-name-".repeat(10));
        assert!(name.len() <= 64);
        assert!(name.starts_with("onion.job."));
        let name = lock_name("onion.job.", "eventbus.dispatch_events");
        assert_eq!(name, "onion.job.eventbus.dispatch_events");
    }

    #[test]
    fn lock_name_short_is_literal() {
        assert_eq!(default_lock_name("x"), "onion.job.x");
    }
}
