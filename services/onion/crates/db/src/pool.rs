//! sqlx pools, mirroring `internal/platform/database/dbclient`.
//!
//! Pool sizes and lifetimes mirror the Go defaults (25 max connections,
//! 30m max lifetime, 10m idle time, 10s connect timeout). DSN resolution
//! order is `DATABASE_URL` first, else the `DB_*` local DSN.

use std::time::Duration;

use sea_query::{Condition, Expr, IntoCondition};
use sqlx::MySql;
use thiserror::Error;

use crate::db_name_from_dsn;
use crate::safety::check_dsn_safety;

pub const DEFAULT_MAX_OPEN_CONNS: u32 = 25;
pub const DEFAULT_MAX_IDLE_CONNS: u32 = 10;
pub const DEFAULT_CONN_MAX_LIFETIME_SECS: u64 = 30 * 60;
pub const DEFAULT_CONN_MAX_IDLE_TIME_SECS: u64 = 10 * 60;
pub const CONNECT_TIMEOUT_SECS: u64 = 10;

#[derive(Debug, Error)]
pub enum DbError {
    #[error("missing critical environment variable(s): {0}")]
    Missing(String),
    #[error("sqlx: {0}")]
    Sqlx(#[from] sqlx::Error),
    #[error("safety: {0}")]
    Safety(#[from] crate::safety::SafetyError),
}

/// A sqlx pool plus optional logging / replica pools.
#[derive(Debug, Clone)]
pub struct Database {
    pool: sqlx::MySqlPool,
    logging_pool: Option<sqlx::MySqlPool>,
    replica_pool: Option<sqlx::MySqlPool>,
}

#[derive(Debug, Clone, PartialEq)]
pub struct PoolLimits {
    pub max_connections: u32,
    pub max_lifetime: Duration,
    pub idle_timeout: Duration,
}

impl PoolLimits {
    pub fn from_env(config: &onion_core::config::AppConfig) -> Self {
        let positive = |value: i32, default: u32| -> u32 {
            if value > 0 {
                value as u32
            } else {
                default
            }
        };
        Self {
            max_connections: positive(config.db_max_open_conns, DEFAULT_MAX_OPEN_CONNS),
            max_lifetime: if config.db_conn_max_lifetime.is_zero() {
                Duration::from_secs(DEFAULT_CONN_MAX_LIFETIME_SECS)
            } else {
                config.db_conn_max_lifetime
            },
            idle_timeout: if config.db_conn_max_idle_time.is_zero() {
                let _ = DEFAULT_MAX_IDLE_CONNS;
                Duration::from_secs(DEFAULT_CONN_MAX_IDLE_TIME_SECS)
            } else {
                config.db_conn_max_idle_time
            },
        }
    }
}

impl Default for PoolLimits {
    fn default() -> Self {
        Self {
            max_connections: DEFAULT_MAX_OPEN_CONNS,
            max_lifetime: Duration::from_secs(DEFAULT_CONN_MAX_LIFETIME_SECS),
            idle_timeout: Duration::from_secs(DEFAULT_CONN_MAX_IDLE_TIME_SECS),
        }
    }
}

fn build_local_dsn(config: &onion_core::config::AppConfig) -> String {
    format!(
        "{}:{}@tcp({}:{})/{}?parseTime=true&loc=UTC&time_zone='UTC'&multiStatements=true&charset=utf8mb4&collation=utf8mb4_unicode_ci",
        config.db_user,
        config.db_pass,
        if config.db_host.is_empty() {
            "localhost".to_string()
        } else {
            config.db_host.clone()
        },
        if config.db_port.is_empty() {
            "3306".to_string()
        } else {
            config.db_port.clone()
        },
        config.db_name,
    )
}

pub fn resolve_main_dsn(config: &onion_core::config::AppConfig) -> String {
    if !config.database_url.is_empty() {
        config.database_url.clone()
    } else {
        build_local_dsn(config)
    }
}

fn build_replica_dsn(config: &onion_core::config::AppConfig) -> Option<String> {
    let host = std::env::var("DB_REPLICA_HOST")
        .ok()
        .filter(|v| !v.is_empty())?;
    let user = std::env::var("DB_AI_USER").ok().filter(|v| !v.is_empty())?;
    let password = std::env::var("DB_AI_PASSWORD")
        .ok()
        .filter(|v| !v.is_empty())?;
    let port = std::env::var("DB_REPLICA_PORT")
        .ok()
        .filter(|v| !v.is_empty())
        .unwrap_or_else(|| {
            if !config.db_port.is_empty() {
                config.db_port.clone()
            } else {
                "3306".to_string()
            }
        });
    let db_name = match std::env::var("DB_NAME") {
        Ok(name) if !name.is_empty() => name,
        _ => {
            let from_main = db_name_from_dsn(&resolve_main_dsn(config));
            if from_main.is_empty() {
                "onion_dev".to_string()
            } else {
                from_main
            }
        }
    };
    Some(format!(
        "{user}:{password}@tcp({host}:{port})/{db_name}?parseTime=true&loc=UTC&time_zone='UTC'&charset=utf8mb4&collation=utf8mb4_unicode_ci"
    ))
}

async fn open_pool(dsn: &str, limits: PoolLimits) -> Result<sqlx::MySqlPool, DbError> {
    let url = crate::to_sqlx_url(dsn)
        .ok_or_else(|| DbError::Missing("unparseable database DSN".to_string()))?;
    let options = sqlx::mysql::MySqlPoolOptions::new()
        .max_connections(limits.max_connections)
        .max_lifetime(limits.max_lifetime)
        .idle_timeout(limits.idle_timeout)
        .acquire_timeout(Duration::from_secs(CONNECT_TIMEOUT_SECS));
    options.connect(&url).await.map_err(DbError::Sqlx)
}

impl Database {
    pub async fn connect_main(config: &onion_core::config::AppConfig) -> Result<Self, DbError> {
        let dsn = resolve_main_dsn(config);
        check_dsn_safety(
            "database",
            &dsn,
            config.is_declared_production(),
            std::env::var("ALLOW_PRODUCTION_DB")
                .unwrap_or_default()
                .trim()
                .eq_ignore_ascii_case("true"),
            &onion_core::config::normalize_app_env(&config.app_env),
            std::env::var("ALLOW_REMOTE_DB")
                .unwrap_or_default()
                .trim()
                .eq_ignore_ascii_case("true"),
        )?;
        let limits = PoolLimits::from_env(config);
        let pool = open_pool(&dsn, limits.clone()).await?;

        let logging = match config.logging_database_url.is_empty() {
            true => None,
            false => {
                check_dsn_safety(
                    "logging database",
                    &config.logging_database_url,
                    config.is_declared_production(),
                    std::env::var("ALLOW_PRODUCTION_DB")
                        .unwrap_or_default()
                        .trim()
                        .eq_ignore_ascii_case("true"),
                    &onion_core::config::normalize_app_env(&config.app_env),
                    std::env::var("ALLOW_REMOTE_DB")
                        .unwrap_or_default()
                        .trim()
                        .eq_ignore_ascii_case("true"),
                )?;
                Some(open_pool(&config.logging_database_url, limits.clone()).await?)
            }
        };

        let replica = match build_replica_dsn(config) {
            Some(replica_dsn) => match check_dsn_safety(
                "read replica",
                &replica_dsn,
                config.is_declared_production(),
                std::env::var("ALLOW_PRODUCTION_DB")
                    .unwrap_or_default()
                    .trim()
                    .eq_ignore_ascii_case("true"),
                &onion_core::config::normalize_app_env(&config.app_env),
                std::env::var("ALLOW_REMOTE_DB")
                    .unwrap_or_default()
                    .trim()
                    .eq_ignore_ascii_case("true"),
            ) {
                Ok(()) => open_pool(&replica_dsn, limits).await.ok(),
                Err(_) => None,
            },
            None => None,
        };

        Ok(Self {
            pool,
            logging_pool: logging,
            replica_pool: replica,
        })
    }

    /// Test handle: construct without connecting (`UseForTesting` equivalent).
    pub fn from_pool(pool: sqlx::MySqlPool) -> Self {
        Self {
            pool,
            logging_pool: None,
            replica_pool: None,
        }
    }

    pub fn pool(&self) -> &sqlx::MySqlPool {
        &self.pool
    }

    pub fn logging_pool(&self) -> Option<&sqlx::MySqlPool> {
        self.logging_pool.as_ref()
    }

    /// Read-replica pool for AI reads, falling back to the primary.
    pub fn replica_pool(&self) -> &sqlx::MySqlPool {
        self.replica_pool.as_ref().unwrap_or(&self.pool)
    }

    pub async fn ping(&self) -> Result<(), DbError> {
        sqlx::query("SELECT 1")
            .execute(&self.pool)
            .await
            .map(|_| ())
            .map_err(DbError::Sqlx)
    }

    /// Run `f` in a transaction; commit on `Ok`, roll back on `Err`.
    pub async fn with_tx<F, Fut, T, E>(&self, f: F) -> Result<T, DbError>
    where
        F: FnOnce(&mut sqlx::Transaction<'_, MySql>) -> Fut,
        Fut: std::future::Future<Output = Result<T, E>>,
        E: Into<DbError>,
    {
        let mut tx = self.pool.begin().await.map_err(DbError::Sqlx)?;
        let result = f(&mut tx).await;
        match result {
            Ok(value) => {
                tx.commit().await.map_err(DbError::Sqlx)?;
                Ok(value)
            }
            Err(error) => {
                let _ = tx.rollback().await;
                Err(error.into())
            }
        }
    }

    pub async fn close(self) {
        self.pool.close().await;
        if let Some(pool) = self.logging_pool {
            pool.close().await;
        }
        if let Some(pool) = self.replica_pool {
            pool.close().await;
        }
    }
}

/// `business_id = $business_id` for sea-query builders.
/// `hr_time_entries.owner_id` is the business ID (see `AGENTS.md`), so HR
/// queries scope on `owner_id` instead; every other scoped query uses
/// `business_id`.
pub fn business_scope_condition(column: &'static str, business_id: i64) -> Condition {
    Expr::col(column).eq(business_id).into_condition()
}

#[cfg(test)]
mod tests {
    use super::{business_scope_condition, resolve_main_dsn, PoolLimits};

    #[test]
    fn main_dsn_prefers_url_and_pool_limits() {
        let cfg = onion_core::config::AppConfig {
            database_url: "user:pw@tcp(localhost:3306)/onion_dev".to_string(),
            ..minimal_config()
        };
        assert_eq!(
            resolve_main_dsn(&cfg),
            "user:pw@tcp(localhost:3306)/onion_dev"
        );
        let limits = PoolLimits::default();
        assert_eq!(limits.max_connections, 25);
    }

    fn minimal_config() -> onion_core::config::AppConfig {
        let map = std::collections::HashMap::from([
            ("DB_HOST".to_string(), "localhost".to_string()),
            ("DB_USER".to_string(), "onion_user".to_string()),
            ("DB_NAME".to_string(), "onion_dev".to_string()),
            ("JWT_SECRET".to_string(), "test-secret".to_string()),
        ]);
        onion_core::config::load_and_validate(&|key| map.get(key).cloned()).unwrap()
    }

    #[test]
    fn tenant_condition_renders() {
        use sea_query::{backend::MysqlQueryBuilder, Query};
        let condition = business_scope_condition("business_id", 42);
        let sql = Query::select()
            .column(("t", "id"))
            .and_where(condition.into())
            .to_string(MysqlQueryBuilder);
        assert!(sql.contains("42"), "sql={sql}");
        assert!(sql.contains("business_id"), "sql={sql}");
    }

    #[tokio::test]
    async fn ping_needs_a_live_pool() {
        // No pool exists without DATABASE_URL pointing at a throwaway MySQL
        // from scripts/test-db.sh; connect is exercised only when configured.
        if std::env::var("DATABASE_URL").is_err() {
            return;
        }
        let map = std::collections::HashMap::from([
            (
                "DATABASE_URL".to_string(),
                std::env::var("DATABASE_URL").unwrap(),
            ),
            ("JWT_SECRET".to_string(), "test-secret".to_string()),
        ]);
        let cfg = onion_core::config::load_and_validate(&|key| map.get(key).cloned()).unwrap();
        let db = super::Database::connect_main(&cfg).await.unwrap();
        db.ping().await.unwrap();
        db.close().await;
    }
}
