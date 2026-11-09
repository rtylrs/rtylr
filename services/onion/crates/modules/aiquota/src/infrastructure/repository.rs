//! Sqlx repository over `crates/modules/aiquota/src/queries.rs`,
//! mirroring `internal/infrastructure/repository.go`. The pool comes
//! from the `db` crate's shared `Database` pool (tests run against a
//! throwaway MySQL via the local-env stack).

use crate::application::{Repository as RepositoryTrait, ServiceError};
use crate::domain::Mode;
use crate::queries;
use sqlx::Row as _;
use std::sync::Arc;

#[derive(Debug)]
pub struct Repository {
    pool: sqlx::MySqlPool,
}

impl Repository {
    pub fn new(pool: sqlx::MySqlPool) -> Self {
        Repository { pool }
    }

    pub fn from_database(database: Arc<db::pool::Database>) -> Self {
        let pool = database.pool().clone();
        Repository { pool }
    }
}

#[async_trait::async_trait]
impl RepositoryTrait for Repository {
    async fn get_mode(&self, business_id: i64) -> Result<Mode, ServiceError> {
        let row = queries::get_quota_mode(&self.pool, &queries::GetQuotaModeParams { business_id })
            .await
            .map_err(|error| ServiceError::Repository(error.to_string()))?;
        match row {
            Some(row) => {
                let mode: String = row
                    .try_get::<String, _>("mode")
                    .map_err(|error| ServiceError::Repository(error.to_string()))?;
                Ok(Mode::parse(&mode))
            }
            None => Ok(Mode::Pool),
        }
    }

    async fn set_mode(&self, business_id: i64, mode: Mode) -> Result<(), ServiceError> {
        queries::upsert_quota_mode(
            &self.pool,
            &queries::UpsertQuotaModeParams {
                business_id,
                mode: mode.as_str().to_string(),
            },
        )
        .await
        .map(|_| ())
        .map_err(|error| ServiceError::Repository(error.to_string()))
    }

    async fn get_used_tokens(
        &self,
        business_id: i64,
        bucket_user_id: i64,
        month: chrono::NaiveDate,
    ) -> Result<i64, ServiceError> {
        let row = queries::get_monthly_token_usage(
            &self.pool,
            &queries::GetMonthlyTokenUsageParams {
                business_id,
                user_id: bucket_user_id,
                billing_month: month,
            },
        )
        .await
        .map_err(|error| ServiceError::Repository(error.to_string()))?;
        match row {
            Some(row) => row
                .try_get::<i64, _>("tokens_used")
                .map_err(|error| ServiceError::Repository(error.to_string())),
            None => Ok(0),
        }
    }

    async fn add_usage(
        &self,
        business_id: i64,
        bucket_user_id: i64,
        month: chrono::NaiveDate,
        tokens: i64,
    ) -> Result<(), ServiceError> {
        queries::add_monthly_token_usage(
            &self.pool,
            &queries::AddMonthlyTokenUsageParams {
                business_id,
                user_id: bucket_user_id,
                billing_month: month,
                tokens_used: tokens,
            },
        )
        .await
        .map(|_| ())
        .map_err(|error| ServiceError::Repository(error.to_string()))
    }

    async fn reserve(
        &self,
        business_id: i64,
        bucket_user_id: i64,
        month: chrono::NaiveDate,
        estimate: i64,
        allowance: i64,
        allow_overage: bool,
    ) -> Result<bool, ServiceError> {
        // `EnsureUsageRow` first so the conditional update always
        // lands on an existing row (Go `Reserve`).
        queries::ensure_usage_row(
            &self.pool,
            &queries::EnsureUsageRowParams {
                business_id,
                user_id: bucket_user_id,
                billing_month: month,
            },
        )
        .await
        .map_err(|error| ServiceError::Repository(error.to_string()))?;
        let reserved = queries::reserve_monthly_tokens(
            &self.pool,
            &queries::ReserveMonthlyTokensParams {
                estimate,
                business_id,
                user_id: bucket_user_id,
                billing_month: month,
                allow_overage,
                allowance,
            },
        )
        .await
        .map_err(|error| ServiceError::Repository(error.to_string()))?;
        Ok(reserved.rows_affected() == 1)
    }

    async fn adjust(
        &self,
        business_id: i64,
        bucket_user_id: i64,
        month: chrono::NaiveDate,
        delta: i64,
    ) -> Result<(), ServiceError> {
        queries::adjust_monthly_token_usage(
            &self.pool,
            &queries::AdjustMonthlyTokenUsageParams {
                delta,
                business_id,
                user_id: bucket_user_id,
                billing_month: month,
            },
        )
        .await
        .map(|_| ())
        .map_err(|error| ServiceError::Repository(error.to_string()))
    }
}

#[cfg(test)]
mod tests {
    use super::Repository;
    use crate::application::Repository as _;

    fn test_pool_url() -> Option<String> {
        std::env::var("DATABASE_URL")
            .ok()
            .filter(|url| !url.is_empty())
    }

    async fn migrate_and_seed() -> Option<sqlx::MySqlPool> {
        let url = test_pool_url()?;
        let pool = sqlx::MySqlPool::connect(&url).await.ok()?;
        // Apply the module migrations from the Go tree (migrations
        // must not move; the Rust side only reads them).
        let manifest = concat!(
            env!("CARGO_MANIFEST_DIR"),
            "/../../../internal/modules/aiquota/sql/migration_20261025_business_ai_quota.sql"
        );
        let sql = std::fs::read_to_string(manifest).ok()?;
        for statement in sql.split(';') {
            let statement = statement.trim();
            if statement.is_empty() {
                continue;
            }
            let _ = sqlx::query(statement).execute(&pool).await.ok();
        }
        let manifest = concat!(
            env!("CARGO_MANIFEST_DIR"),
            "/../../../internal/modules/aiquota/sql/schema/shared/ai_usage.sql"
        );
        let sql = std::fs::read_to_string(manifest).ok()?;
        for statement in sql.split(';') {
            let statement = statement.trim();
            if statement.is_empty() {
                continue;
            }
            let _ = sqlx::query(statement).execute(&pool).await.ok();
        }
        Some(pool)
    }

    #[tokio::test]
    async fn repository_round_trip() {
        let Some(pool) = migrate_and_seed().await else {
            return;
        };
        let month = crate::application::month_start();
        let repo = Repository::new(pool);
        repo.set_mode(1, crate::domain::Mode::Pool).await.unwrap();
        assert_eq!(repo.get_mode(1).await.unwrap(), crate::domain::Mode::Pool);
        repo.reserve(1, 0, month, 4000, 100_000, false)
            .await
            .unwrap();
        repo.adjust(1, 0, month, -1000).await.unwrap();
        assert_eq!(repo.get_used_tokens(1, 0, month).await.unwrap(), 3000);
        // second reservation beyond allowance must not land
        let ok = repo.reserve(1, 0, month, 4000, 5000, false).await.unwrap();
        assert!(!ok);
        assert_eq!(repo.get_used_tokens(1, 0, month).await.unwrap(), 3000);
    }
}
