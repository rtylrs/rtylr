//! Resend email, mirroring `internal/platform/email/email.go`.

use serde::{Deserialize, Serialize};
use thiserror::Error;

pub const MIN_OUTBOUND_INTERVAL_MILLIS: u64 = 100;

#[derive(Debug, Error, PartialEq)]
pub enum EmailError {
    #[error("email sending disabled")]
    Disabled,
    #[error("email provider rate limited the request; retry after {0:?}: {1}")]
    RateLimited(std::time::Duration, String),
    #[error("resend api returned {0}: {1}")]
    ApiError(String, String),
    #[error("transport: {0}")]
    Transport(String),
}

pub struct SendParams {
    pub to: Vec<String>,
    pub subject: String,
    pub html: String,
    pub text: String,
    pub from: String,
    pub reply_to: String,
}

#[derive(Debug, Clone, Serialize)]
pub struct SendEmailRequest {
    pub from: String,
    pub to: Vec<String>,
    pub subject: String,
    #[serde(default, skip_serializing_if = "String::is_empty")]
    pub html: String,
    #[serde(default, skip_serializing_if = "String::is_empty")]
    pub text: String,
    #[serde(default, skip_serializing_if = "String::is_empty", rename = "reply_to")]
    pub reply_to: String,
}

#[derive(Debug, Clone, Deserialize)]
pub struct SendEmailResponse {
    #[serde(default)]
    pub data: Option<EmailData>,
    #[serde(default, rename = "error")]
    pub error: Option<EmailErrorBody>,
}

#[derive(Debug, Clone, Deserialize)]
pub struct EmailData {
    #[serde(default)]
    pub id: String,
}

#[derive(Debug, Clone, Deserialize)]
pub struct EmailErrorBody {
    #[serde(default, rename = "type")]
    pub error_type: String,
    #[serde(default)]
    pub code: String,
    #[serde(default)]
    pub detail: String,
}

pub fn default_from(from: &str) -> String {
    if from.is_empty() {
        "rtylr <noreply@voxire.com>".to_string()
    } else {
        from.to_string()
    }
}

pub fn is_enabled(api_key: &str) -> bool {
    !api_key.is_empty()
}

pub struct ResendClient {
    api_key: String,
    timeout: std::time::Duration,
    next_slot_millis: std::sync::atomic::AtomicU64,
}

impl ResendClient {
    pub fn new(api_key: &str) -> Self {
        Self {
            api_key: api_key.to_string(),
            timeout: std::time::Duration::from_secs(30),
            next_slot_millis: std::sync::atomic::AtomicU64::new(0),
        }
    }

    /// Pace outbound sends so concurrent calls space out by
    /// `MIN_OUTBOUND_INTERVAL_MILLIS` (Go `waitForOutboundSlot`).
    async fn wait_for_outbound_slot(&self) {
        let now_ms = std::time::SystemTime::now()
            .duration_since(std::time::UNIX_EPOCH)
            .map(|d| d.as_millis().min(u64::MAX as u128) as u64)
            .unwrap_or(0);
        let scheduled = self
            .next_slot_millis
            .fetch_update(
                std::sync::atomic::Ordering::SeqCst,
                std::sync::atomic::Ordering::SeqCst,
                |current| {
                    let start = if current > now_ms { current } else { now_ms };
                    Some(start.saturating_add(MIN_OUTBOUND_INTERVAL_MILLIS))
                },
            )
            .map(|prev| prev.max(now_ms))
            .unwrap_or(now_ms);
        if scheduled > now_ms {
            tokio::time::sleep(std::time::Duration::from_millis(scheduled - now_ms)).await;
        }
    }

    pub async fn send(&self, params: &SendParams) -> Result<String, EmailError> {
        if !is_enabled(&self.api_key) {
            return Err(EmailError::Disabled);
        }
        self.wait_for_outbound_slot().await;
        let request = SendEmailRequest {
            from: default_from(&params.from),
            to: params.to.clone(),
            subject: params.subject.clone(),
            html: params.html.clone(),
            text: params.text.clone(),
            reply_to: params.reply_to.clone(),
        };
        let client = reqwest::Client::builder()
            .timeout(self.timeout)
            .build()
            .map_err(|error| EmailError::Transport(error.to_string()))?;
        let response = client
            .post("https://api.resend.com/emails")
            .header("Authorization", format!("Bearer {}", self.api_key))
            .json(&request)
            .send()
            .await
            .map_err(|error| EmailError::Transport(error.to_string()))?;
        let status = response.status();
        // Read `Retry-After` before consuming the body, mirroring Go's
        // `resend.RateLimitError.RetryAfter` handling.
        let retry_after = response
            .headers()
            .get("Retry-After")
            .and_then(|value| value.to_str().ok())
            .and_then(|value| value.trim().parse::<u64>().ok())
            .filter(|seconds| *seconds > 0);
        if status.as_u16() == 429 {
            return Err(EmailError::RateLimited(
                std::time::Duration::from_secs(retry_after.unwrap_or(60)),
                "resend rate limit".to_string(),
            ));
        }
        let body = response
            .bytes()
            .await
            .map_err(|error| EmailError::Transport(error.to_string()))?;
        let parsed: SendEmailResponse = serde_json::from_slice(&body)
            .map_err(|error| EmailError::Transport(error.to_string()))?;
        if let Some(error) = parsed.error {
            if error.code == "rate_limit" {
                return Err(EmailError::RateLimited(
                    std::time::Duration::from_secs(retry_after.unwrap_or(60)),
                    format!("{}: {}", error.error_type, error.detail),
                ));
            }
            return Err(EmailError::ApiError(status.to_string(), error.detail));
        }
        if !status.is_success() {
            return Err(EmailError::ApiError(status.to_string(), String::new()));
        }
        Ok(parsed.data.map(|data| data.id).unwrap_or_default())
    }
}

#[cfg(test)]
mod tests {
    use super::{default_from, is_enabled};

    #[test]
    fn defaults_and_disabled() {
        assert_eq!(default_from(""), "rtylr <noreply@voxire.com>");
        assert_eq!(default_from("a@b"), "a@b");
        assert!(is_enabled("key"));
        assert!(!is_enabled(""));
    }
}
