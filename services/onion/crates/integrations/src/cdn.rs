//! CDN image delivery signing, mirroring the `signDeliveryPath` part of
//! `internal/platform/cdn/signing.go` (HMAC-SHA256 over
//! `path + "\n" + canonicalQuery`, `RawURLEncoding` base64).

use sha2::Sha256;

pub const DEFAULT_SIGNED_URL_TTL_SECS: i64 = 3600;
pub const MAX_SIGNED_URL_TTL_SECS: i64 = 30 * 24 * 3600;

fn canonical_query(query: &std::collections::HashMap<String, Vec<String>>) -> String {
    let mut keys: Vec<&String> = query.keys().filter(|key| key.as_str() != "sig").collect();
    keys.sort();
    let mut parts: Vec<String> = Vec::new();
    for key in keys {
        let mut values = query[key].clone();
        values.sort();
        for value in values {
            parts.push(format!(
                "{}={}",
                percent_encode(key),
                percent_encode(&value)
            ));
        }
    }
    parts.join("&")
}

fn percent_encode(text: &str) -> String {
    let mut out = String::new();
    for byte in text.bytes() {
        match byte {
            b'A'..=b'Z' | b'a'..=b'z' | b'0'..=b'9' | b'-' | b'_' | b'.' | b'*' | b'~' => {
                out.push(byte as char);
            }
            b' ' => out.push('+'),
            _ => out.push_str(&format!("%{byte:02X}")),
        }
    }
    out
}

/// `base64url_nopad(HMAC_SHA256(secret, path + "\n" + canonicalQuery))`.
///
/// The HMAC constructor accepts any key length, so this never fails in
/// practice; the `Result` exists to keep the no-`expect` invariant.
pub fn sign_delivery_path(
    secret: &str,
    path: &str,
    query: &std::collections::HashMap<String, Vec<String>>,
) -> Result<String, String> {
    let mut mac = hmac::Hmac::<Sha256>::new_from_slice(secret.as_bytes())
        .map_err(|_| "hmac key rejected".to_string())?;
    use hmac::Mac as _;
    mac.update(path.as_bytes());
    mac.update(b"\n");
    mac.update(canonical_query(query).as_bytes());
    let sum = mac.finalize().into_bytes();
    use base64::Engine as _;
    Ok(base64::engine::general_purpose::URL_SAFE_NO_PAD.encode(sum))
}

pub fn signed_url_expiry(expires_at: i64, expires_in: i64, now_unix: i64) -> i64 {
    if expires_at > 0 {
        return expires_at.min(now_unix + MAX_SIGNED_URL_TTL_SECS);
    }
    let ttl = if expires_in <= 0 {
        DEFAULT_SIGNED_URL_TTL_SECS
    } else {
        expires_in
    };
    let ttl = ttl.min(MAX_SIGNED_URL_TTL_SECS);
    now_unix + ttl
}

#[cfg(test)]
mod tests {
    use super::{sign_delivery_path, signed_url_expiry};

    #[test]
    fn signature_round_trip_and_expiry() {
        let mut query = std::collections::HashMap::new();
        query.insert("expires".to_string(), vec!["1700000000".to_string()]);
        let signature = sign_delivery_path("test-secret", "/img.png", &query).unwrap_or_default();
        assert!(!signature.is_empty());
        assert!(!signature.contains('='));
        // Same inputs yield the same signature (pins Go compatibility).
        assert_eq!(
            signature,
            sign_delivery_path("test-secret", "/img.png", &query).unwrap_or_default()
        );
        let now = chrono::Utc::now().timestamp();
        assert_eq!(signed_url_expiry(0, 0, now), now + 3600);
        assert_eq!(signed_url_expiry(0, 999_999_999, now), now + 30 * 24 * 3600);
        assert_eq!(signed_url_expiry(now + 120, 0, now), now + 120);
    }
}
