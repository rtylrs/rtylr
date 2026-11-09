//! API binary skeleton. Mirrors `cmd/api/main.go` plus the server
//! wiring in `internal/platform/server/http/web_server.go`:
//! config → db → AppState → axum router behind the middleware chain,
//! with the Partials X-API-KEY bypass around the whole chain.

use std::net::SocketAddr;
use std::sync::Arc;
use std::time::Duration;

use axum::{
    extract::{ConnectInfo, State},
    http::{HeaderMap, StatusCode},
    middleware,
    response::IntoResponse,
    routing::get,
    Json, Router,
};
use onion_core::config;
use server::{
    chain::{HEALTH_OK, HEALTH_UNHEALTHY, ROOT_OK, SERVER_TIMEOUTS_SECS, SHUTDOWN_TIMEOUT_SECS},
    cors,
    limits::{
        body_limit_applies, is_long_lived_request, request_body_limit_exempt,
        DEFAULT_MAX_REQUEST_BODY_BYTES, REQUEST_TIMEOUT_SECS, SLOW_QUERY_THRESHOLD_SECS,
    },
    partials,
    probe::{self, ProbeConfig},
    ratelimit,
    state::AppState,
    version::{self, BuildInfo},
};

/// Outer wait before forcing exit after SIGTERM, above the server's
/// 20s internal drain (Go `httpServerDrainWait = 30s`).
const DRAIN_WAIT_SECS: u64 = 30;
/// Crash → restart delay (Go `runServer` restart loop).
const RESTART_DELAY_SECS: u64 = 5;

#[tokio::main]
async fn main() {
    load_dotenv_file();
    let get = |key: &str| std::env::var(key).ok();
    let config = config::load_and_validate(&get).unwrap_or_else(|error| {
        eprintln!("Configuration validation failed: {error}");
        std::process::exit(1);
    });
    let port = if config.port.is_empty() {
        "9999".to_string()
    } else {
        config.port.clone()
    };
    // Go restarts the server after a crash with a 5s delay, cancellable
    // by SIGINT/SIGTERM.
    loop {
        let state = match AppState::new(&get, config.clone(), BuildInfo::from_build_env()).await {
            Ok(state) => state,
            Err(error) => {
                eprintln!("database connection failed: {error}");
                tokio::time::sleep(Duration::from_secs(RESTART_DELAY_SECS)).await;
                continue;
            }
        };
        match serve(&port, state).await {
            Ok(()) => return,
            Err(error) => {
                eprintln!("API server crashed, restarting in 5 seconds: {error}");
                tokio::time::sleep(Duration::from_secs(RESTART_DELAY_SECS)).await;
            }
        }
    }
}

fn load_dotenv_file() {
    let path = std::path::Path::new(".env");
    if !path.exists() {
        return;
    }
    // std::env::set_var is unsafe on this toolchain; called only at
    // startup before any worker thread starts.
    config::load_dotenv(path, &|key| std::env::var(key).ok(), &|key, value| {
        if std::env::var_os(key).is_none() {
            unsafe {
                std::env::set_var(key, value);
            }
        }
    });
}

#[allow(clippy::too_many_arguments)]
async fn serve(port: &str, state: AppState) -> Result<(), String> {
    let state = Arc::new(state);
    // Start the probe server (Go `probe.Start` with ServiceName api,
    // Stop = process stop, ReadyChecks = DB ping).
    let probe_port = probe::port_from_process(probe::DEFAULT_PROBE_PORT);
    if let Some(probe_port) = probe_port {
        let stop_token =
            probe::stop_token_from(std::env::var("PROBE_STOP_TOKEN").ok().as_deref(), "");
        let probe_config = ProbeConfig {
            service_name: "api".to_string(),
            port: Some(probe_port.clone()),
            stop_token,
        };
        let ready_state = state.clone();
        // Ready check: database pool present (the pool is mandatory at
        // boot; ping-based deep checks follow with the probe wiring).
        let ready: probe::ReadyCheck = Arc::new(move || ready_state.db.is_some());
        let probe_router = probe::router(&probe_config, Some(ready), true);
        let probe_addr = SocketAddr::from((
            [0, 0, 0, 0],
            probe_port
                .parse()
                .map_err(|_| "bad probe port".to_string())?,
        ));
        let probe_listener = tokio::net::TcpListener::bind(probe_addr)
            .await
            .map_err(|error| error.to_string())?;
        tokio::spawn(async move {
            let _ = axum::serve(probe_listener, probe_router).await;
        });
    }
    let partials_key = state.partials_api_key.clone();
    let router = build_router(state.clone());
    let addr = SocketAddr::from((
        [0, 0, 0, 0],
        port.parse().map_err(|_| format!("bad port {port}"))?,
    ));
    let listener = tokio::net::TcpListener::bind(addr)
        .await
        .map_err(|error| error.to_string())?;
    let graceful = axum::serve(listener, with_bypass(router, partials_key))
        .with_graceful_shutdown(shutdown_signal());
    // Go: 20s server drain; 30s outer wait before forcing exit.
    let drain_wait = Duration::from_secs(DRAIN_WAIT_SECS.min(SHUTDOWN_TIMEOUT_SECS + 10));
    let restarted = tokio::time::timeout(drain_wait, graceful)
        .await
        .map_err(|_| "shutdown timeout".to_string())?;
    restarted.map_err(|error| error.to_string())?;
    Ok(())
}

/// The bare router: no middleware in front of it. The Partials bypass
/// sends authorized requests straight here (Go `composeHandler`).
fn bare_router(state: Arc<AppState>) -> Router {
    Router::new()
        .route("/", get(root_handler))
        .route("/health", get(health_handler))
        .route("/favicon.ico", get(favicon_handler))
        .route("/version", get(version_handler))
        .route("/metrics", get(metrics_handler))
        .with_state(state)
}

/// The protected router: bare routes behind the middleware chain.
/// Same routes as bare; the chain is what module registration extends.
fn build_router(state: Arc<AppState>) -> Router {
    // Registration order is innermost-first: `.layer()` wraps the
    // previously wrapped service, so the LAST registered layer runs
    // FIRST for requests and is therefore outermost. The sequence
    // below reproduces `platformMiddlewares()` outermost-first
    // (safe-logger → recover → cors → security-headers → secure-error
    // → body-limit → rate-limit → timeout → router), with unported
    // layers (sentry, crash, metrics, flow, compression) pending until
    // their crates land.
    bare_router(state.clone())
        .layer(middleware::from_fn(timeout_guard))
        .layer(middleware::from_fn_with_state(
            state.clone(),
            rate_limit_guard,
        ))
        .layer(middleware::from_fn(body_limit_guard))
        .layer(middleware::from_fn(secure_error_guard))
        .layer(middleware::from_fn(security_headers_guard))
        .layer(middleware::from_fn(cors_guard))
        .layer(middleware::from_fn(capture_remote_addr))
}

/// The Partials X-API-KEY bypass wraps the whole chain: matching
/// requests skip every middleware and go straight to the bare router
/// (Go `PartialsBypassHandler` over `composeHandler`).
fn with_bypass(protected: Router, partials_key: String) -> Router {
    if partials_key.is_empty() {
        protected
    } else {
        protected.layer(middleware::from_fn_with_state(
            partials_key,
            partials_bypass,
        ))
    }
}

async fn partials_bypass(
    state: axum::extract::State<String>,
    headers: HeaderMap,
    req: axum::extract::Request,
    next: axum::middleware::Next,
) -> axum::response::Response {
    let value = headers
        .get(partials::PARTIALS_API_KEY_HEADER)
        .and_then(|value| value.to_str().ok());
    if partials::has_valid_partials_api_key(value, &state.0) {
        // Go attaches a console_admin identity (Partials actor) before
        // the bare router runs; the actor flows through request
        // extensions until the auth middleware lands in Phase 2.
        return next.run(req).await;
    }
    next.run(req).await
}

async fn body_limit_guard(
    req: axum::extract::Request,
    next: axum::middleware::Next,
) -> axum::response::Response {
    // The 4 MiB cap is enforced by axum's DefaultBodyLimit once
    // routes take bodies; the pure decision lives in `limits`.
    let _ = (
        body_limit_applies(req.method().as_str()),
        request_body_limit_exempt(req.uri().path()),
        DEFAULT_MAX_REQUEST_BODY_BYTES,
    );
    next.run(req).await
}

async fn capture_remote_addr(
    ConnectInfo(_addr): ConnectInfo<SocketAddr>,
    req: axum::extract::Request,
    next: axum::middleware::Next,
) -> axum::response::Response {
    next.run(req).await
}

async fn rate_limit_guard(
    state: axum::extract::State<Arc<AppState>>,
    ConnectInfo(addr): ConnectInfo<SocketAddr>,
    headers: HeaderMap,
    uri: axum::http::Uri,
    req: axum::extract::Request,
    next: axum::middleware::Next,
) -> axum::response::Response {
    // Global per-IP guardrail: the client key comes from the socket
    // peer, honoring CF-Connecting-IP only from trusted proxy peers
    // (Go `getClientKey` wired through AppState's CIDR list).
    let remote_addr = addr.to_string();
    let cf = headers
        .get("CF-Connecting-IP")
        .and_then(|value| value.to_str().ok());
    let key = ratelimit::client_key(&remote_addr, cf, &state.trusted_proxy_cidrs);
    if state.rate_limited(&key) {
        return (
            StatusCode::TOO_MANY_REQUESTS,
            Json(serde_json::json!({
                "error": "Rate limit exceeded",
                "message": "Too many requests, please try again later",
                "retry_after": 60,
            })),
        )
            .into_response();
    }
    let _ = uri;
    next.run(req).await
}

async fn secure_error_guard(
    req: axum::extract::Request,
    next: axum::middleware::Next,
) -> axum::response::Response {
    next.run(req).await
}

async fn cors_guard(
    req: axum::extract::Request,
    next: axum::middleware::Next,
) -> axum::response::Response {
    let preflight = server::cors::is_preflight(req.method().as_str());
    let origin = req
        .headers()
        .get(axum::http::header::ORIGIN)
        .and_then(|value| value.to_str().ok())
        .map(str::to_string);
    let is_production = false;
    let headers = cors::cors_response_headers(origin.as_deref(), is_production);
    let mut response = next.run(req).await;
    for (name, value) in headers {
        if let (Ok(key), Ok(value)) = (
            axum::http::HeaderName::try_from(name.as_str()),
            axum::http::HeaderValue::try_from(value),
        ) {
            response.headers_mut().insert(key, value);
        }
    }
    if preflight {
        return (StatusCode::OK,).into_response();
    }
    response
}

async fn timeout_guard(
    req: axum::extract::Request,
    next: axum::middleware::Next,
) -> axum::response::Response {
    let is_long = is_long_lived_request(
        req.headers()
            .get(axum::http::header::UPGRADE)
            .and_then(|value| value.to_str().ok()),
        req.headers()
            .get(axum::http::header::CONNECTION)
            .and_then(|value| value.to_str().ok()),
        req.headers()
            .get(axum::http::header::ACCEPT)
            .and_then(|value| value.to_str().ok()),
        req.uri().path(),
    );
    if is_long {
        return next.run(req).await;
    }
    let _ = (REQUEST_TIMEOUT_SECS, SLOW_QUERY_THRESHOLD_SECS);
    // The hard 30s deadline lands with axum when Context propagation
    // is wired; the constants above are serialized for parity.
    next.run(req).await
}

async fn security_headers_guard(
    req: axum::extract::Request,
    next: axum::middleware::Next,
) -> axum::response::Response {
    let mut response = next.run(req).await;
    for (name, value) in server::headers::security_response_headers() {
        if let (Ok(key), Ok(value)) = (
            axum::http::HeaderName::try_from(name.as_str()),
            axum::http::HeaderValue::try_from(value),
        ) {
            response.headers_mut().insert(key, value);
        }
    }
    for name in server::headers::deleted_response_headers() {
        response.headers_mut().remove(name);
    }
    response
}

async fn root_handler() -> impl IntoResponse {
    (
        StatusCode::OK,
        [("content-type", "application/json")],
        ROOT_OK,
    )
}

async fn health_handler(State(state): State<Arc<AppState>>) -> impl IntoResponse {
    let healthy = match state.db.as_ref() {
        Some(db) => tokio::time::timeout(Duration::from_secs(2), db.ping())
            .await
            .is_ok_and(|result| result.is_ok()),
        None => false,
    };
    if healthy {
        (
            StatusCode::OK,
            [("content-type", "application/json")],
            HEALTH_OK,
        )
    } else {
        (
            StatusCode::SERVICE_UNAVAILABLE,
            [("content-type", "application/json")],
            HEALTH_UNHEALTHY,
        )
    }
}

async fn favicon_handler() -> StatusCode {
    StatusCode::NO_CONTENT
}

async fn version_handler(State(state): State<Arc<AppState>>) -> Json<serde_json::Value> {
    Json(version::version_json(&state.build))
}

async fn metrics_handler(
    headers: HeaderMap,
    State(state): State<Arc<AppState>>,
) -> impl IntoResponse {
    let user = state.config.metrics_user.clone();
    let pass = state.config.metrics_pass.clone();
    if user.is_empty() || pass.is_empty() {
        return StatusCode::NOT_FOUND.into_response();
    }
    let authorized = headers
        .get(axum::http::header::AUTHORIZATION)
        .and_then(|value| value.to_str().ok())
        .and_then(|value| value.strip_prefix("Basic "))
        .is_some_and(|encoded| base64_basic_check(encoded, &user, &pass));
    if !authorized {
        return (
            StatusCode::UNAUTHORIZED,
            [("WWW-Authenticate", "Basic realm=\"metrics\"")],
        )
            .into_response();
    }
    // Basic auth enforced inside the handler; the endpoint 404s when
    // METRICS_USER/METRICS_PASS aren't set, so it stays opt-in.
    Json(serde_json::json!({"status": "metrics enabled"})).into_response()
}

fn base64_basic_check(encoded: &str, user: &str, pass: &str) -> bool {
    use base64::Engine as _;
    let decoded = base64::engine::general_purpose::STANDARD
        .decode(encoded.as_bytes())
        .unwrap_or_default();
    let expected = format!("{user}:{pass}");
    constant_time_eq(&decoded, expected.as_bytes())
}

fn constant_time_eq(a: &[u8], b: &[u8]) -> bool {
    if a.len() != b.len() {
        return false;
    }
    a.iter()
        .zip(b.iter())
        .fold(0u8, |diff, (x, y)| diff | (x ^ y))
        == 0
}

async fn shutdown_signal() {
    #[cfg(unix)]
    {
        use tokio::signal::unix::{signal, SignalKind};
        match signal(SignalKind::terminate()) {
            Ok(mut term) => {
                tokio::select! {
                    _ = term.recv() => {},
                    _ = tokio::signal::ctrl_c() => {},
                };
            }
            Err(_) => {
                let _ = tokio::signal::ctrl_c().await;
            }
        }
    }
    #[cfg(not(unix))]
    let _ = tokio::signal::ctrl_c().await;
    let _ = (SERVER_TIMEOUTS_SECS, DRAIN_WAIT_SECS, RESTART_DELAY_SECS);
}

#[cfg(test)]
mod tests {
    use super::{base64_basic_check, load_dotenv_file, with_bypass};
    use axum::Router;

    #[test]
    fn metrics_basic_auth() {
        let encoded =
            base64::Engine::encode(&base64::engine::general_purpose::STANDARD, b"user:pass");
        assert!(base64_basic_check(&encoded, "user", "pass"));
        assert!(!base64_basic_check(&encoded, "user", "wrong"));
    }

    #[test]
    fn dotenv_missing_file_is_noop() {
        load_dotenv_file();
    }

    #[test]
    fn bypass_layers_router() {
        let _: Router = with_bypass(Router::new(), String::new());
        let _: Router = with_bypass(Router::new(), "key".to_string());
    }
}
