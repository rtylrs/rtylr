//! Application state: `Arc`'d services built once in `bins/api`.
//! Mirrors the process wiring in `cmd/api/main.go` (unit 1.9). Modules
//! receive `&AppState` at registration; new services are added here as
//! their crates land. No global singletons: every value below is held
//! by the state value itself.

use std::collections::HashMap;
use std::sync::Arc;
use std::time::Duration;

use crate::ratelimit::{SecurityRateLimiter, AUTH_ROUTE_BUDGETS, GLOBAL_LIMIT, GLOBAL_WINDOW_SECS};
use crate::version::BuildInfo;
use db::pool::{Database, DbError};
use onion_core::config::{self, AppConfig, EnvGet};
use realtime::Hub;

/// Shared application state.
pub struct AppState {
    pub config: Arc<AppConfig>,
    /// Mandatory database pool. `bins` abort at boot when connect
    /// fails, so a running process always has one (the health
    /// handler reports 503 only if the ping times out).
    pub db: Option<Database>,
    /// `None` only when the mandatory pool failed to connect; bins
    /// exit non-zero in that case, so a running process always has a
    /// pool. Kept `Option` so the probe ready check can report it.
    pub wsbus: Arc<Hub>,
    /// Global per-IP guardrail (1000/s), built once per process.
    pub global_rate_limiter: Arc<SecurityRateLimiter>,
    /// Tight per-route limiters (login/register/MFA/offers etc.).
    pub auth_route_limiters: Arc<HashMap<String, Arc<SecurityRateLimiter>>>,
    /// Partials internal key (`PARTIALS_API_KEY`, empty = bypass off).
    pub partials_api_key: String,
    /// Trusted proxy CIDRs for rate-limit client keys.
    pub trusted_proxy_cidrs: Vec<String>,
    /// Build identity injected at build time.
    pub build: BuildInfo,
    /// Resend client API key (empty = email disabled).
    pub resend_api_key: String,
    /// Limiter cleanup handles: dropping or aborting these stops
    /// the bucket cleanup (Go `Stop()` on the limiter ticker).
    _cleanup_handles: Vec<tokio::task::JoinHandle<()>>,
    /// Realtime topics require no auth until the module authorizer
    /// lands in Phase 2; the default is deny-all (fail-closed).
    pub hub_authorizer_note: &'static str,
}

impl AppState {
    /// Build the process-scoped services. Mirrors Go's boot order:
    /// config already validated, the database pool connects, and a
    /// failed connect aborts startup (Go returns the error to the
    /// restart loop in `main`).
    pub async fn new(
        get: EnvGet<'_>,
        config: AppConfig,
        build: BuildInfo,
    ) -> Result<Self, DbError> {
        let db = Database::connect_main(&config).await?;
        let partials_api_key = config.partials_api_key.clone();
        let trusted_proxy_cidrs = config::get_trusted_proxy_cidrs(get);
        let resend_api_key = config.resend_key.clone();
        let mut auth_route_limiters = HashMap::new();
        let mut cleanup_handles = Vec::new();
        for (route, budget) in AUTH_ROUTE_BUDGETS {
            let limiter = Arc::new(SecurityRateLimiter::new(budget, Duration::from_secs(60)));
            cleanup_handles.push(limiter.clone().spawn_cleanup());
            auth_route_limiters.insert(route.to_string(), limiter);
        }
        let global_rate_limiter = Arc::new(SecurityRateLimiter::new(
            GLOBAL_LIMIT,
            Duration::from_secs(GLOBAL_WINDOW_SECS),
        ));
        let global_cleanup = global_rate_limiter.clone().spawn_cleanup();
        cleanup_handles.push(global_cleanup);
        Ok(AppState {
            config: Arc::new(config),
            db: Some(db),
            wsbus: Arc::new(Hub::new(std::sync::Arc::new(|_, _| false))),
            global_rate_limiter,
            auth_route_limiters: Arc::new(auth_route_limiters),
            partials_api_key,
            trusted_proxy_cidrs,
            build,
            resend_api_key,
            _cleanup_handles: cleanup_handles,
            hub_authorizer_note:
                "realtime authorizer required before ws routes (deny-all until Phase 2)",
        })
    }

    /// Global rate-limit decision for one client key.
    pub fn rate_limited(&self, key: &str) -> bool {
        !self.global_rate_limiter.is_allowed(key)
    }

    /// Partials bypass decision for one request.
    pub fn is_partials_request(&self, header_value: Option<&str>) -> bool {
        crate::partials::has_valid_partials_api_key(header_value, &self.partials_api_key)
    }
}

#[cfg(test)]
mod tests {
    use super::{AUTH_ROUTE_BUDGETS, GLOBAL_LIMIT};

    #[test]
    fn budgets_cover_global() {
        assert_eq!(GLOBAL_LIMIT, 1000);
        assert_eq!(AUTH_ROUTE_BUDGETS.len(), 24);
    }
}
