//! AES-256-GCM email-body encryption at rest, mirroring
//! `internal/platform/emailcrypto/crypto.go`.
//!
//! Row shape: `enc:v1:` + base64(nonce || ciphertext). Unprefixed rows are
//! legacy plaintext and pass through unchanged. Encryption keys are read
//! from the environment on every use so rotation needs no restart.

use aes_gcm::aead::{Aead, KeyInit};
use aes_gcm::{Aes256Gcm, Nonce};
use base64::Engine as _;
use rand::RngCore;

pub const ENCRYPTED_PREFIX: &str = "enc:v1:";
pub const INVALID_ENCRYPTED_BODY: &str = "enc:v1:invalid-key";
pub const PREVIOUS_KEYS_ENV: &str = "EMAIL_ENCRYPTION_PREVIOUS_KEYS";

pub type EnvGet<'a> = &'a dyn Fn(&str) -> Option<String>;

#[derive(Debug, Clone, PartialEq)]
pub struct KeyConfig {
    pub keys: Vec<Vec<u8>>,
    pub configured: bool,
    pub invalid_current: bool,
}

pub fn parse_encryption_key(raw: &str) -> Option<Vec<u8>> {
    if let Ok(key) = base64::engine::general_purpose::STANDARD.decode(raw) {
        if key.len() == 32 {
            return Some(key);
        }
    }
    if let Ok(key) = decode_hex(raw) {
        if key.len() == 32 {
            return Some(key);
        }
    }
    None
}

pub fn decode_hex(text: &str) -> Result<Vec<u8>, String> {
    if !text.len().is_multiple_of(2) {
        return Err("odd-length hex".to_string());
    }
    let bytes = text.as_bytes();
    let mut out = Vec::with_capacity(text.len() / 2);
    for index in 0..text.len() / 2 {
        let hi = hex_val(bytes[index * 2]).ok_or_else(|| "bad hex".to_string())?;
        let lo = hex_val(bytes[index * 2 + 1]).ok_or_else(|| "bad hex".to_string())?;
        out.push(hi << 4 | lo);
    }
    Ok(out)
}

fn hex_val(byte: u8) -> Option<u8> {
    match byte {
        b'0'..=b'9' => Some(byte - b'0'),
        b'a'..=b'f' => Some(byte - b'a' + 10),
        b'A'..=b'F' => Some(byte - b'A' + 10),
        _ => None,
    }
}

pub fn encryption_keys(get: EnvGet<'_>) -> KeyConfig {
    let raw = get("EMAIL_ENCRYPTION_KEY").unwrap_or_default();
    let previous = get(PREVIOUS_KEYS_ENV).unwrap_or_default();
    if raw.trim().is_empty() && previous.trim().is_empty() {
        return KeyConfig {
            keys: vec![],
            configured: false,
            invalid_current: false,
        };
    }
    let mut keys = Vec::new();
    let mut seen = std::collections::HashSet::new();
    let mut invalid_current = false;
    if raw.trim().is_empty() {
        invalid_current = true;
    } else if let Some(current) = parse_encryption_key(raw.trim()) {
        seen.insert(current.clone());
        keys.push(current);
    } else {
        invalid_current = true;
    }
    for part in previous.split(',') {
        let part = part.trim();
        if part.is_empty() {
            continue;
        }
        if let Some(key) = parse_encryption_key(part) {
            if seen.insert(key.clone()) {
                keys.push(key);
            }
        }
    }
    KeyConfig {
        keys,
        configured: true,
        invalid_current,
    }
}

pub fn seal_html(html: &str, get: EnvGet<'_>) -> String {
    if html.is_empty() || html.starts_with(ENCRYPTED_PREFIX) {
        return html.to_string();
    }
    let config = encryption_keys(get);
    if !config.configured {
        return html.to_string();
    }
    if config.invalid_current || config.keys.is_empty() {
        return INVALID_ENCRYPTED_BODY.to_string();
    }
    let Some(first) = config.keys.first() else {
        return INVALID_ENCRYPTED_BODY.to_string();
    };
    let cipher = Aes256Gcm::new_from_slice(first);
    let Ok(cipher) = cipher else {
        return INVALID_ENCRYPTED_BODY.to_string();
    };
    let mut nonce_bytes = [0u8; 12];
    rand::thread_rng().fill_bytes(&mut nonce_bytes);
    let nonce = Nonce::from_slice(&nonce_bytes);
    let ciphertext = cipher.encrypt(nonce, html.as_bytes());
    let Ok(ciphertext) = ciphertext else {
        return INVALID_ENCRYPTED_BODY.to_string();
    };
    let mut sealed = nonce_bytes.to_vec();
    sealed.extend_from_slice(&ciphertext);
    format!(
        "{}{}",
        ENCRYPTED_PREFIX,
        base64::engine::general_purpose::STANDARD.encode(sealed)
    )
}

pub fn open_html(stored: &str, get: EnvGet<'_>) -> String {
    if !stored.starts_with(ENCRYPTED_PREFIX) {
        return stored.to_string();
    }
    if stored == INVALID_ENCRYPTED_BODY {
        return String::new();
    }
    let Ok(raw) = base64::engine::general_purpose::STANDARD
        .decode(stored.strip_prefix(ENCRYPTED_PREFIX).unwrap_or(""))
    else {
        return String::new();
    };
    let config = encryption_keys(get);
    if !config.configured || config.keys.is_empty() {
        return String::new();
    }
    for key in &config.keys {
        if let Some(plaintext) = open_with_key(&raw, key) {
            return plaintext;
        }
    }
    String::new()
}

fn open_with_key(raw: &[u8], key: &[u8]) -> Option<String> {
    if raw.len() < 12 {
        return None;
    }
    let cipher = Aes256Gcm::new_from_slice(key).ok()?;
    let (nonce, ciphertext) = raw.split_at(12);
    let plaintext = cipher.decrypt(Nonce::from_slice(nonce), ciphertext).ok()?;
    String::from_utf8(plaintext).ok()
}

#[cfg(test)]
mod tests {
    use super::{decode_hex, encryption_keys, open_html, parse_encryption_key, seal_html};

    fn env_map(pairs: &[(&str, &str)]) -> std::collections::HashMap<String, String> {
        pairs
            .iter()
            .map(|(key, value)| ((*key).to_string(), (*value).to_string()))
            .collect()
    }

    fn hex_key() -> String {
        "01".repeat(32)
    }

    #[test]
    fn keys_parse() {
        let hex = decode_hex(&"0f".repeat(32)).unwrap();
        assert_eq!(hex.len(), 32);
        assert!(decode_hex("abc").is_err());
        assert_eq!(parse_encryption_key(&hex_key()), Some(vec![0x01; 32]));
        assert_eq!(parse_encryption_key("nope"), None);
    }

    #[test]
    fn seal_round_trips() {
        let key_hex = hex_key();
        let map = env_map(&[("EMAIL_ENCRYPTION_KEY", &key_hex)]);
        let get = |key: &str| map.get(key).cloned();
        assert!(encryption_keys(&get).configured);
        let sealed = seal_html("<p>hi</p>", &get);
        assert!(sealed.starts_with("enc:v1:"));
        assert_eq!(seal_html(&sealed, &get), sealed);
        assert_eq!(open_html(&sealed, &get), "<p>hi</p>");
        assert_eq!(open_html("legacy plain", &get), "legacy plain");
        assert_eq!(open_html("enc:v1:invalid-key", &get), "");
    }

    #[test]
    fn plaintext_without_key() {
        let get = |_: &str| None;
        assert_eq!(seal_html("<p>hi</p>", &get), "<p>hi</p>");
        assert_eq!(open_html("legacy plain", &get), "legacy plain");
    }
}
