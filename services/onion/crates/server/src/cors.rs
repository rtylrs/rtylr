//! CORS allowlist, mirrored request headers, and preflight. Mirrors
//! `cors.go`; origin matching comes from the `auth` crate
//! (`origins::is_allowed`).

/// Headers a cross-origin caller may send (`allowedHeaders`).
pub fn allowed_headers() -> Vec<&'static str> {
    vec![
        "Authorization",
        "Content-Type",
        "Idempotency-Key",
        "Accept",
        "Origin",
        "User-Agent",
        "Referer",
        "X-Requested-With",
        "Access-Control-Allow-Origin",
        "Access-Control-Allow-Credentials",
        "Access-Control-Allow-Headers",
        "Access-Control-Allow-Methods",
        "Access-Control-Expose-Headers",
        "x-request-id",
        "profile_token",
        "client_platform",
        "X-Product-Platform",
        "Product-Platform",
        "client_version",
        "language",
        "client_ip",
    ]
}

/// Response headers for one request (Go sets methods/headers/expose/max-age
/// unconditionally and reflects the origin when allowed). `vary: Origin` is
/// always added so a shared cache cannot serve one origin's allow-origin to
/// another.
pub fn cors_response_headers(origin: Option<&str>, is_production: bool) -> Vec<(String, String)> {
    let mut headers = vec![("Vary".to_string(), "Origin".to_string())];
    if let Some(origin) = origin {
        if auth::origins::is_allowed(origin, is_production) {
            headers.push((
                "Access-Control-Allow-Origin".to_string(),
                origin.to_string(),
            ));
            headers.push((
                "Access-Control-Allow-Credentials".to_string(),
                "true".to_string(),
            ));
        }
    }
    headers.push((
        "Access-Control-Allow-Methods".to_string(),
        "GET, POST, PUT, DELETE, PATCH, OPTIONS".to_string(),
    ));
    headers.push((
        "Access-Control-Allow-Headers".to_string(),
        allowed_headers().join(", "),
    ));
    headers.push((
        "Access-Control-Expose-Headers".to_string(),
        "Content-Length".to_string(),
    ));
    headers.push(("Access-Control-Max-Age".to_string(), "600".to_string()));
    headers
}

/// `OPTIONS` preflights short-circuit with 200.
pub fn is_preflight(method: &str) -> bool {
    method == "OPTIONS"
}

#[cfg(test)]
mod tests {
    use super::{allowed_headers, cors_response_headers, is_preflight};

    #[test]
    fn headers() {
        assert!(allowed_headers().contains(&"X-Product-Platform"));
        let headers = cors_response_headers(Some("http://localhost:3000"), false);
        assert!(headers
            .iter()
            .any(|(name, value)| name == "Vary" && value == "Origin"));
        assert!(headers
            .iter()
            .any(|(name, value)| name == "Access-Control-Allow-Origin"
                && value == "http://localhost:3000"));
        assert!(headers
            .iter()
            .any(|(name, value)| name == "Access-Control-Max-Age" && value == "600"));
        assert!(is_preflight("OPTIONS"));
        assert!(!is_preflight("GET"));
    }

    #[test]
    fn disallowed_origin_not_reflected() {
        let headers = cors_response_headers(Some("http://evil.example"), true);
        assert!(!headers
            .iter()
            .any(|(name, _)| name == "Access-Control-Allow-Origin"));
    }
}
