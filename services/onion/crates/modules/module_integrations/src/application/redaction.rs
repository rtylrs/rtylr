//! Secret redaction for stored configs. Mirrors the `RedactConfig` part
//! of `internal/application/service.go`: a listing never hands credentials
//! back to the browser.

use serde_json::Value;

pub const SECRET_KEY_FRAGMENTS: [&str; 8] = [
    "secret",
    "token",
    "key",
    "password",
    "passwd",
    "credential",
    "authorization",
    "private",
];

/// Whether a config key looks like it holds a secret (case-insensitive
/// substring match, as Go's `strings.Contains` on the lowered key).
pub fn is_secret_key(key: &str) -> bool {
    let lowered = key.to_ascii_lowercase();
    SECRET_KEY_FRAGMENTS
        .iter()
        .any(|fragment| lowered.contains(fragment))
}

fn redact_value(value: Value) -> Value {
    match value {
        Value::Object(map) => {
            let mut out = serde_json::Map::new();
            for (key, child) in map {
                if is_secret_key(&key) {
                    out.insert(key, Value::String("[redacted]".into()));
                } else {
                    out.insert(key, redact_value(child));
                }
            }
            Value::Object(out)
        }
        Value::Array(items) => Value::Array(items.into_iter().map(redact_value).collect()),
        other => other,
    }
}

/// Masks secret-looking values anywhere in a stored config. Invalid or
/// empty JSON yields `{}` (Go returns `json.RawMessage("{}")`).
pub fn redact_config(raw: &Value) -> Value {
    match raw {
        Value::Null => serde_json::json!({}),
        Value::Object(_) | Value::Array(_) => redact_value(raw.clone()),
        _ => serde_json::json!({}),
    }
}

#[cfg(test)]
mod tests {
    use super::{is_secret_key, redact_config};
    use serde_json::Value;

    #[test]
    fn key_detection() {
        assert!(is_secret_key("client_secret"));
        assert!(is_secret_key("apiKey"));
        assert!(is_secret_key("password"));
        assert!(is_secret_key("Authorization"));
        assert!(!is_secret_key("business_id"));
    }

    #[test]
    fn nested_redaction() {
        let config = serde_json::json!({
            "business_id": 7,
            "stripe": {"publishable": "pk_test_1", "secret_key": "sk_test_2"},
            "nested": [{"token": "abc"}, {"safe": 1}]
        });
        let redacted = redact_config(&config);
        assert_eq!(redacted["business_id"], 7);
        assert_eq!(redacted["stripe"]["publishable"], "pk_test_1");
        assert_eq!(redacted["stripe"]["secret_key"], "[redacted]");
        assert_eq!(redacted["nested"][0]["token"], "[redacted]");
        assert_eq!(redacted["nested"][1]["safe"], 1);
    }

    #[test]
    fn non_object_falls_back() {
        assert_eq!(
            redact_config(&serde_json::json!("x")),
            serde_json::json!({})
        );
        assert_eq!(redact_config(&Value::Null), serde_json::json!({}));
    }
}
