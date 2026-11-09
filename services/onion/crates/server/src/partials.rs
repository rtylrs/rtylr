//! Partials internal-identity bypass. Mirrors `partials_key.go`.

/// Request header Partials sends to identify itself to internal endpoints.
pub const PARTIALS_API_KEY_HEADER: &str = "X-API-KEY";
/// Env var (from Secrets Manager in production) holding the shared key.
pub const PARTIALS_API_KEY_ENV_VAR: &str = "PARTIALS_API_KEY";
/// The bypass acts as a console admin with this token string.
pub const PARTIALS_ACTOR_TOKEN: &str = "partials-api-key";

/// Report whether the request carries the configured internal key.
/// Returns false when no key is configured, so the bypass cannot be
/// triggered by accident. Comparison is constant-time.
pub fn has_valid_partials_api_key(header_value: Option<&str>, expected_key: &str) -> bool {
    let expected = expected_key.trim();
    if expected.is_empty() {
        return false;
    }
    let Some(actual) = header_value else {
        return false;
    };
    if actual.is_empty() {
        return false;
    }
    constant_time_eq(actual.as_bytes(), expected.as_bytes())
}

fn constant_time_eq(a: &[u8], b: &[u8]) -> bool {
    if a.len() != b.len() {
        return false;
    }
    let mut diff = 0u8;
    for i in 0..a.len() {
        diff |= a[i] ^ b[i];
    }
    diff == 0
}

#[cfg(test)]
mod tests {
    use super::has_valid_partials_api_key;

    #[test]
    fn bypass_predicate() {
        assert!(has_valid_partials_api_key(Some("secret"), "secret"));
        assert!(!has_valid_partials_api_key(Some("wrong"), "secret"));
        assert!(!has_valid_partials_api_key(Some("secret"), ""));
        assert!(!has_valid_partials_api_key(None, "secret"));
        assert!(!has_valid_partials_api_key(Some(""), "secret"));
        // Length-mismatched inputs still fail in constant-time shape.
        assert!(!has_valid_partials_api_key(Some("s"), "secret"));
    }
}
