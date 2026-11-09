//! JWT helper, mirroring `internal/platform/utils/jwtutils/jwt_utils.go`.
//!
//! HS256 only (algorithm confusion rejected), issuer required, `exp`
//! required, URL-decoded before parsing. Empty tokens are invalid;
//! expired tokens surface as `JwtError::TokenExpired`, never `InvalidToken`.

use std::time::{SystemTime, UNIX_EPOCH};

use jsonwebtoken::{decode, encode, Algorithm, DecodingKey, EncodingKey, Header, Validation};
use serde::{Deserialize, Serialize};
use thiserror::Error;

#[derive(Debug, Error, PartialEq)]
pub enum JwtError {
    #[error("token is invalid")]
    InvalidToken,
    #[error("token has expired")]
    TokenExpired,
    #[error("token encode failed: {0}")]
    Encode(String),
}

impl From<jsonwebtoken::errors::Error> for JwtError {
    fn from(error: jsonwebtoken::errors::Error) -> Self {
        match *error.kind() {
            jsonwebtoken::errors::ErrorKind::ExpiredSignature => Self::TokenExpired,
            _ => Self::InvalidToken,
        }
    }
}

/// Claims issued for sessions and OAuth access tokens. JSON names mirror
/// Go's `AccessTokenClaims` / `OAuthTokenClaims` (`cid`, `token_epoch`).
#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
pub struct TokenClaims {
    #[serde(default, rename = "user_id")]
    pub user_id: i64,
    #[serde(default, rename = "business_id", skip_serializing_if = "is_zero_i64")]
    pub business_id: i64,
    #[serde(default, rename = "is_console_admin", skip_serializing_if = "is_false")]
    pub is_console_admin: bool,
    #[serde(default, rename = "platform", skip_serializing_if = "String::is_empty")]
    pub platform: String,
    #[serde(
        default,
        rename = "platform_user_id",
        skip_serializing_if = "is_zero_i64"
    )]
    pub platform_user_id: i64,
    #[serde(default, rename = "plan_id", skip_serializing_if = "is_zero_i64")]
    pub plan_id: i64,
    #[serde(default, rename = "token_epoch", skip_serializing_if = "is_zero_i64")]
    pub token_epoch: i64,
    #[serde(default, rename = "scope", skip_serializing_if = "String::is_empty")]
    pub scope: String,
    #[serde(default, rename = "cid", skip_serializing_if = "String::is_empty")]
    pub oauth_client_id: String,
    #[serde(default, rename = "iss", skip_serializing_if = "String::is_empty")]
    pub issuer: String,
    #[serde(default, rename = "exp", skip_serializing_if = "is_zero_i64")]
    pub expires_at: i64,
}

fn is_zero_i64(value: &i64) -> bool {
    *value == 0
}

fn is_false(value: &bool) -> bool {
    !*value
}

impl TokenClaims {
    pub fn new(user_id: i64) -> Self {
        Self {
            user_id,
            business_id: 0,
            is_console_admin: false,
            platform: String::new(),
            platform_user_id: 0,
            plan_id: 0,
            token_epoch: 0,
            scope: String::new(),
            oauth_client_id: String::new(),
            issuer: String::new(),
            expires_at: 0,
        }
    }
}

pub struct JwtHelper {
    pub secret: Vec<u8>,
    pub issuer: String,
    pub duration_secs: i64,
}

impl JwtHelper {
    pub fn new(secret: &str, issuer: &str) -> Self {
        Self {
            secret: secret.as_bytes().to_vec(),
            issuer: issuer.to_string(),
            duration_secs: 24 * 3600,
        }
    }

    pub fn create_token(&self, mut claims: TokenClaims) -> Result<String, JwtError> {
        claims.issuer = self.issuer.clone();
        if self.duration_secs > 0 {
            claims.expires_at = now_secs() + self.duration_secs;
        }
        encode(
            &Header::new(Algorithm::HS256),
            &claims,
            &EncodingKey::from_secret(&self.secret),
        )
        .map_err(|error| JwtError::Encode(error.to_string()))
    }

    pub fn validate_token(&self, token: &str) -> Result<TokenClaims, JwtError> {
        if token.is_empty() {
            return Err(JwtError::InvalidToken);
        }
        let decoded = percent_decode(token).ok_or(JwtError::InvalidToken)?;
        let mut validation = Validation::new(Algorithm::HS256);
        validation.iss = Some(
            [self.issuer.clone()]
                .into_iter()
                .collect::<std::collections::HashSet<String>>(),
        );
        validation.required_spec_claims.insert("exp".to_string());
        decode::<TokenClaims>(
            &decoded,
            &DecodingKey::from_secret(&self.secret),
            &validation,
        )
        .map(|data| data.claims)
        .map_err(JwtError::from)
    }
}

fn now_secs() -> i64 {
    SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .map(|d| d.as_secs() as i64)
        .unwrap_or(0)
}

/// Minimal %-decoder; `None` on malformed escapes (Go `PathUnescape`).
fn percent_decode(text: &str) -> Option<String> {
    let bytes = text.as_bytes();
    let mut out = Vec::with_capacity(bytes.len());
    let mut index = 0;
    while index < bytes.len() {
        if bytes[index] == b'%' {
            if index + 2 >= bytes.len() {
                return None;
            }
            let hex = (bytes[index + 1] as char)
                .to_digit(16)
                .zip((bytes[index + 2] as char).to_digit(16))?;
            out.push((hex.0 * 16 + hex.1) as u8);
            index += 3;
        } else {
            out.push(bytes[index]);
            index += 1;
        }
    }
    String::from_utf8(out).ok()
}

#[cfg(test)]
mod tests {
    use super::{JwtError, JwtHelper, TokenClaims};

    fn helper() -> JwtHelper {
        JwtHelper::new("test-secret", "rtylr")
    }

    #[test]
    fn round_trip_and_expiry() {
        let helper = helper();
        let claims = TokenClaims {
            user_id: 7,
            business_id: 42,
            ..TokenClaims::new(7)
        };
        let token = helper.create_token(claims).unwrap();
        let claims = helper.validate_token(&token).unwrap();
        assert_eq!(claims.user_id, 7);
        assert_eq!(claims.business_id, 42);
        assert_eq!(claims.issuer, "rtylr");
        assert!(claims.expires_at > 0);
    }

    #[test]
    fn rejects_bad_tokens() {
        let helper = helper();
        let token = helper.create_token(TokenClaims::new(7)).unwrap();
        assert_eq!(helper.validate_token(""), Err(JwtError::InvalidToken));
        assert_eq!(
            helper.validate_token("not-a-token"),
            Err(JwtError::InvalidToken)
        );
        // Tamper with the last character of the signature.
        let tampered = {
            let mut chars: Vec<char> = token.chars().collect();
            let last = chars.len() - 1;
            chars[last] = if chars[last] == 'a' { 'b' } else { 'a' };
            chars.into_iter().collect::<String>()
        };
        assert_eq!(
            helper.validate_token(&tampered),
            Err(JwtError::InvalidToken)
        );
        let other = JwtHelper::new("other-secret", "rtylr");
        assert_eq!(other.validate_token(&token), Err(JwtError::InvalidToken));
    }

    #[test]
    fn expired_token_is_distinct() {
        let mut helper = helper();
        helper.duration_secs = -1;
        // Negative duration encodes exp in the past only when > 0 is false:
        // Go leaves zero time (no exp) then, which must fail as invalid
        // because exp is required. Expired path is covered by construction:
        let claims = TokenClaims {
            user_id: 7,
            expires_at: now_secs_for_test() - 3600,
            ..TokenClaims::new(7)
        };
        let token = helper.create_token(claims).unwrap_or_default();
        assert_eq!(
            JwtHelper::new("test-secret", "rtylr").validate_token(&token),
            Err(JwtError::TokenExpired)
        );
    }

    fn now_secs_for_test() -> i64 {
        std::time::SystemTime::now()
            .duration_since(std::time::UNIX_EPOCH)
            .unwrap()
            .as_secs() as i64
    }

    #[test]
    fn percent_decoding() {
        assert_eq!(super::percent_decode("a%20b"), Some("a b".to_string()));
        assert_eq!(super::percent_decode("a%zz"), None);
    }
}
