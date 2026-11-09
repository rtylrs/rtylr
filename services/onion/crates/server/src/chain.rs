//! Middleware chain order, health/root payloads, and server timeouts.
//! Mirrors `web_server.go`.

/// Graceful shutdown window in seconds (`serverShutdownTimeoutSeconds`).
pub const SHUTDOWN_TIMEOUT_SECS: u64 = 20;

/// Read/write/read-header/idle timeouts in seconds
/// (`http.Server` fields).
pub const SERVER_TIMEOUTS_SECS: (u64, u64, u64, u64) = (30, 30, 10, 120);
/// Max request header bytes (`1 << 20`, 1 MiB).
pub const MAX_HEADER_BYTES: usize = 1 << 20;

/// The middleware chain in platform order, outermost first
/// (`platformMiddlewares`). Sentries and tracing hook in here later.
/// `accessLogs` is `safeRequestLogger`, present only when
/// `HTTP_ACCESS_LOGS`/`ACCESS_LOGS` is on.
pub fn middleware_chain(access_logs: bool) -> Vec<&'static str> {
    let mut chain = Vec::with_capacity(16);
    if access_logs {
        chain.push("safe-request-logger");
    }
    chain.extend([
        "recover",
        "sentry-recover",
        "cors",
        "crash-notifier",
        "security-headers",
        "secure-error",
        "response-compress",
        "request-body-limit",
        "rate-limit",
        "auth-route-rate-limit",
        "request-timeout",
        "http-metrics",
        "flow-v1-usage",
    ]);
    chain
}

/// `/health` body when the world is healthy.
pub const HEALTH_OK: &str = "{\"status\":\"ok\"}";
/// `/health` body when the database ping fails.
pub const HEALTH_UNHEALTHY: &str = "{\"status\":\"unhealthy\"}";
/// `GET /{$}` root payload.
pub const ROOT_OK: &str = "{\"message\":\"Why are you here?\",\"status\":\"ok\"}";

/// Whether the run mode needs a database health check (default: yes;
/// the `DatabaseHealthMode` override lives with the run mode in
/// `bins/api`, unit 1.9).
pub fn uses_database_health_check() -> bool {
    true
}

#[cfg(test)]
mod tests {
    use super::{middleware_chain, HEALTH_OK, HEALTH_UNHEALTHY, ROOT_OK};

    #[test]
    fn chain_order() {
        let without = middleware_chain(false);
        let with = middleware_chain(true);
        assert_eq!(without[0], "recover");
        assert_eq!(with[0], "safe-request-logger");
        assert_eq!(with.len(), without.len() + 1);
        // Partials bypass wraps this whole chain: the global guardrail
        // is the 1000/s rate limiter, not per-route.
        assert!(without.contains(&"rate-limit"));
        assert!(without.contains(&"request-timeout"));
    }

    #[test]
    fn payloads() {
        assert_eq!(HEALTH_OK, "{\"status\":\"ok\"}");
        assert_eq!(HEALTH_UNHEALTHY, "{\"status\":\"unhealthy\"}");
        assert_eq!(
            ROOT_OK,
            "{\"message\":\"Why are you here?\",\"status\":\"ok\"}"
        );
    }
}
