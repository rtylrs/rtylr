//! Request body limits and the hard request timeout. Mirrors
//! `body_limit.go` and `query_timeout.go`.

/// Global ceiling for POST/PUT/PATCH bodies on the shared API server.
/// Routes that need a larger bound (multipart resumes, CDN uploads,
/// webhooks) are exempt and enforce their own `MaxBytesReader`.
pub const DEFAULT_MAX_REQUEST_BODY_BYTES: i64 = 4 << 20; // 4 MiB

/// Hard deadline for any single HTTP request; in-flight DB calls on
/// the request context cancel when it fires.
pub const REQUEST_TIMEOUT_SECS: u64 = 30;
/// Soft warn: requests completing after this are logged with
/// `slow_query=true`.
pub const SLOW_QUERY_THRESHOLD_SECS: u64 = 5;

/// Whether the body limit applies to this method.
pub fn body_limit_applies(method: &str) -> bool {
    matches!(method, "POST" | "PUT" | "PATCH")
}

/// Paths exempt from the global body limit (multipart resumes and the
/// CDN runmode paths).
pub fn request_body_limit_exempt(path: &str) -> bool {
    matches!(
        path,
        "/careers/apply" | "/voxhire/applications" | "/voxhire/careers/apply"
    ) || matches!(path, "/upload" | "/purge" | "/sign")
}

/// `isLongLivedRequest`: WS/SSE/Upgrade requests skip the 30s timeout.
pub fn is_long_lived_request(
    upgrade: Option<&str>,
    connection: Option<&str>,
    accept: Option<&str>,
    path: &str,
) -> bool {
    if upgrade.is_some_and(|value| value.eq_ignore_ascii_case("websocket")) {
        return true;
    }
    if connection.is_some_and(|value| value.to_lowercase().contains("upgrade")) {
        return true;
    }
    if accept.is_some_and(|value| value.contains("text/event-stream")) {
        return true;
    }
    // Browser fetch()-based SSE sends `Accept: */*`, so path matching
    // covers it (same rule as Go).
    path.ends_with("/generate-text-stream")
}

#[cfg(test)]
mod tests {
    use super::{body_limit_applies, is_long_lived_request, request_body_limit_exempt};

    #[test]
    fn limits() {
        assert!(body_limit_applies("POST"));
        assert!(!body_limit_applies("GET"));
        assert!(request_body_limit_exempt("/upload"));
        assert!(request_body_limit_exempt("/careers/apply"));
        assert!(!request_body_limit_exempt("/auth/login"));
        assert!(is_long_lived_request(Some("websocket"), None, None, "/ws"));
        assert!(is_long_lived_request(
            None,
            Some("keep-alive, Upgrade"),
            None,
            "/ws"
        ));
        assert!(is_long_lived_request(
            None,
            None,
            Some("text/event-stream"),
            "/ai"
        ));
        assert!(is_long_lived_request(
            None,
            None,
            Some("*/*"),
            "/ai/generate-text-stream"
        ));
        assert!(!is_long_lived_request(None, None, None, "/api"));
        assert!(!is_long_lived_request(
            None,
            Some("keep-alive"),
            None,
            "/api"
        ));
    }
}
