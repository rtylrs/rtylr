//! Worker binary skeleton. Mirrors `cmd/worker/main.go`:
//! signal handling → config → logger → config first, db connect,
//! probe on `PROBE_PORT`, and the job scheduler with an empty registry
//! (module jobs register in Phase 4). Go's worker does not call
//! `LoadDotEnv` (the api binary does); we keep the same split.

use onion_core::config;
use server::probe;

#[tokio::main]
async fn main() {
    let get = |key: &str| std::env::var(key).ok();
    let config = config::load_and_validate(&get).unwrap_or_else(|error| {
        eprintln!("config load failed: {error}");
        std::process::exit(1);
    });
    let db = db::pool::Database::connect_main(&config)
        .await
        .unwrap_or_else(|error| {
            eprintln!("failed to connect to database: {error}");
            std::process::exit(1);
        });
    // Probe: reports process identity on PROBE_PORT; the ready check
    // is the mandatory database pool.
    let probe_port = probe::port_from_process(probe::DEFAULT_PROBE_PORT);
    if let Some(probe_port) = probe_port {
        let stop_token =
            probe::stop_token_from(std::env::var("PROBE_STOP_TOKEN").ok().as_deref(), "");
        let probe_config = probe::ProbeConfig {
            service_name: "worker".to_string(),
            port: Some(probe_port.clone()),
            stop_token,
        };
        let probe_router = probe::router(&probe_config, None, false);
        let probe_addr =
            std::net::SocketAddr::from(([0, 0, 0, 0], probe_port.parse().unwrap_or(8888)));
        if let Ok(listener) = tokio::net::TcpListener::bind(probe_addr).await {
            tokio::spawn(async move {
                let _ = axum::serve(listener, probe_router).await;
            });
        }
    }
    // Job scheduler (Go `platformjobs.AutoRegister`): the registry is
    // empty until module jobs land in Phase 4; the wait keeps the
    // process alive for job execution and SIGTERM handling.
    let _ = db;
    run_until_signal().await;
}

/// Wait for SIGINT/SIGTERM, then return (Go waits on the shutdown
/// context; the drain is the runtime's job).
async fn run_until_signal() {
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
}
