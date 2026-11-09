//! Configuration and environment parity with `internal/platform/config`.
//!
//! Env names and semantics are byte-identical to Go (see
//! `services/onion/.env.example`, checked by
//! `scripts/check_env_example.py`). The Go `GlobalConfig` pointer becomes
//! an explicit `AppConfig` value passed by callers; getters are methods on
//! `AppConfig` or free functions reading the environment, never a
//! process-global singleton.

use std::collections::HashMap;
use std::net::IpAddr;
use std::path::Path;
use std::time::Duration;

use thiserror::Error;

// Base URLs of the rtylr web apps (product constants, `appurls.go`).
pub const AUTH_APP_URL: &str = "https://sso.rtylr.com";
pub const HUB_APP_URL: &str = "https://portal.rtylr.com";
pub const POS_APP_URL: &str = "https://pos.rtylr.com";
pub const ERP_APP_URL: &str = "https://erp.rtylr.com";
pub const HR_APP_URL: &str = "https://hr.rtylr.com";
pub const CRM_APP_URL: &str = "https://crm.rtylr.com";
pub const FINANCE_APP_URL: &str = "https://finance.rtylr.com";
pub const FLOW_APP_URL: &str = "https://portal.rtylr.com/flow";
pub const INSIGHTS_APP_URL: &str = "https://portal.rtylr.com/insights";
pub const RECRUIT_APP_URL: &str = "https://recruit.rtylr.com";
pub const DEFAULT_API_URL: &str = "https://api.voxire.com";
pub const LANDING_URL: &str = "https://rtylr.com";

pub const DEFAULT_JWT_DURATION: Duration = Duration::from_secs(24 * 60 * 60);
pub const OPENCODE_BASE_URL: &str = "https://opencode.ai/zen/go/v1";
pub const OPENCODE_MODEL: &str = "deepseek-v4.1-flash";
pub const DEEPSEEK_BASE_URL: &str = "https://api.deepseek.com";
pub const DEEPSEEK_MODEL: &str = "deepseek-chat";
pub const AI_REQUEST_TIMEOUT_SECS: u64 = 90;
pub const AI_MAX_OUTPUT_TOKENS: i32 = 2048;
pub const AI_MAX_RETRIES: i32 = 2;
pub const AI_TOKENS_PER_CREDIT: i32 = 1000;

pub const DEFAULT_TRINO_HOST: &str = "localhost";
pub const DEFAULT_TRINO_PORT: &str = "8080";
pub const DEFAULT_TRINO_SOURCE: &str = "onion";

pub const DEFAULT_LEAD_NOTIFY_EMAILS: [&str; 5] = [
    "info@voxire.com",
    "sales@voxire.com",
    "mohammad.saab@voxire.com",
    "mohamad.homsi@voxire.com",
    "amounehabed@gmail.com",
];
pub const DEFAULT_NEWSLETTER_NOTIFY_EMAILS: [&str; 1] = ["info@voxire.com"];

/// Cloudflare proxy ranges (`trusted_proxy.go` defaults).
pub const DEFAULT_TRUSTED_PROXY_CIDRS: [&str; 22] = [
    "173.245.48.0/20",
    "103.21.244.0/22",
    "103.22.200.0/22",
    "103.31.4.0/22",
    "141.101.64.0/18",
    "108.162.192.0/18",
    "190.93.240.0/20",
    "188.114.96.0/20",
    "197.234.240.0/22",
    "198.41.128.0/17",
    "162.158.0.0/15",
    "104.16.0.0/13",
    "104.24.0.0/14",
    "172.64.0.0/13",
    "131.0.72.0/22",
    "2400:cb00::/32",
    "2606:4700::/32",
    "2803:f800::/32",
    "2405:b500::/32",
    "2405:8100::/32",
    "2a06:98c0::/29",
    "2c0f:f248::/32",
];

pub type EnvMap = HashMap<String, String>;

/// Lookup function used by the loader; `None` means unset.
pub type EnvGet<'a> = &'a dyn Fn(&str) -> Option<String>;

#[derive(Debug, Error, PartialEq)]
pub enum ConfigError {
    #[error("invalid JWT_DURATION {0:?}: {1}")]
    InvalidJwtDuration(String, String),
    #[error("JWT_DURATION must be greater than zero, got {0:?}")]
    NonPositiveJwtDuration(String),
    #[error("REDIS_HOSTS contains no usable host")]
    EmptyRedisHosts,
    #[error("missing critical environment variable(s): {0}")]
    Missing(String),
}

pub fn env_or_default(get: EnvGet<'_>, key: &str, fallback: &str) -> String {
    match get(key) {
        Some(value) if !value.is_empty() => value,
        _ => fallback.to_string(),
    }
}

pub fn env_first(get: EnvGet<'_>, keys: &[&str]) -> String {
    for key in keys {
        if let Some(value) = get(key) {
            let trimmed = value.trim();
            if !trimmed.is_empty() {
                return trimmed.to_string();
            }
        }
    }
    String::new()
}

pub fn env_int_or_default(get: EnvGet<'_>, key: &str, fallback: i32) -> i32 {
    match get(key) {
        Some(raw) if !raw.is_empty() => raw.parse::<i32>().unwrap_or(fallback),
        _ => fallback,
    }
}

pub fn positive_env_int_or_default(get: EnvGet<'_>, key: &str, fallback: i32) -> i32 {
    let value = env_int_or_default(get, key, fallback);
    if value <= 0 {
        fallback
    } else {
        value
    }
}

fn parse_go_duration(raw: &str) -> Option<Duration> {
    let text = raw.trim();
    if text.is_empty() {
        return None;
    }
    let mut rest = text;
    let mut total_ns: f64 = 0.0;
    let mut saw_unit = false;
    while !rest.is_empty() {
        if rest == "0" {
            return Some(Duration::ZERO);
        }
        let digits_end = rest
            .find(|c: char| !(c.is_ascii_digit() || c == '.' || c == '-' || c == '+'))
            .unwrap_or(rest.len());
        if digits_end == 0 {
            return None;
        }
        let number: f64 = rest[..digits_end].parse().ok()?;
        let unit_start = digits_end;
        let unit_end = rest[unit_start..]
            .find(|c: char| c.is_ascii_digit() || c == '.' || c == '-' || c == '+')
            .map(|i| unit_start + i)
            .unwrap_or(rest.len());
        let unit = &rest[unit_start..unit_end];
        let factor = match unit {
            "ns" => 1.0,
            "us" | "µs" | "μs" => 1e3,
            "ms" => 1e6,
            "s" => 1e9,
            "m" => 60e9,
            "h" => 3600e9,
            _ => return None,
        };
        total_ns += number * factor;
        saw_unit = true;
        rest = &rest[unit_end..];
    }
    if !saw_unit || !total_ns.is_finite() {
        return None;
    }
    Some(Duration::from_nanos(total_ns.max(0.0) as u64))
}

pub fn normalize_redis_hosts(raw: &str) -> Vec<String> {
    let trimmed = raw.trim();
    if trimmed.is_empty() {
        return vec!["localhost:6379".to_string()];
    }
    trimmed
        .split(',')
        .map(str::trim)
        .filter(|host| !host.is_empty())
        .map(str::to_string)
        .collect()
}

pub fn get_in_memory_cache_max_size_mb(get: EnvGet<'_>) -> i32 {
    match get("INMEMORY_CACHE_MAX_SIZE_MB") {
        Some(raw) if !raw.is_empty() => raw.parse::<i32>().unwrap_or(100),
        _ => 100,
    }
}

pub fn parse_email_list(raw: &str, fallback: &[&str]) -> Vec<String> {
    let mut seen = std::collections::HashSet::new();
    let mut out = Vec::new();
    for part in raw.split(',') {
        let addr = part.trim().to_lowercase();
        if addr.is_empty() || !seen.insert(addr.clone()) {
            continue;
        }
        out.push(addr);
    }
    if out.is_empty() {
        fallback.iter().map(|s| (*s).to_string()).collect()
    } else {
        out
    }
}

#[derive(Debug, Clone, PartialEq)]
pub enum AppEnvClass {
    Empty,
    Unrecognized,
    NonProduction,
    Production,
}

pub fn normalize_app_env(raw: &str) -> String {
    raw.trim().to_ascii_lowercase()
}

pub fn classify_app_env(raw: &str) -> AppEnvClass {
    match normalize_app_env(raw).as_str() {
        "" => AppEnvClass::Empty,
        "dev" | "development" | "local" | "test" => AppEnvClass::NonProduction,
        "production" | "prod" => AppEnvClass::Production,
        _ => AppEnvClass::Unrecognized,
    }
}

pub fn is_non_production_env(normalized: &str) -> bool {
    matches!(normalized, "dev" | "development" | "local" | "test")
}

pub fn is_declared_production_env(normalized: &str) -> bool {
    matches!(normalized, "production" | "prod")
}

/// Boot classification of the raw `APP_ENV` value (`assertAppEnvIsUnderstood`).
pub fn app_env_class(raw: &str) -> AppEnvClass {
    let normalized = normalize_app_env(raw);
    if normalized.is_empty() {
        return AppEnvClass::Empty;
    }
    if is_non_production_env(&normalized) {
        return AppEnvClass::NonProduction;
    }
    if is_declared_production_env(&normalized) {
        return AppEnvClass::Production;
    }
    AppEnvClass::Unrecognized
}

pub fn get_trusted_proxy_cidrs(get: EnvGet<'_>) -> Vec<String> {
    let raw = get("TRUSTED_PROXY_CIDRS").unwrap_or_default();
    let trimmed = raw.trim();
    if trimmed.eq_ignore_ascii_case("none") {
        return Vec::new();
    }
    if trimmed.is_empty() {
        return DEFAULT_TRUSTED_PROXY_CIDRS
            .iter()
            .map(|s| (*s).to_string())
            .collect();
    }
    trimmed
        .split(',')
        .map(str::trim)
        .filter(|entry| !entry.is_empty())
        .map(str::to_string)
        .collect()
}

fn remote_addr_host(remote_addr: &str) -> String {
    let trimmed = remote_addr.trim();
    if let Some(stripped) = trimmed.strip_prefix('[') {
        if let Some(end) = stripped.find(']') {
            return stripped[..end].to_string();
        }
    }
    if let Some(colon) = trimmed.rfind(':') {
        let (host, port) = trimmed.split_at(colon);
        if port.len() > 1 && port[1..].chars().all(|c| c.is_ascii_digit()) {
            return host.trim_matches(['[', ']']).to_string();
        }
    }
    trimmed.trim().trim_matches(['[', ']']).to_string()
}

fn parse_cidr(raw: &str) -> Option<(IpAddr, u32)> {
    let (addr_text, prefix_text) = raw.split_once('/')?;
    let addr: IpAddr = addr_text.trim().parse().ok()?;
    let bits: u32 = prefix_text.trim().parse().ok()?;
    let max = match addr {
        IpAddr::V4(_) => 32,
        IpAddr::V6(_) => 128,
    };
    if bits > max {
        return None;
    }
    Some((addr, bits))
}

fn ip_in_cidr(ip: IpAddr, network: IpAddr, bits: u32) -> bool {
    match (ip, network) {
        (IpAddr::V4(ip4), IpAddr::V4(net4)) => {
            let mask = if bits == 0 {
                0u32
            } else {
                u32::MAX << (32 - bits)
            };
            u32::from(ip4) & mask == u32::from(net4) & mask
        }
        (IpAddr::V6(ip6), IpAddr::V6(net6)) => {
            if bits == 0 {
                return true;
            }
            let blocks = u128::MAX << (128 - bits);
            u128::from(ip6) & blocks == u128::from(net6) & blocks
        }
        _ => false,
    }
}

pub fn is_trusted_proxy_ip(remote_addr: &str, cidrs: &[String]) -> bool {
    let host = remote_addr_host(remote_addr);
    let ip: IpAddr = match host.parse() {
        Ok(ip) => ip,
        Err(_) => return false,
    };
    cidrs
        .iter()
        .filter_map(|raw| parse_cidr(raw))
        .any(|(network, bits)| ip_in_cidr(ip, network, bits))
}

pub fn load_dotenv(path: &Path, get: EnvGet<'_>, set: &dyn Fn(&str, &str)) {
    let Ok(text) = std::fs::read_to_string(path) else {
        return;
    };
    for line in text.lines() {
        let line = line.trim();
        if line.is_empty() || line.starts_with('#') {
            continue;
        }
        let Some((key, value)) = line.split_once('=') else {
            continue;
        };
        let key = key.trim();
        if key.is_empty() {
            continue;
        }
        if get(key).is_some() {
            continue;
        }
        let value = value.trim().trim_matches(['"', '\'']);
        set(key, value);
    }
}

pub fn listen_port(get: EnvGet<'_>, fallback: &str) -> String {
    match get("PORT") {
        Some(value) if !value.trim().is_empty() => value.trim().to_string(),
        _ => fallback.to_string(),
    }
}

#[derive(Debug, Clone, PartialEq)]
pub struct AppConfig {
    pub run_mode: String,
    pub database_url: String,
    pub db_host: String,
    pub db_port: String,
    pub db_user: String,
    pub db_pass: String,
    pub db_name: String,
    pub jwt_secret: String,
    pub jwt_duration: Duration,
    pub resend_key: String,
    pub lead_notify_emails: Vec<String>,
    pub newsletter_notify_emails: Vec<String>,
    pub resend_audience_id: String,
    pub resend_audience_key: String,
    pub app_env: String,
    pub app_release: String,
    pub port: String,
    pub hostname: String,
    pub auth_app_url: String,
    pub google_oauth_client_id: String,
    pub google_oauth_client_secret: String,
    pub google_oauth_redirect_uri: String,
    pub db_max_open_conns: i32,
    pub db_max_idle_conns: i32,
    pub db_conn_max_lifetime: Duration,
    pub db_conn_max_idle_time: Duration,
    pub openai_api_key: String,
    pub openai_org_id: String,
    pub openai_base_url: String,
    pub gemini_api_key: String,
    pub opencode_api_key: String,
    pub deepseek_api_key: String,
    pub ai_primary_provider: String,
    pub searxng_url: String,
    pub brave_search_api_key: String,
    pub form_mapping_checkpoint_path: String,
    pub redis_url: String,
    pub redis_hosts: Vec<String>,
    pub redis_username: String,
    pub redis_password: String,
    pub redis_enabled: bool,
    pub in_memory_cache_max_size_mb: i32,
    pub trino_enabled: bool,
    pub trino_host: String,
    pub trino_port: String,
    pub trino_user: String,
    pub trino_password: String,
    pub trino_access_token: String,
    pub trino_catalog: String,
    pub trino_schema: String,
    pub trino_tls: bool,
    pub trino_source: String,
    pub paddle_api_key: String,
    pub paddle_webhook_secret: String,
    pub paddle_client_token: String,
    pub paddle_sandbox: bool,
    pub logging_database_url: String,
    pub better_stack_source_token: String,
    pub better_stack_ingesting_host: String,
    pub slack_logging_enabled: bool,
    pub slack_bot_token: String,
    pub slack_alerts_channel: String,
    pub slack_request_logs_channel: String,
    pub slack_dashboard_channel: String,
    pub sentry_dsn: String,
    pub google_sheets_lead_sheet_id: String,
    pub google_sheets_credentials: String,
    pub s3_bucket_name: String,
    pub s3_region: String,
    pub s3_endpoint: String,
    pub s3_force_path_style: bool,
    pub floci_enabled: bool,
    pub floci_endpoint: String,
    pub floci_region: String,
    pub floci_bucket: String,
    pub cdn_base_url: String,
    pub public_api_url: String,
    pub cdn_max_upload_bytes: i64,
    pub cdn_media_engine_path: String,
    pub cdn_api_key: String,
    pub cdn_moderation_enabled: bool,
    pub cdn_moderation_provider: String,
    pub cdn_moderation_fail_closed: bool,
    pub internal_token: String,
    pub partials_api_key: String,
    pub partials_actor_user_id: i64,
    pub metrics_user: String,
    pub metrics_pass: String,
}

impl AppConfig {
    pub fn is_production(&self) -> bool {
        !is_non_production_env(&normalize_app_env(&self.app_env))
    }

    pub fn is_declared_production(&self) -> bool {
        is_declared_production_env(&normalize_app_env(&self.app_env))
    }

    pub fn openai_base_url(&self) -> String {
        if self.openai_base_url.is_empty() {
            "https://api.openai.com/v1".to_string()
        } else {
            self.openai_base_url.clone()
        }
    }

    pub fn is_openai_configured(&self) -> bool {
        !self.openai_api_key.is_empty()
    }

    pub fn is_gemini_configured(&self) -> bool {
        !self.gemini_api_key.is_empty()
    }

    pub fn opencode_api_key(&self) -> String {
        self.opencode_api_key.trim().to_string()
    }

    pub fn is_opencode_configured(&self) -> bool {
        !self.opencode_api_key().is_empty()
    }

    pub fn is_deepseek_configured(&self) -> bool {
        !self.deepseek_api_key.is_empty()
    }

    pub fn is_any_ai_configured(&self) -> bool {
        self.is_deepseek_configured()
            || self.is_openai_configured()
            || self.is_gemini_configured()
            || self.is_opencode_configured()
    }

    pub fn paddle_base_url(&self) -> &'static str {
        if self.paddle_sandbox {
            "https://sandbox-api.paddle.com"
        } else {
            "https://api.paddle.com"
        }
    }

    pub fn redis_configured(&self) -> bool {
        self.redis_enabled && (!self.redis_url.is_empty() || !self.redis_hosts.is_empty())
    }

    pub fn trino_host(&self) -> String {
        if !self.trino_host.is_empty() {
            self.trino_host.clone()
        } else {
            DEFAULT_TRINO_HOST.to_string()
        }
    }

    pub fn trino_port(&self) -> String {
        if !self.trino_port.is_empty() {
            self.trino_port.clone()
        } else {
            DEFAULT_TRINO_PORT.to_string()
        }
    }

    pub fn trino_source(&self) -> String {
        if !self.trino_source.is_empty() {
            self.trino_source.clone()
        } else {
            DEFAULT_TRINO_SOURCE.to_string()
        }
    }

    pub fn trino_configured(&self) -> bool {
        self.trino_enabled && !self.trino_host().is_empty() && !self.trino_user.is_empty()
    }

    pub fn in_memory_cache_max_size_mb(&self) -> i32 {
        if self.in_memory_cache_max_size_mb > 0 {
            self.in_memory_cache_max_size_mb
        } else {
            100
        }
    }
}

pub fn load_and_validate(get: EnvGet<'_>) -> Result<AppConfig, ConfigError> {
    let raw_duration = get("JWT_DURATION").unwrap_or_default();
    let jwt_duration = if raw_duration.is_empty() {
        DEFAULT_JWT_DURATION
    } else {
        let parsed = parse_go_duration(&raw_duration).ok_or_else(|| {
            ConfigError::InvalidJwtDuration(raw_duration.clone(), "unknown unit".to_string())
        })?;
        if parsed.is_zero() {
            return Err(ConfigError::NonPositiveJwtDuration(raw_duration));
        }
        parsed
    };

    let raw_redis = get("REDIS_HOSTS").unwrap_or_default();
    let redis_hosts = normalize_redis_hosts(&raw_redis);
    if !raw_redis.trim().is_empty() && redis_hosts.is_empty() {
        return Err(ConfigError::EmptyRedisHosts);
    }

    let int_positive = |key: &str| -> i32 {
        get(key)
            .and_then(|raw| raw.parse::<i32>().ok())
            .filter(|value| *value > 0)
            .unwrap_or(0)
    };
    let duration_positive = |key: &str| -> Duration {
        get(key)
            .and_then(|raw| parse_go_duration(&raw))
            .filter(|value| !value.is_zero())
            .unwrap_or(Duration::ZERO)
    };
    let b_true =
        |key: &str| -> bool { get(key).unwrap_or_default().to_ascii_lowercase().trim() == "true" };
    let b_equal_fold = |key: &str, want: &str| -> bool {
        get(key)
            .unwrap_or_default()
            .trim()
            .eq_ignore_ascii_case(want)
    };
    let text = |key: &str| -> String { get(key).unwrap_or_default() };
    let trimmed = |key: &str| -> String { text(key).trim().to_string() };

    let auth_app_url = env_first(get, &["AUTH_URL", "AUTH_APP_URL"]);
    let google_redirect = {
        let value = trimmed("GOOGLE_OAUTH_REDIRECT_URI");
        if value.is_empty() {
            let base = env_or_default(get, "AUTH_URL", AUTH_APP_URL);
            format!("{}/callback", base.trim_end_matches('/'))
        } else {
            value
        }
    };
    let mut cfg = AppConfig {
        run_mode: env_or_default(get, "RUN_MODE", "api"),
        database_url: text("DATABASE_URL"),
        db_host: text("DB_HOST"),
        db_port: text("DB_PORT"),
        db_user: text("DB_USER"),
        db_pass: text("DB_PASSWORD"),
        db_name: text("DB_NAME"),
        jwt_secret: text("JWT_SECRET"),
        jwt_duration,
        resend_key: text("RESEND_API_KEY"),
        lead_notify_emails: parse_email_list(
            &text("LEAD_NOTIFY_EMAILS"),
            &DEFAULT_LEAD_NOTIFY_EMAILS,
        ),
        newsletter_notify_emails: parse_email_list(
            &text("NEWSLETTER_NOTIFY_EMAILS"),
            &DEFAULT_NEWSLETTER_NOTIFY_EMAILS,
        ),
        resend_audience_id: trimmed("RESEND_AUDIENCE_ID"),
        resend_audience_key: trimmed("RESEND_AUDIENCE_KEY"),
        app_env: text("APP_ENV"),
        app_release: text("APP_RELEASE"),
        port: text("PORT"),
        hostname: text("HOSTNAME"),
        auth_app_url: auth_app_url.clone(),
        google_oauth_client_id: text("GOOGLE_OAUTH_CLIENT_ID"),
        google_oauth_client_secret: text("GOOGLE_OAUTH_CLIENT_SECRET"),
        google_oauth_redirect_uri: google_redirect,
        db_max_open_conns: int_positive("DB_MAX_OPEN_CONNS"),
        db_max_idle_conns: int_positive("DB_MAX_IDLE_CONNS"),
        db_conn_max_lifetime: duration_positive("DB_CONN_MAX_LIFETIME"),
        db_conn_max_idle_time: duration_positive("DB_CONN_MAX_IDLE_TIME"),
        openai_api_key: text("OPENAI_API_KEY"),
        openai_org_id: text("OPENAI_ORG_ID"),
        openai_base_url: text("OPENAI_BASE_URL"),
        gemini_api_key: text("GEMINI_API_KEY"),
        opencode_api_key: trimmed("OPENCODE_API_KEY"),
        deepseek_api_key: text("DEEPSEEK_API_KEY"),
        ai_primary_provider: text("AI_PRIMARY_PROVIDER"),
        searxng_url: trimmed("SEARXNG_URL"),
        brave_search_api_key: trimmed("BRAVE_SEARCH_API_KEY"),
        form_mapping_checkpoint_path: trimmed("FORM_MAPPING_CHECKPOINT_PATH"),
        redis_url: text("REDIS_URL"),
        redis_hosts,
        redis_username: text("REDIS_USERNAME"),
        redis_password: text("REDIS_PASSWORD"),
        redis_enabled: b_true("REDIS_ENABLED"),
        in_memory_cache_max_size_mb: get("INMEMORY_CACHE_MAX_SIZE_MB")
            .and_then(|raw| raw.parse::<i32>().ok())
            .unwrap_or(0),
        trino_enabled: b_true("TRINO_ENABLED"),
        trino_host: trimmed("TRINO_HOST"),
        trino_port: trimmed("TRINO_PORT"),
        trino_user: trimmed("TRINO_USER"),
        trino_password: text("TRINO_PASSWORD"),
        trino_access_token: text("TRINO_ACCESS_TOKEN"),
        trino_catalog: trimmed("TRINO_CATALOG"),
        trino_schema: trimmed("TRINO_SCHEMA"),
        trino_tls: b_equal_fold("TRINO_TLS", "true"),
        trino_source: trimmed("TRINO_SOURCE"),
        paddle_api_key: text("PADDLE_API_KEY"),
        paddle_webhook_secret: text("PADDLE_WEBHOOK_SECRET"),
        paddle_client_token: text("PADDLE_CLIENT_TOKEN"),
        paddle_sandbox: text("PADDLE_SANDBOX") == "true",
        logging_database_url: text("LOGGING_DATABASE_URL"),
        better_stack_source_token: text("BETTERSTACK_SOURCE_TOKEN"),
        better_stack_ingesting_host: trimmed("BETTERSTACK_INGESTING_HOST"),
        slack_logging_enabled: b_true("SLACK_LOGGING_ENABLED"),
        slack_bot_token: text("SLACK_BOT_TOKEN"),
        slack_alerts_channel: text("SLACK_ALERTS_CHANNEL_ID"),
        slack_request_logs_channel: text("SLACK_REQUEST_LOGS_CHANNEL"),
        slack_dashboard_channel: text("SLACK_DASHBOARD_CHANNEL_ID"),
        sentry_dsn: text("SENTRY_DSN"),
        google_sheets_lead_sheet_id: text("GOOGLE_SHEETS_LEAD_SHEET_ID"),
        google_sheets_credentials: text("GOOGLE_SHEETS_CREDENTIALS"),
        s3_bucket_name: text("S3_BUCKET_NAME"),
        s3_region: text("S3_REGION"),
        s3_endpoint: text("S3_ENDPOINT"),
        s3_force_path_style: b_equal_fold("S3_FORCE_PATH_STYLE", "true"),
        floci_enabled: b_true("FLOCI_ENABLED"),
        floci_endpoint: text("FLOCI_ENDPOINT"),
        floci_region: text("FLOCI_REGION"),
        floci_bucket: text("FLOCI_BUCKET"),
        cdn_base_url: env_first(get, &["UPLOAD_URL", "CDN_BASE_URL"]),
        public_api_url: env_or_default(get, "PUBLIC_API_URL", DEFAULT_API_URL)
            .trim_end_matches('/')
            .to_string(),
        cdn_max_upload_bytes: i64::from(positive_env_int_or_default(get, "CDN_MAX_UPLOAD_MB", 100))
            * 1024
            * 1024,
        cdn_media_engine_path: env_or_default(
            get,
            "MEDIA_ENGINE_PATH",
            "./media-engine/build/media-engine",
        ),
        cdn_api_key: text("CDN_API_KEY"),
        cdn_moderation_enabled: !b_equal_fold("CDN_MODERATION_ENABLED", "false"),
        cdn_moderation_provider: env_or_default(get, "CDN_MODERATION_PROVIDER", "openai")
            .to_ascii_lowercase(),
        cdn_moderation_fail_closed: b_equal_fold("CDN_MODERATION_FAIL_CLOSED", "true"),
        internal_token: text("INTERNAL_TOKEN"),
        partials_api_key: trimmed("PARTIALS_API_KEY"),
        partials_actor_user_id: get("PARTIALS_ACTOR_USER_ID")
            .and_then(|raw| raw.trim().parse::<i64>().ok())
            .unwrap_or(0),
        metrics_user: text("METRICS_USER"),
        metrics_pass: text("METRICS_PASS"),
    };
    if cfg.auth_app_url.is_empty() {
        cfg.auth_app_url = AUTH_APP_URL.to_string();
    }
    if cfg.cdn_base_url.is_empty() {
        cfg.cdn_base_url = "https://api.voxire.com".to_string();
    }

    let mut missing = Vec::new();
    if cfg.database_url.is_empty()
        && (cfg.db_host.is_empty() || cfg.db_name.is_empty() || cfg.db_user.is_empty())
    {
        missing.push("DATABASE_URL (or DB_HOST, DB_NAME, DB_USER)");
    }
    if cfg.jwt_secret.is_empty() {
        missing.push("JWT_SECRET");
    }
    if !missing.is_empty() {
        return Err(ConfigError::Missing(missing.join(", ")));
    }
    Ok(cfg)
}

/// Read the process environment.
pub fn load_and_validate_from_process() -> Result<AppConfig, ConfigError> {
    let get_env = |key: &str| std::env::var(key).ok();
    load_and_validate(&get_env)
}

pub fn env_getter_from_map(map: &EnvMap) -> impl Fn(&str) -> Option<String> + '_ {
    move |key: &str| map.get(key).cloned()
}

#[cfg(test)]
mod tests {
    use super::{
        app_env_class, get_in_memory_cache_max_size_mb, get_trusted_proxy_cidrs,
        is_trusted_proxy_ip, listen_port, load_and_validate, normalize_app_env,
        normalize_redis_hosts, parse_email_list, AppEnvClass, ConfigError,
        DEFAULT_LEAD_NOTIFY_EMAILS, DEFAULT_NEWSLETTER_NOTIFY_EMAILS,
    };
    use std::collections::HashMap;

    fn env_map(pairs: &[(&str, &str)]) -> HashMap<String, String> {
        pairs
            .iter()
            .map(|(key, value)| ((*key).to_string(), (*value).to_string()))
            .collect()
    }

    fn minimal_env() -> HashMap<String, String> {
        env_map(&[
            ("RUN_MODE", "api"),
            ("DB_HOST", "127.0.0.1"),
            ("DB_PORT", "3306"),
            ("DB_USER", "test"),
            ("DB_PASSWORD", "test"),
            ("DB_NAME", "onion_test"),
            ("JWT_SECRET", "test-secret"),
        ])
    }

    fn load(map: &HashMap<String, String>) -> Result<super::AppConfig, ConfigError> {
        load_and_validate(&|key| map.get(key).cloned())
    }

    #[test]
    fn redis_hosts_trim_and_filter() {
        let map = minimal_env();
        let get = |key: &str| map.get(key).cloned();
        assert_eq!(
            normalize_redis_hosts(" redis-a:6379, , redis-b:6380, "),
            vec!["redis-a:6379".to_string(), "redis-b:6380".to_string()]
        );
        assert_eq!(
            super::env_getter_from_map(&map)("DB_HOST"),
            Some("127.0.0.1".to_string())
        );
        let _ = get("DB_HOST");
        assert_eq!(get_in_memory_cache_max_size_mb(&|_| None), 100);
    }

    #[test]
    fn missing_critical_variables() {
        let map = env_map(&[]);
        let err = load(&map).unwrap_err();
        assert_eq!(
            err,
            ConfigError::Missing(
                "DATABASE_URL (or DB_HOST, DB_NAME, DB_USER), JWT_SECRET".to_string()
            )
        );
    }

    #[test]
    fn jwt_duration_parsing() {
        let mut map = minimal_env();
        map.insert("JWT_DURATION".to_string(), "banana".to_string());
        assert!(load(&map).is_err());
        map.insert("JWT_DURATION".to_string(), "0".to_string());
        assert!(load(&map).is_err());
        map.insert("JWT_DURATION".to_string(), "12h".to_string());
        let cfg = load(&map).unwrap();
        assert_eq!(cfg.jwt_duration, std::time::Duration::from_secs(12 * 3600));
        map.insert("JWT_DURATION".to_string(), "90s".to_string());
        assert_eq!(
            load(&map).unwrap().jwt_duration,
            std::time::Duration::from_secs(90)
        );
    }

    #[test]
    fn email_lists_ported() {
        let fallback = vec!["info@voxire.com"];
        assert_eq!(parse_email_list("", &fallback), fallback);
        assert_eq!(parse_email_list(" \t ", &fallback), fallback);
        assert_eq!(
            parse_email_list("a@voxire.com", &fallback),
            vec!["a@voxire.com"]
        );
        assert_eq!(
            parse_email_list(" A@Voxire.com , b@voxire.com ", &fallback),
            vec!["a@voxire.com", "b@voxire.com"]
        );
        assert_eq!(
            parse_email_list("a@voxire.com,,a@voxire.com,", &fallback),
            vec!["a@voxire.com"]
        );
        assert_eq!(
            parse_email_list("b@voxire.com,a@voxire.com,b@voxire.com", &fallback),
            vec!["b@voxire.com", "a@voxire.com"]
        );
        assert_eq!(
            DEFAULT_LEAD_NOTIFY_EMAILS,
            [
                "info@voxire.com",
                "sales@voxire.com",
                "mohammad.saab@voxire.com",
                "mohamad.homsi@voxire.com",
                "amounehabed@gmail.com",
            ]
        );
        assert_eq!(DEFAULT_NEWSLETTER_NOTIFY_EMAILS, ["info@voxire.com"]);
    }

    #[test]
    fn notify_lists_load() {
        let mut map = minimal_env();
        map.insert(
            "LEAD_NOTIFY_EMAILS".to_string(),
            "x@voxire.com, y@voxire.com".to_string(),
        );
        let cfg = load(&map).unwrap();
        assert_eq!(
            cfg.lead_notify_emails,
            vec!["x@voxire.com".to_string(), "y@voxire.com".to_string()]
        );
        assert_eq!(
            cfg.newsletter_notify_emails,
            DEFAULT_NEWSLETTER_NOTIFY_EMAILS
                .iter()
                .map(|s| (*s).to_string())
                .collect::<Vec<_>>()
        );
    }

    #[test]
    fn app_env_production_defaults_closed() {
        for (env, want_production) in [
            ("production", true),
            ("prod", true),
            ("PROD", true),
            ("Production", true),
            ("PRODUCTION", true),
            ("production-eu", true),
            (" prod ", true),
            ("   ", true),
            ("", true),
            ("banana", true),
            ("staging", true),
            ("dev", false),
            ("development", false),
            ("Development", false),
            (" development ", false),
            ("local", false),
            ("test", false),
            ("TEST", false),
        ] {
            let cfg = load(&{
                let mut map = minimal_env();
                map.insert("APP_ENV".to_string(), env.to_string());
                map
            })
            .unwrap();
            assert_eq!(cfg.is_production(), want_production, "APP_ENV={env:?}");
            assert_eq!(normalize_app_env(env), env.trim().to_ascii_lowercase());
        }
        assert_eq!(app_env_class(""), AppEnvClass::Empty);
        assert_eq!(app_env_class("dev"), AppEnvClass::NonProduction);
        assert_eq!(app_env_class("prod"), AppEnvClass::Production);
        assert_eq!(app_env_class("banana"), AppEnvClass::Unrecognized);
    }

    #[test]
    fn trino_configuration_ported() {
        let mut map = minimal_env();
        assert!(!load(&map).unwrap().trino_configured());
        map.insert("TRINO_ENABLED".to_string(), "true".to_string());
        map.insert("TRINO_HOST".to_string(), "trino.example.com".to_string());
        assert!(!load(&map).unwrap().trino_configured());
        map.insert("TRINO_PORT".to_string(), "8443".to_string());
        map.insert("TRINO_USER".to_string(), "analytics".to_string());
        map.insert("TRINO_CATALOG".to_string(), "iceberg".to_string());
        map.insert("TRINO_SCHEMA".to_string(), "rtylr".to_string());
        map.insert("TRINO_TLS".to_string(), "true".to_string());
        let cfg = load(&map).unwrap();
        assert!(cfg.trino_configured());
        assert_eq!(cfg.trino_host(), "trino.example.com");
        assert_eq!(cfg.trino_port(), "8443");
        assert_eq!(cfg.trino_user, "analytics");
        assert_eq!(cfg.trino_catalog, "iceberg");
        assert_eq!(cfg.trino_schema, "rtylr");
        assert!(cfg.trino_tls);
        assert_eq!(cfg.trino_source(), "onion");
    }

    #[test]
    fn trusted_proxy_cidr_matching() {
        let get = |key: &str| -> Option<String> {
            if key == "TRUSTED_PROXY_CIDRS" {
                Some("203.0.113.0/24,2001:db8::/32".to_string())
            } else {
                None
            }
        };
        let cidrs = get_trusted_proxy_cidrs(&get);
        assert!(is_trusted_proxy_ip("203.0.113.7:443", &cidrs));
        assert!(is_trusted_proxy_ip("[2001:db8::7]:443", &cidrs));
        assert!(!is_trusted_proxy_ip("198.51.100.7:443", &cidrs));
        assert!(!is_trusted_proxy_ip("203.0.114.7:443", &cidrs));
        let none = get_trusted_proxy_cidrs(&|key| {
            if key == "TRUSTED_PROXY_CIDRS" {
                Some("none".to_string())
            } else {
                None
            }
        });
        assert!(none.is_empty());
    }

    #[test]
    fn listen_port_ported() {
        let get = |key: &str| -> Option<String> {
            if key == "PORT" {
                Some("4242".to_string())
            } else {
                None
            }
        };
        assert_eq!(listen_port(&get, "9999"), "4242");
        assert_eq!(listen_port(&|_| None, "8080"), "8080");
    }

    #[test]
    fn ai_configuration_ported() {
        let mut map = minimal_env();
        let cfg = load(&map).unwrap();
        assert!(!cfg.is_any_ai_configured());
        map.insert("OPENCODE_API_KEY".to_string(), " \t".to_string());
        assert!(!load(&map).unwrap().is_opencode_configured());
        map.insert("OPENCODE_API_KEY".to_string(), " test-zen-key ".to_string());
        let cfg = load(&map).unwrap();
        assert!(cfg.is_any_ai_configured());
        assert_eq!(cfg.opencode_api_key(), "test-zen-key");
        map.insert(
            "SEARXNG_URL".to_string(),
            " http://search.internal:8080 ".to_string(),
        );
        assert_eq!(
            load(&map).unwrap().searxng_url,
            "http://search.internal:8080"
        );
    }
}
