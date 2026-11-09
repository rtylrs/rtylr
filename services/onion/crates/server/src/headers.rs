//! Static security response headers. Mirrors `security_headers.go`.
//!
//! Deliberate cross-origin choices: this server is an API + CDN consumed
//! cross-site by `*.rtylr.com` / `*.voxire.com`, so CORP is `cross-origin`
//! and COOP allows popups for the Google sign-in flow. COEP and a same-site
//! CORP are intentionally omitted so legitimate cross-origin embedding does
//! not break.

/// The static hardening set applied to every response, in audit order.
pub fn security_response_headers() -> Vec<(String, String)> {
    vec![
        ("X-Content-Type-Options".into(), "nosniff".into()),
        ("X-Frame-Options".into(), "DENY".into()),
        // Disabled (0) per current OWASP guidance: the legacy auditor it
        // enabled is removed from modern browsers and can leak information
        // via side channels in old ones.
        ("X-XSS-Protection".into(), "0".into()),
        (
            "Strict-Transport-Security".into(),
            "max-age=31536000; includeSubDomains; preload".into(),
        ),
        (
            "Referrer-Policy".into(),
            "strict-origin-when-cross-origin".into(),
        ),
        (
            "Permissions-Policy".into(),
            "accelerometer=(), autoplay=(), browsing-topics=(), camera=(), display-capture=(), geolocation=(), gyroscope=(), magnetometer=(), microphone=(), payment=(), usb=()".into(),
        ),
        ("X-Permitted-Cross-Domain-Policies".into(), "none".into()),
        (
            "Cross-Origin-Opener-Policy".into(),
            "same-origin-allow-popups".into(),
        ),
        ("Cross-Origin-Resource-Policy".into(), "cross-origin".into()),
        (
            "Content-Security-Policy".into(),
            "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' data:; connect-src 'self' https:; frame-ancestors 'none'; base-uri 'self'; form-action 'self'; object-src 'none'".into(),
        ),
    ]
}

/// Header names the Go middleware deletes from every response.
pub fn deleted_response_headers() -> [&'static str; 2] {
    ["Server", "X-Powered-By"]
}

#[cfg(test)]
mod tests {
    use super::{deleted_response_headers, security_response_headers};

    #[test]
    fn headers() {
        let headers = security_response_headers();
        let get = |name: &str| {
            headers
                .iter()
                .find(|(key, _)| key == name)
                .map(|(_, value)| value.as_str())
        };
        assert_eq!(get("X-Content-Type-Options"), Some("nosniff"));
        assert_eq!(get("X-Frame-Options"), Some("DENY"));
        assert_eq!(
            get("Strict-Transport-Security"),
            Some("max-age=31536000; includeSubDomains; preload")
        );
        assert_eq!(get("Cross-Origin-Resource-Policy"), Some("cross-origin"));
        assert!(get("Content-Security-Policy")
            .is_some_and(|value| value.contains("frame-ancestors 'none'")));
        assert_eq!(deleted_response_headers(), ["Server", "X-Powered-By"]);
    }
}
