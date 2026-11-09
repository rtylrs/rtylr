//! Opaque API keys, mirroring `internal/platform/securekey/key.go`.

use rand::RngCore;
use sha2::{Digest, Sha256};
use thiserror::Error;

pub const DEFAULT_RANDOM_BYTES: usize = 32;

#[derive(Debug, Error, PartialEq)]
pub enum KeyError {
    #[error("generate key: {0}")]
    Entropy(String),
}

#[derive(Debug, Clone, PartialEq)]
pub struct Secret {
    pub plaintext: String,
    pub hash: String,
    pub display_prefix: String,
}

#[derive(Debug, Clone, Default, PartialEq)]
pub struct Options {
    pub prefix: String,
    pub random_bytes: usize,
    pub display_prefix_len: usize,
}

pub fn hash(secret: &str) -> String {
    let digest = Sha256::digest(secret.trim().as_bytes());
    hex_encode(&digest)
}

pub fn generate(options: Options) -> Result<Secret, KeyError> {
    let random_bytes = if options.random_bytes == 0 {
        DEFAULT_RANDOM_BYTES
    } else {
        options.random_bytes
    };
    let mut raw = vec![0u8; random_bytes];
    rand::thread_rng()
        .try_fill_bytes(&mut raw)
        .map_err(|error| KeyError::Entropy(error.to_string()))?;
    let plaintext = format!("{}{}", options.prefix, hex_encode(&raw));
    let display_len =
        if options.display_prefix_len == 0 || options.display_prefix_len > plaintext.len() {
            plaintext.len().min(12)
        } else {
            options.display_prefix_len
        };
    Ok(Secret {
        display_prefix: plaintext[..display_len].to_string(),
        hash: hash(&plaintext),
        plaintext,
    })
}

fn hex_encode(bytes: &[u8]) -> String {
    const HEX: &[u8; 16] = b"0123456789abcdef";
    let mut out = String::with_capacity(bytes.len() * 2);
    for byte in bytes {
        out.push(HEX[(byte >> 4) as usize] as char);
        out.push(HEX[(byte & 0x0f) as usize] as char);
    }
    out
}

#[cfg(test)]
mod tests {
    use super::{generate, hash, Options};

    #[test]
    fn generate_and_hash() {
        let secret = generate(Options::default()).unwrap();
        assert_eq!(secret.plaintext.len(), 64);
        assert_eq!(secret.hash, hash(&secret.plaintext));
        assert_eq!(secret.display_prefix.len(), 12);
        let prefixed = generate(Options {
            prefix: "pk_".to_string(),
            random_bytes: 0,
            display_prefix_len: 0,
        })
        .unwrap();
        assert!(prefixed.plaintext.starts_with("pk_"));
        assert!(prefixed.display_prefix.starts_with("pk_"));
        let trimmed = generate(Options {
            prefix: String::new(),
            random_bytes: 0,
            display_prefix_len: 99,
        })
        .unwrap();
        assert_eq!(trimmed.display_prefix.len(), 12);
    }

    #[test]
    fn hash_trims() {
        assert_eq!(hash("  abc  "), hash("abc"));
    }
}
