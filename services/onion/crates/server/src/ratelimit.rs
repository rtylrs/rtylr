//! Sliding-window rate limiter, per-IP client keys, and the tight
//! budgets on anonymous auth routes. Mirrors `rate_limit.go` and
//! `meta_rate_limit.go`.
//!
//! Limiter lifecycle: `spawn_cleanup` returns a handle that is the
//! Stop — aborting it ends the cleanup loop (Go `Stop()` cancels the
//! ticker goroutine). The composition root (`bins/api`) keeps the
//! limiters in an `Arc` for the server lifetime.

use std::collections::HashMap;
use std::sync::Mutex;
use std::time::{Duration, Instant, SystemTime};
use thiserror::Error;

/// Global per-IP guardrail: same bucket shape as Go
/// `RateLimitMiddleware(1000, time.Second)`.
pub const GLOBAL_LIMIT: usize = 1000;
pub const GLOBAL_WINDOW_SECS: u64 = 1;

/// JSON sent with every 429.
pub const RATE_LIMITED_BODY: &str =
    "{\"error\":\"Rate limit exceeded\",\"message\":\"Too many requests, please try again later\",\"retry_after\":60}";
/// Per-route 429 body.
pub const PER_ROUTE_LIMITED_BODY: &str =
    "{\"error\":\"Rate limit exceeded\",\"message\":\"Too many attempts, please try again shortly\",\"retry_after\":60}";
/// `Retry-After` header value on every 429.
pub const RETRY_AFTER_HEADER: &str = "60";

#[derive(Debug, Error, PartialEq)]
pub enum RateLimitError {
    #[error("rate limit {0:?}: want <count>/<s|m|h>")]
    BadSpec(String),
}

/// Sliding-window limiter keyed by any string (IP, path|IP, path|user).
/// Buckets are in-process; with N replicas the effective budget is
/// limit × N, matching Go's stated limitation.
pub struct SecurityRateLimiter {
    limit: usize,
    window: Duration,
    requests: Mutex<HashMap<String, Vec<Instant>>>,
}

impl std::fmt::Debug for SecurityRateLimiter {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        f.debug_struct("SecurityRateLimiter")
            .field("limit", &self.limit)
            .field("window", &self.window)
            .finish()
    }
}

impl SecurityRateLimiter {
    pub fn new(limit: usize, window: Duration) -> Self {
        Self {
            limit,
            window,
            requests: Mutex::new(HashMap::new()),
        }
    }

    pub fn limit(&self) -> usize {
        self.limit
    }

    pub fn window(&self) -> Duration {
        self.window
    }

    /// Insert `now` into `key`'s bucket and report whether the caller
    /// is inside the budget.
    pub fn is_allowed_at(&self, key: &str, now: Instant) -> bool {
        let mut requests = self.requests.lock().unwrap_or_else(|err| err.into_inner());
        let window_start = now.checked_sub(self.window).unwrap_or(now);
        let entry = requests.entry(key.to_string()).or_default();
        retain_after(entry, window_start);
        if entry.len() >= self.limit {
            return false;
        }
        entry.push(now);
        true
    }

    /// Allowed with the wall clock (Go `time.Now`).
    pub fn is_allowed(&self, key: &str) -> bool {
        self.is_allowed_at(key, Instant::now())
    }

    /// Drop expired entries (Go periodic `cleanup` ticker will call
    /// this; keeping it synchronous makes it testable).
    pub fn cleanup_at(&self, now: Instant) {
        let mut requests = self.requests.lock().unwrap_or_else(|err| err.into_inner());
        let window_start = now.checked_sub(self.window).unwrap_or(now);
        requests.retain(|_, times| {
            retain_after(times, window_start);
            !times.is_empty()
        });
    }

    /// Spawn the cleanup loop (Go `cleanup()` ticker).
    pub fn spawn_cleanup(self: std::sync::Arc<Self>) -> tokio::task::JoinHandle<()> {
        let limiter = self.clone();
        tokio::spawn(async move {
            let mut interval = tokio::time::interval(limiter.window);
            loop {
                interval.tick().await;
                limiter.cleanup_at(Instant::now());
            }
        })
    }
}

fn retain_after(times: &mut Vec<Instant>, cutoff: Instant) {
    times.retain(|time| *time > cutoff);
}

/// Strip the port from `remote_addr` (Go `peerIP`).
pub fn peer_ip(remote_addr: &str) -> &str {
    let remote_addr = remote_addr.trim();
    if remote_addr.starts_with('[') {
        // `[host]:port` — the host itself contains `:`.
        if let Some(end) = remote_addr.find(']') {
            return &remote_addr[1..end];
        }
        return remote_addr;
    }
    // Bare IPv6 literal (no port) — do not split on the last `:`.
    if remote_addr.parse::<std::net::IpAddr>().is_ok() {
        return remote_addr;
    }
    remote_addr
        .rsplit_once(':')
        .filter(|(host, port)| !host.is_empty() && port.chars().all(|c| c.is_ascii_digit()))
        .map(|(host, _)| host)
        .unwrap_or(remote_addr)
}

/// Segment-wise `{param}` pattern match (Go `routePatternMatchesPath`).
pub fn route_pattern_matches_path(pattern: &str, path: &str) -> bool {
    let pattern_segments: Vec<&str> = pattern.split('/').collect();
    let path_segments: Vec<&str> = path.split('/').collect();
    if pattern_segments.len() != path_segments.len() {
        return false;
    }
    pattern_segments
        .iter()
        .zip(path_segments.iter())
        .all(|(pat, segment)| (pat.starts_with('{') && pat.ends_with('}')) || pat == segment)
}

/// Per-client key for rate limiting (Go `getClientKey`): the socket
/// peer IP, honoring `CF-Connecting-IP` only when the peer is in the
/// configured trusted-proxy CIDR set. `X-Forwarded-For`/`X-Real-IP`
/// are deliberately not honored.
pub fn client_key(
    remote_addr: &str,
    cf_connecting_ip: Option<&str>,
    trusted_cidrs: &[String],
) -> String {
    let peer = peer_ip(remote_addr);
    if onion_core::config::is_trusted_proxy_ip(remote_addr, trusted_cidrs) {
        if let Some(cf) = cf_connecting_ip.map(str::trim) {
            if !cf.is_empty() && cf.parse::<std::net::IpAddr>().is_ok() {
                return cf.to_string();
            }
        }
    }
    peer.to_string()
}

/// Per-route bucket key (path | client). AI route variant keys on the
/// authenticated user instead: `path|user:<id>`.
pub fn route_key(path: &str, client: &str) -> String {
    format!("{path}|{client}")
}

/// AI-user key when authenticated (Go `AIUserRateLimit`).
pub fn ai_user_key(path: &str, user_id: Option<i64>, client: &str) -> String {
    match user_id.filter(|id| *id > 0) {
        Some(id) => format!("{path}|user:{id}"),
        None => route_key(path, client),
    }
}

/// Tight per-IP budget for the anonymous auth routes (Go
/// `authRouteLimiters`, limit per minute).
pub fn auth_route_budget(path: &str) -> Option<usize> {
    for (pattern, budget) in AUTH_ROUTE_BUDGETS {
        if route_pattern_matches_path(pattern, path) {
            return Some(budget);
        }
    }
    None
}

/// Exact table of per-route budgets (Go `authRouteLimiters`). `{param}`
/// segments match any single path segment (Go
/// `routePatternMatchesPath` runtime fallback).
pub const AUTH_ROUTE_BUDGETS: [(&str, usize); 24] = [
    // --- account credentials ---
    ("/common/auth/v1/login", 5),
    ("/common/auth/v1/register", 3),
    ("/common/auth/v1/register/google", 3),
    ("/common/auth/v1/refresh", 30),
    ("/common/auth/v1/revoke", 30),
    ("/common/auth/v1/password-reset/request", 5),
    ("/common/auth/v1/password-reset/confirm", 5),
    ("/common/auth/v1/invite/accept", 5),
    ("/hr/employees/invite/accept", 5),
    ("/hr/employees/invite/verify", 5),
    // OAuth endpoints are anonymous (256-bit codes; brake on replay).
    ("/common/oauth/v1/token", 30),
    ("/common/oauth/v1/revoke", 30),
    ("/oauth/authorize", 30),
    // Re-proves the current password with bcrypt on every call.
    ("/hub/user/v1/users/password", 5),
    // --- short secrets, sized to their search space ---
    ("/common/auth/v1/verify-email", 3),
    ("/pos/user/v1/users/verify-pin", 5),
    ("/pos/manager/v1/verify", 5),
    ("/hub/security/v1/mfa/verify", 10),
    ("/hub/security/v1/mfa/recovery", 10),
    // --- anonymous oracles, not brute force ---
    ("/common/auth/v1/email-exists", 20),
    ("/recruit/public/resume-upload-url", 10),
    // --- public read surfaces (unchanged) ---
    ("/pos/orders/v1/orders/public", 10),
    ("/pos/orders/v1/orders/{order_id}/public-status", 60),
    ("/pos/menu-item/v1/menu-items/public", 60),
];

/// Parse a `<count>/<s|m|h>` budget (Go `ParseRouteRateLimit`).
pub fn parse_route_rate_limit(spec: &str) -> Result<(usize, Duration), RateLimitError> {
    let (count, unit) = spec
        .trim()
        .split_once('/')
        .ok_or_else(|| RateLimitError::BadSpec(spec.to_string()))?;
    let limit: usize = count
        .parse()
        .map_err(|_| RateLimitError::BadSpec(spec.to_string()))?;
    if limit == 0 {
        return Err(RateLimitError::BadSpec(spec.to_string()));
    }
    let window = match unit {
        "s" => Duration::from_secs(1),
        "m" => Duration::from_secs(60),
        "h" => Duration::from_secs(3600),
        _ => return Err(RateLimitError::BadSpec(spec.to_string())),
    };
    Ok((limit, window))
}

pub fn now_wallclock() -> SystemTime {
    SystemTime::now()
}

#[cfg(test)]
mod tests {
    use super::{
        ai_user_key, auth_route_budget, client_key, parse_route_rate_limit, peer_ip, route_key,
        SecurityRateLimiter, PER_ROUTE_LIMITED_BODY, RATE_LIMITED_BODY,
    };
    use std::time::{Duration, Instant};

    #[test]
    fn sliding_window() {
        let limiter = SecurityRateLimiter::new(3, Duration::from_secs(60));
        let now = Instant::now();
        assert!(limiter.is_allowed_at("ip", now));
        assert!(limiter.is_allowed_at("ip", now));
        assert!(limiter.is_allowed_at("ip", now));
        assert!(!limiter.is_allowed_at("ip", now));
        // Other buckets are independent.
        assert!(limiter.is_allowed_at("other", now));
    }

    #[test]
    fn window_expiry_and_cleanup() {
        let limiter = SecurityRateLimiter::new(1, Duration::from_secs(60));
        let start = Instant::now();
        assert!(limiter.is_allowed_at("ip", start));
        let later = start + Duration::from_secs(61);
        assert!(limiter.is_allowed_at("ip", later));
        limiter.cleanup_at(later + Duration::from_secs(61));
        assert!(limiter.is_allowed_at("ip", later + Duration::from_secs(62)));
    }

    #[test]
    fn client_key_rules() {
        assert_eq!(peer_ip("203.0.113.7:443"), "203.0.113.7");
        assert_eq!(peer_ip("[2001:db8::7]:443"), "2001:db8::7");
        assert_eq!(peer_ip("2001:db8::7"), "2001:db8::7");
        assert_eq!(
            client_key("203.0.113.7:443", Some("1.2.3.4"), &[]),
            "203.0.113.7"
        );
        let cidrs = vec!["203.0.113.0/24".to_string()];
        assert_eq!(
            client_key("203.0.113.7:443", Some("1.2.3.4"), &cidrs),
            "1.2.3.4"
        );
        assert_eq!(
            client_key("203.0.113.7:443", Some("not-an-ip"), &cidrs),
            "203.0.113.7"
        );
    }

    #[test]
    fn budgets_and_spec() {
        assert_eq!(auth_route_budget("/common/auth/v1/login"), Some(5));
        assert_eq!(auth_route_budget("/common/auth/v1/verify-email"), Some(3));
        assert_eq!(auth_route_budget("/hub/security/v1/mfa/verify"), Some(10));
        assert_eq!(
            auth_route_budget("/pos/orders/v1/orders/ORD-9/public-status"),
            Some(60)
        );
        assert_eq!(auth_route_budget("/pos/orders/v1/orders/public"), Some(10));
        assert_eq!(auth_route_budget("/unknown"), None);
        assert_eq!(
            parse_route_rate_limit("10/m"),
            Ok((10, Duration::from_secs(60)))
        );
        assert_eq!(
            parse_route_rate_limit("0/s"),
            Err(super::RateLimitError::BadSpec("0/s".into()))
        );
        assert_eq!(
            parse_route_rate_limit("5/q"),
            Err(super::RateLimitError::BadSpec("5/q".into()))
        );
        assert_eq!(route_key("/p", "ip"), "/p|ip");
        assert_eq!(ai_user_key("/p", Some(7), "ip"), "/p|user:7");
        assert_eq!(ai_user_key("/p", None, "ip"), "/p|ip");
        assert!(RATE_LIMITED_BODY.contains("retry_after"));
        assert!(PER_ROUTE_LIMITED_BODY.contains("shortly"));
    }
}
