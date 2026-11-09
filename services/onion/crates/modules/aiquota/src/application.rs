//! Application ports and the quota service. Mirrors
//! `internal/application/service.go`.
//! AI tokens per AI credit (`platform/config.AITokensPerCredit`).
pub const AI_TOKENS_PER_CREDIT: i64 = 1000;

/// Optimistically debited before the provider call so N parallel
/// requests cannot all pass a nearly exhausted allowance. The
/// post-stream reconciliation credits back the unused portion.
pub const RESERVATION_ESTIMATE_TOKENS: i64 = 4000;

/// Subscription states that still grant plan AI credits; anything else
/// (cancelled, past_due, ...) falls back to the free tier.
const ACTIVE_ENTITLEMENT_STATUSES: [&str; 2] = ["active", "trialing"];

use crate::domain::{Entitlement, Mode, QuotaStatus};
use chrono::Datelike;
use thiserror::Error;

#[derive(Debug, Error, PartialEq)]
pub enum ServiceError {
    #[error("unknown business id")]
    UnknownBusiness,
    #[error("entitlement: {0}")]
    Entitlement(String),
    #[error("repository: {0}")]
    Repository(String),
}

/// What the service needs from the tenant scope (Go
/// `tenant.ActiveBusinessID`).
pub trait TenantResolver: Send + Sync {
    fn active_business_id(&self, business_id: i64) -> Result<i64, ServiceError>;
}

/// Quota persistence port (Go `repository` interface).
#[async_trait::async_trait]
pub trait Repository: Send + Sync {
    async fn get_mode(&self, business_id: i64) -> Result<Mode, ServiceError>;
    async fn set_mode(&self, business_id: i64, mode: Mode) -> Result<(), ServiceError>;
    async fn get_used_tokens(
        &self,
        business_id: i64,
        bucket_user_id: i64,
        month: chrono::NaiveDate,
    ) -> Result<i64, ServiceError>;
    async fn add_usage(
        &self,
        business_id: i64,
        bucket_user_id: i64,
        month: chrono::NaiveDate,
        tokens: i64,
    ) -> Result<(), ServiceError>;
    async fn reserve(
        &self,
        business_id: i64,
        bucket_user_id: i64,
        month: chrono::NaiveDate,
        estimate: i64,
        allowance: i64,
        allow_overage: bool,
    ) -> Result<bool, ServiceError>;
    async fn adjust(
        &self,
        business_id: i64,
        bucket_user_id: i64,
        month: chrono::NaiveDate,
        delta: i64,
    ) -> Result<(), ServiceError>;
}

/// What the service needs from the subscriptions module
/// (Go `entitlementProvider`).
#[async_trait::async_trait]
pub trait EntitlementProvider: Send + Sync {
    async fn quota_entitlement(
        &self,
        business_id: i64,
    ) -> Result<(String, i32, i32, String), ServiceError>;
    async fn free_plan_ai_credits(&self) -> Result<i64, ServiceError>;
}

/// The quota service: reserve → allow → reconcile.
pub struct Service<R: Repository, E: EntitlementProvider, T: TenantResolver> {
    repository: R,
    entitlements: E,
    tenants: T,
}

impl<R: Repository, E: EntitlementProvider, T: TenantResolver> Service<R, E, T> {
    pub fn new(repository: R, entitlements: E, tenants: T) -> Self {
        Service {
            repository,
            entitlements,
            tenants,
        }
    }

    /// Authorize: reserve the estimate atomically and report whether
    /// the caller may make an AI request. Failing closed: when the
    /// entitlement cannot be read the request is blocked; when only
    /// usage cannot be read, overage plans keep serving and capped
    /// plans are blocked.
    pub async fn reserve(&self, business_id: i64, user_id: i64) -> (bool, QuotaStatus, i64) {
        let business = match self.tenants.active_business_id(business_id) {
            Ok(business) => business,
            Err(_) => {
                return (
                    false,
                    QuotaStatus {
                        blocked: true,
                        ..Default::default()
                    },
                    0,
                )
            }
        };
        let entitlement = match self.entitlement(business).await {
            Ok(entitlement) => entitlement,
            Err(_) => {
                return (
                    false,
                    QuotaStatus {
                        blocked: true,
                        ..Default::default()
                    },
                    0,
                )
            }
        };
        let mode = self
            .repository
            .get_mode(business)
            .await
            .unwrap_or(Mode::Pool);
        let (bucket_user, allowance) = bucket_and_allowance(&entitlement, mode, user_id);
        let used = match self
            .repository
            .get_used_tokens(business, bucket_user, month_start())
            .await
        {
            Ok(used) => used,
            Err(_) => {
                if allows_overage(&entitlement.plan_slug) {
                    return (true, QuotaStatus::default(), 0);
                }
                return (false, QuotaStatus::default(), 0);
            }
        };
        let overage_allowed = allows_overage(&entitlement.plan_slug);
        if !overage_allowed && used >= allowance {
            return (
                false,
                build_status(&entitlement, mode, allowance, used),
                RESERVATION_ESTIMATE_TOKENS,
            );
        }
        let reserved = match self
            .repository
            .reserve(
                business,
                bucket_user,
                month_start(),
                RESERVATION_ESTIMATE_TOKENS,
                allowance,
                overage_allowed,
            )
            .await
        {
            Ok(reserved) => reserved,
            Err(_) => {
                if overage_allowed {
                    return (true, QuotaStatus::default(), 0);
                }
                return (false, QuotaStatus::default(), 0);
            }
        };
        if !reserved {
            return (
                false,
                build_status(&entitlement, mode, allowance, used),
                RESERVATION_ESTIMATE_TOKENS,
            );
        }
        let used = used + RESERVATION_ESTIMATE_TOKENS;
        (
            true,
            build_status(&entitlement, mode, allowance, used),
            RESERVATION_ESTIMATE_TOKENS,
        )
    }

    /// Reconcile a prior reservation against actual usage. Writes are
    /// retried with backoff because a lost reconciliation permanently
    /// under-counts usage.
    pub async fn reconcile(&self, business_id: i64, user_id: i64, actual: i64, reserved: i64) {
        let delta = actual - reserved;
        if delta == 0 {
            return;
        }
        let business = match self.tenants.active_business_id(business_id) {
            Ok(business) => business,
            Err(_) => return,
        };
        let mode = self
            .repository
            .get_mode(business)
            .await
            .unwrap_or(Mode::Pool);
        let bucket = if mode == Mode::PerPerson { user_id } else { 0 };
        const ATTEMPTS: usize = 3;
        for i in 0..ATTEMPTS {
            if self
                .repository
                .adjust(business, bucket, month_start(), delta)
                .await
                .is_ok()
            {
                return;
            }
            if i == ATTEMPTS - 1 {
                return;
            }
            tokio::time::sleep(std::time::Duration::from_millis(250 * (i as u64 + 1))).await;
        }
    }

    /// Status view.
    pub async fn status(
        &self,
        business_id: i64,
        user_id: i64,
    ) -> Result<QuotaStatus, ServiceError> {
        let business = self.tenants.active_business_id(business_id)?;
        let entitlement = self.entitlement(business).await?;
        let mode = self.repository.get_mode(business).await?;
        let (bucket_user, allowance) = bucket_and_allowance(&entitlement, mode, user_id);
        let used = self
            .repository
            .get_used_tokens(business, bucket_user, month_start())
            .await?;
        Ok(build_status(&entitlement, mode, allowance, used))
    }

    pub async fn get_mode(&self, business_id: i64) -> Result<Mode, ServiceError> {
        let business = self.tenants.active_business_id(business_id)?;
        self.repository.get_mode(business).await
    }

    pub async fn set_mode(&self, business_id: i64, mode: Mode) -> Result<(), ServiceError> {
        let business = self.tenants.active_business_id(business_id)?;
        self.repository.set_mode(business, mode.normalize()).await
    }

    async fn entitlement(&self, business: i64) -> Result<Entitlement, ServiceError> {
        let (plan_slug, credits, seats, status) =
            self.entitlements.quota_entitlement(business).await?;
        let mut entitlement = Entitlement {
            plan_slug,
            ai_credits_per_month: credits,
            seat_count: seats,
            status,
        };
        if !ACTIVE_ENTITLEMENT_STATUSES.contains(&entitlement.status.as_str())
            && entitlement.status != "none"
        {
            if let Ok(credits) = self.entitlements.free_plan_ai_credits().await {
                entitlement = Entitlement {
                    plan_slug: "free".to_string(),
                    ai_credits_per_month: credits as i32,
                    seat_count: entitlement.seat_count,
                    status: "free".to_string(),
                };
            }
        }
        Ok(entitlement)
    }
}

fn per_seat_bonus_credits(slug: &str) -> i32 {
    match slug {
        "pro" => 500,
        "scale" => 1000,
        _ => 0,
    }
}

fn allows_overage(slug: &str) -> bool {
    slug == "pro" || slug == "scale"
}

fn bucket_and_allowance(entitlement: &Entitlement, mode: Mode, user_id: i64) -> (i64, i64) {
    let seats = entitlement.seat_count.max(1);
    let team_credits =
        entitlement.ai_credits_per_month + per_seat_bonus_credits(&entitlement.plan_slug) * seats;
    let team_allowance = team_credits as i64 * AI_TOKENS_PER_CREDIT;
    if mode == Mode::PerPerson {
        return (user_id, team_allowance / seats as i64);
    }
    (0, team_allowance)
}

fn build_status(entitlement: &Entitlement, mode: Mode, allowance: i64, used: i64) -> QuotaStatus {
    let remaining = (allowance - used).max(0);
    let overage = (used - allowance).max(0);
    QuotaStatus {
        plan_slug: entitlement.plan_slug.clone(),
        mode,
        allowance_tokens: allowance,
        used_tokens: used,
        remaining_tokens: remaining,
        overage_tokens: overage,
        allow_overage: allows_overage(&entitlement.plan_slug),
        blocked: !allows_overage(&entitlement.plan_slug) && used >= allowance,
        period_end: month_end(),
    }
}

/// Go `monthStart()`: first day of the current UTC month.
pub fn month_start() -> chrono::NaiveDate {
    chrono::Utc::now()
        .date_naive()
        .with_day(1)
        .unwrap_or(chrono::NaiveDate::MIN)
}

/// Go `monthEnd()`: first day of next month, as UTC.
pub fn month_end() -> chrono::DateTime<chrono::Utc> {
    let now = chrono::Utc::now();
    let (year, month) = (now.year(), now.month());
    let (next_year, next_month) = if month == 12 {
        (year + 1, 1)
    } else {
        (year, month + 1)
    };
    let date =
        chrono::NaiveDate::from_ymd_opt(next_year, next_month, 1).unwrap_or(chrono::NaiveDate::MIN);
    date.and_hms_opt(0, 0, 0)
        .map(|naive| naive.and_utc())
        .unwrap_or(chrono::DateTime::<chrono::Utc>::MIN_UTC)
}

#[cfg(test)]
mod tests {
    use super::{allows_overage, bucket_and_allowance, month_start};
    use crate::domain::{Entitlement, Mode};
    use chrono::Datelike;

    #[test]
    fn money_math() {
        let ent = Entitlement {
            plan_slug: "scale".into(),
            ai_credits_per_month: 100,
            seat_count: 4,
            status: "active".into(),
        };
        let (bucket, allowance) = bucket_and_allowance(&ent, Mode::Pool, 9);
        assert_eq!(bucket, 0);
        assert_eq!(allowance, (100 + 1000 * 4) * 1000);
        let (bucket, allowance) = bucket_and_allowance(&ent, Mode::PerPerson, 9);
        assert_eq!(bucket, 9);
        assert_eq!(allowance, (100 + 1000 * 4) * 1000 / 4);
        assert!(allows_overage("pro"));
        assert!(!allows_overage("free"));
        assert_eq!(month_start().day(), 1);
    }
}
