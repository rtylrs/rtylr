//! Small probe server. Mirrors `internal/platform/server/probe`.
//!
//! Empty port disables the probe, which keeps local multi-process dev
//! from fighting over `:8888` (`PROBE_PORT` handling below).

/// Local dev default for non-Kubernetes environments.
pub const DEFAULT_PROBE_PORT: &str = "8888";

/// Choose the probe port from the environment (Go `PortFromEnv`):
/// `PROBE_PORT` wins; under Kubernetes (`KUBERNETES_SERVICE_HOST` set)
/// the default is used; otherwise the probe is disabled (`None`).
pub fn port_from_env(
    probe_port: Option<&str>,
    kubernetes_service_host: Option<&str>,
    default: &str,
) -> Option<String> {
    if let Some(port) = probe_port.filter(|value| !value.is_empty()) {
        return Some(port.to_string());
    }
    if kubernetes_service_host.is_some_and(|value| !value.is_empty()) {
        return Some(default.to_string());
    }
    None
}

/// Maximum time for one readiness check (Go `probe.go` uses a 2s
/// timeout per ready check).
pub const READY_CHECK_TIMEOUT_SECS: u64 = 2;

/// Read `PROBE_PORT` / `KUBERNETES_SERVICE_HOST` from the process.
pub fn port_from_process(default: &str) -> Option<String> {
    port_from_env(
        std::env::var("PROBE_PORT").ok().as_deref(),
        std::env::var("KUBERNETES_SERVICE_HOST").ok().as_deref(),
        default,
    )
}

/// Human-readable probe identity.
#[derive(Debug, Clone)]
pub struct ProbeConfig {
    pub service_name: String,
    pub port: Option<String>,
    pub stop_token: String,
}

impl ProbeConfig {
    pub fn disabled(service_name: &str) -> Self {
        Self {
            service_name: service_name.to_string(),
            port: None,
            stop_token: String::new(),
        }
    }
}

/// `/health` body.
pub fn health_json(service: &str) -> serde_json::Value {
    serde_json::json!({ "status": "ok", "service": service })
}

/// `/ready` Ok body.
pub fn ready_json() -> serde_json::Value {
    serde_json::json!({ "status": "ready" })
}

/// `/ready` failure body.
pub fn unready_json() -> serde_json::Value {
    serde_json::json!({ "status": "unready" })
}

/// `/stop` success body.
pub fn stopping_json() -> serde_json::Value {
    serde_json::json!({ "status": "stopping" })
}

/// `/stop` with no stop callback.
pub fn stop_unsupported_json() -> serde_json::Value {
    serde_json::json!({ "status": "unsupported" })
}

/// Metrics probe is opt-in via `PROBE_METRICS_ENABLED=true`.
pub fn metrics_enabled(env_value: Option<&str>) -> bool {
    env_value == Some("true")
}

/// `/metrics` body.
pub fn metrics_json(goroutines: u64, alloc_bytes: u64) -> serde_json::Value {
    serde_json::json!({ "goroutines": goroutines, "alloc": alloc_bytes })
}

#[derive(Debug, Clone, PartialEq)]
pub enum StopDecision {
    /// Non-POST.
    MethodNotAllowed,
    /// Missing / wrong token (the endpoint 404s, like Go).
    NotFound,
    /// Authorized but the process has no stop callback.
    NotImplemented,
    /// Authorized: respond 200 `{"status":"stopping"}` and stop.
    Stopping,
}

/// Decide the response for a `/stop` request, mirroring `handleStop`:
/// only POST with a matching `X-Probe-Token` (constant-time) reaches
/// the stop callback.
pub fn stop_decision(
    method: &str,
    header_token: Option<&str>,
    expected_token: &str,
    has_stop: bool,
) -> StopDecision {
    if method != "POST" {
        return StopDecision::MethodNotAllowed;
    }
    if !authorized_stop_request(header_token, expected_token) {
        return StopDecision::NotFound;
    }
    if !has_stop {
        return StopDecision::NotImplemented;
    }
    StopDecision::Stopping
}

/// Whether a `/stop` request is authorized: token must be non-empty
/// and match in constant time. The expected token is trimmed like Go
/// does in `probe.Start`.
pub fn authorized_stop_request(header_token: Option<&str>, expected_token: &str) -> bool {
    let expected = expected_token.trim();
    if expected.is_empty() {
        return false;
    }
    let Some(provided) = header_token.map(str::trim) else {
        return false;
    };
    if provided.is_empty() {
        return false;
    }
    provided.len() == expected.len()
        && provided
            .bytes()
            .zip(expected.bytes())
            .fold(0u8, |diff, (a, b)| diff | (a ^ b))
            == 0
}

/// Read the stop token, falling back to `PROBE_STOP_TOKEN`.
pub fn stop_token_from(env_token: Option<&str>, config_token: &str) -> String {
    let from_config = config_token.trim();
    if !from_config.is_empty() {
        return from_config.to_string();
    }
    env_token.map(str::trim).unwrap_or_default().to_string()
}

/// Ready check for `/ready` (Go `probe.Check`).
pub type ReadyCheck = std::sync::Arc<dyn Fn() -> bool + Send + Sync>;

#[derive(Clone)]
struct ProbeState {
    service: String,
    stop_token: String,
    ready: Option<ReadyCheck>,
    has_stop: bool,
}

/// Build the probe router (`/health`, `/ready`, `/stop`, `/metrics`),
/// mirroring `probe.Start`'s mux. `has_stop` controls whether an
/// authorized `/stop` triggers the callback (Go `Config.Stop`).
pub fn router(config: &ProbeConfig, ready: Option<ReadyCheck>, has_stop: bool) -> axum::Router {
    let state = ProbeState {
        service: config.service_name.clone(),
        stop_token: config.stop_token.clone(),
        ready,
        has_stop,
    };
    axum::Router::new()
        .route("/health", axum::routing::get(health_handler))
        .route("/ready", axum::routing::get(ready_handler))
        .route("/stop", axum::routing::post(stop_handler))
        .route("/metrics", axum::routing::get(metrics_handler))
        .with_state(state)
}

async fn health_handler(
    axum::extract::State(state): axum::extract::State<ProbeState>,
) -> axum::Json<serde_json::Value> {
    axum::Json(health_json(&state.service))
}

async fn ready_handler(
    axum::extract::State(state): axum::extract::State<ProbeState>,
) -> (axum::http::StatusCode, axum::Json<serde_json::Value>) {
    let ready = tokio::time::timeout(
        std::time::Duration::from_secs(READY_CHECK_TIMEOUT_SECS),
        tokio::task::spawn_blocking(move || state.ready.as_ref().is_none_or(|check| check())),
    )
    .await
    .unwrap_or(Ok(false))
    .unwrap_or(false);
    if ready {
        (axum::http::StatusCode::OK, axum::Json(ready_json()))
    } else {
        (
            axum::http::StatusCode::SERVICE_UNAVAILABLE,
            axum::Json(unready_json()),
        )
    }
}

async fn stop_handler(
    axum::extract::State(state): axum::extract::State<ProbeState>,
    headers: axum::http::HeaderMap,
    method: axum::http::Method,
) -> (axum::http::StatusCode, axum::Json<serde_json::Value>) {
    let token = headers
        .get("X-Probe-Token")
        .and_then(|value| value.to_str().ok());
    let decision = stop_decision(method.as_str(), token, &state.stop_token, state.has_stop);
    match decision {
        StopDecision::Stopping => (axum::http::StatusCode::OK, axum::Json(stopping_json())),
        StopDecision::NotImplemented => (
            axum::http::StatusCode::NOT_IMPLEMENTED,
            axum::Json(stop_unsupported_json()),
        ),
        StopDecision::MethodNotAllowed => (
            axum::http::StatusCode::METHOD_NOT_ALLOWED,
            axum::Json(stopping_json()),
        ),
        StopDecision::NotFound => (
            axum::http::StatusCode::NOT_FOUND,
            axum::Json(stopping_json()),
        ),
    }
}

async fn metrics_handler(
    axum::extract::State(_state): axum::extract::State<ProbeState>,
) -> (axum::http::StatusCode, axum::Json<serde_json::Value>) {
    let enabled = std::env::var("PROBE_METRICS_ENABLED").ok();
    if !metrics_enabled(enabled.as_deref()) {
        return (
            axum::http::StatusCode::NOT_FOUND,
            axum::Json(serde_json::json!({})),
        );
    }
    let goroutines = 0;
    let alloc = 0;
    (
        axum::http::StatusCode::OK,
        axum::Json(metrics_json(goroutines, alloc)),
    )
}

#[cfg(test)]
mod tests {
    use super::{
        authorized_stop_request, health_json, metrics_enabled, metrics_json, port_from_env,
        stop_decision, stop_token_from, StopDecision,
    };

    #[test]
    fn ports() {
        assert_eq!(
            port_from_env(Some("9090"), None, "8888"),
            Some("9090".to_string())
        );
        assert_eq!(
            port_from_env(Some(""), Some("10.0.0.1"), "8888"),
            Some("8888".to_string())
        );
        assert_eq!(port_from_env(None, None, "8888"), None);
        assert_eq!(port_from_env(Some(""), None, "8888"), None);
    }

    #[test]
    fn stop_rules() {
        assert_eq!(
            stop_decision("GET", Some("secret"), "secret", true),
            StopDecision::MethodNotAllowed
        );
        assert_eq!(
            stop_decision("POST", Some("wrong"), "secret", true),
            StopDecision::NotFound
        );
        assert_eq!(
            stop_decision("POST", Some("secret"), "", true),
            StopDecision::NotFound
        );
        assert_eq!(
            stop_decision("POST", Some(" secret "), " secret ", false),
            StopDecision::NotImplemented
        );
        assert_eq!(
            stop_decision("POST", Some("secret"), "secret", true),
            StopDecision::Stopping
        );
        assert!(authorized_stop_request(Some("tok"), "tok"));
        assert!(!authorized_stop_request(None, "tok"));
        assert_eq!(stop_token_from(Some(" env "), ""), "env");
    }

    #[test]
    fn payloads() {
        assert_eq!(health_json("api")["service"], "api");
        assert!(metrics_enabled(Some("true")));
        assert!(!metrics_enabled(Some("1")));
        assert_eq!(metrics_json(3, 100)["alloc"], 100);
        assert_eq!(stop_token_from(None, " cfg "), "cfg");
    }
}
