//! Paddle Billing, mirroring `internal/platform/paddle/client.go`.

use serde::{Deserialize, Serialize};
use sha2::Sha256;
use thiserror::Error;

const REQUEST_TIMEOUT_SECS: u64 = 15;
pub const MAX_PADDLE_RESPONSE_BODY_BYTES: usize = 1 << 20;
pub const WEBHOOK_SIGNATURE_TOLERANCE_SECS: i64 = 5 * 60;

#[derive(Debug, Error, PartialEq)]
pub enum PaddleError {
    #[error("paddle not configured")]
    NotConfigured,
    #[error("paddle idempotency key is required")]
    IdempotencyRequired,
    #[error("PADDLE_WEBHOOK_SECRET is required to verify webhook signatures")]
    WebhookSecretMissing,
    #[error("invalid Paddle-Signature header format")]
    BadSignatureHeader,
    #[error("invalid Paddle-Signature timestamp")]
    BadSignatureTimestamp,
    #[error("Paddle-Signature timestamp outside tolerance")]
    SignatureOutsideTolerance,
    #[error("webhook signature mismatch")]
    SignatureMismatch,
    #[error("paddle API returned {0}")]
    ApiStatus(String),
    #[error("paddle error {0}: {1}")]
    ApiError(String, String),
    #[error("paddle response body too large")]
    ResponseTooLarge,
    #[error("transport: {0}")]
    Transport(String),
}

#[derive(Debug, Clone, Serialize)]
pub struct TransactionItem {
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub price_id: Option<String>,
    pub quantity: i32,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub price: Option<OneOffPrice>,
}

#[derive(Debug, Clone, Serialize)]
pub struct OneOffPrice {
    pub description: String,
    pub unit_price: OneOffUnitPrice,
}

#[derive(Debug, Clone, Serialize)]
pub struct OneOffUnitPrice {
    pub amount: String,
    #[serde(rename = "currency_code")]
    pub currency_code: String,
}

#[derive(Debug, Clone, Serialize)]
pub struct CreateTransactionRequest {
    pub items: Vec<TransactionItem>,
    #[serde(default, skip_serializing_if = "String::is_empty")]
    pub customer_email: String,
    #[serde(default)]
    pub custom_data: std::collections::HashMap<String, String>,
    pub collection_mode: String,
}

#[derive(Debug, Clone, Deserialize)]
pub struct TransactionResponse {
    #[serde(default)]
    pub data: TransactionData,
    #[serde(default, rename = "error")]
    pub error: Option<PaddleErrorBody>,
}

#[derive(Debug, Clone, Default, Deserialize)]
pub struct TransactionData {
    #[serde(default)]
    pub id: String,
    #[serde(default)]
    pub checkout: Checkout,
    #[serde(default)]
    pub status: String,
}

#[derive(Debug, Clone, Default, Deserialize)]
pub struct Checkout {
    #[serde(default)]
    pub url: String,
}

#[derive(Debug, Clone, Deserialize)]
pub struct PaddleErrorBody {
    #[serde(default, rename = "type")]
    pub error_type: String,
    #[serde(default)]
    pub code: String,
    #[serde(default)]
    pub detail: String,
}

#[derive(Debug, Clone, Deserialize)]
pub struct WebhookEvent {
    #[serde(default, rename = "event_id")]
    pub event_id: String,
    #[serde(default, rename = "event_type")]
    pub event_type: String,
    #[serde(default, rename = "occurred_at")]
    pub occurred_at: String,
    #[serde(default, rename = "data")]
    pub data: serde_json::Value,
}

pub fn is_configured(api_key: &str) -> bool {
    !api_key.is_empty()
}

/// `Paddle-Signature: ts=<unix>;h1=<hex>`, h1 =
/// `lower_hex(HMAC_SHA256(secret, "ts:" + raw_body))`, ±5m tolerance.
pub fn verify_webhook_signature(
    raw_body: &[u8],
    signature_header: &str,
    secret: &str,
    allow_insecure_non_production: bool,
) -> Result<(), PaddleError> {
    if secret.is_empty() {
        if allow_insecure_non_production {
            return Ok(());
        }
        return Err(PaddleError::WebhookSecretMissing);
    }
    let (ts, h1) = signature_header.split(';').fold(
        (String::new(), String::new()),
        |(mut ts, mut h1), part| {
            if let Some(value) = part.strip_prefix("ts=") {
                ts = value.to_string();
            }
            if let Some(value) = part.strip_prefix("h1=") {
                h1 = value.to_string();
            }
            (ts, h1)
        },
    );
    if ts.is_empty() || h1.is_empty() {
        return Err(PaddleError::BadSignatureHeader);
    }
    let timestamp: i64 = ts.parse().map_err(|_| PaddleError::BadSignatureTimestamp)?;
    let now = chrono::Utc::now().timestamp();
    let delta = now - timestamp;
    if !(-WEBHOOK_SIGNATURE_TOLERANCE_SECS..=WEBHOOK_SIGNATURE_TOLERANCE_SECS).contains(&delta) {
        return Err(PaddleError::SignatureOutsideTolerance);
    }
    let mut mac = hmac::Hmac::<Sha256>::new_from_slice(secret.as_bytes())
        .map_err(|_| PaddleError::WebhookSecretMissing)?;
    use hmac::Mac as _;
    mac.update(format!("{ts}:").as_bytes());
    mac.update(raw_body);
    let expected = hex_lower(&mac.finalize().into_bytes());
    if expected.len() != h1.len() || !constant_time_eq(expected.as_bytes(), h1.as_bytes()) {
        return Err(PaddleError::SignatureMismatch);
    }
    Ok(())
}

fn hex_lower(bytes: &[u8]) -> String {
    const HEX: &[u8; 16] = b"0123456789abcdef";
    let mut out = String::with_capacity(bytes.len() * 2);
    for byte in bytes {
        out.push(HEX[(byte >> 4) as usize] as char);
        out.push(HEX[(byte & 0x0f) as usize] as char);
    }
    out
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

#[derive(Debug, Clone)]
pub struct PosPaymentParams<'a> {
    pub owner_id: i64,
    pub order_id: i64,
    pub amount_minor: i64,
    pub currency: &'a str,
    pub customer_email: &'a str,
    pub description: &'a str,
    pub idempotency_key: &'a str,
}

pub struct PaddleClient {
    base_url: String,
    api_key: String,
    timeout: std::time::Duration,
}

impl PaddleClient {
    pub fn new(base_url: &str, api_key: &str) -> Self {
        Self {
            base_url: base_url.trim_end_matches('/').to_string(),
            api_key: api_key.to_string(),
            timeout: std::time::Duration::from_secs(REQUEST_TIMEOUT_SECS),
        }
    }

    pub fn create_transaction_request(
        &self,
        business_id: i64,
        price_id: &str,
        customer_email: &str,
    ) -> CreateTransactionRequest {
        let mut custom_data = std::collections::HashMap::new();
        custom_data.insert("business_id".to_string(), business_id.to_string());
        CreateTransactionRequest {
            items: vec![TransactionItem {
                price_id: Some(price_id.to_string()),
                quantity: 1,
                price: None,
            }],
            customer_email: customer_email.to_string(),
            custom_data,
            collection_mode: "automatic".to_string(),
        }
    }

    pub fn pos_payment_request(
        &self,
        owner_id: i64,
        order_id: i64,
        amount_minor: i64,
        currency: &str,
        customer_email: &str,
        description: &str,
    ) -> CreateTransactionRequest {
        let mut custom_data = std::collections::HashMap::new();
        custom_data.insert("owner_id".to_string(), owner_id.to_string());
        custom_data.insert("order_id".to_string(), order_id.to_string());
        CreateTransactionRequest {
            items: vec![TransactionItem {
                price_id: None,
                quantity: 1,
                price: Some(OneOffPrice {
                    description: description.to_string(),
                    unit_price: OneOffUnitPrice {
                        amount: amount_minor.to_string(),
                        currency_code: currency.to_uppercase(),
                    },
                }),
            }],
            customer_email: customer_email.to_string(),
            custom_data,
            collection_mode: "automatic".to_string(),
        }
    }

    async fn post_transaction(
        &self,
        request: &CreateTransactionRequest,
        idempotency: Option<&str>,
    ) -> Result<TransactionResponse, PaddleError> {
        let client = reqwest::Client::builder()
            .timeout(self.timeout)
            .build()
            .map_err(|error| PaddleError::Transport(error.to_string()))?;
        let mut request = client
            .post(format!("{}/transactions", self.base_url))
            .header("Content-Type", "application/json")
            .header("Authorization", format!("Bearer {}", self.api_key))
            .json(request);
        if let Some(key) = idempotency {
            request = request.header("Paddle-Idempotency-Key", key);
        }
        let response = request
            .send()
            .await
            .map_err(|error| PaddleError::Transport(error.to_string()))?;
        let status = response.status();
        let body = response
            .bytes()
            .await
            .map_err(|error| PaddleError::Transport(error.to_string()))?;
        if body.len() > MAX_PADDLE_RESPONSE_BODY_BYTES {
            return Err(PaddleError::ResponseTooLarge);
        }
        if !(status.is_success()) {
            return Err(PaddleError::ApiStatus(status.to_string()));
        }
        let parsed: TransactionResponse = serde_json::from_slice(&body)
            .map_err(|error| PaddleError::Transport(error.to_string()))?;
        if let Some(error) = parsed.error {
            return Err(PaddleError::ApiError(error.code, error.detail));
        }
        Ok(parsed)
    }

    pub async fn create_transaction(
        &self,
        business_id: i64,
        price_id: &str,
        customer_email: &str,
    ) -> Result<String, PaddleError> {
        if !is_configured(&self.api_key) {
            return Err(PaddleError::NotConfigured);
        }
        let request = self.create_transaction_request(business_id, price_id, customer_email);
        let response = self.post_transaction(&request, None).await?;
        Ok(response.data.checkout.url)
    }

    pub async fn create_pos_transaction(
        &self,
        params: PosPaymentParams<'_>,
    ) -> Result<(String, String), PaddleError> {
        if !is_configured(&self.api_key) {
            return Err(PaddleError::NotConfigured);
        }
        if params.idempotency_key.is_empty() {
            return Err(PaddleError::IdempotencyRequired);
        }
        let request = self.pos_payment_request(
            params.owner_id,
            params.order_id,
            params.amount_minor,
            params.currency,
            params.customer_email,
            params.description,
        );
        let response = self
            .post_transaction(&request, Some(params.idempotency_key))
            .await?;
        Ok((response.data.id, response.data.checkout.url))
    }
}

#[cfg(test)]
mod tests {
    use super::{is_configured, verify_webhook_signature, PaddleError};

    fn header(body: &[u8], secret: &str, ts: i64) -> String {
        use hmac::Mac;
        let mut mac = hmac::Hmac::<sha2::Sha256>::new_from_slice(secret.as_bytes()).unwrap();
        mac.update(format!("{ts}:").as_bytes());
        mac.update(body);
        let expected = mac
            .finalize()
            .into_bytes()
            .iter()
            .map(|b| format!("{b:02x}"))
            .collect::<Vec<_>>()
            .join("");
        format!("ts={ts};h1={expected}")
    }

    #[test]
    fn webhook_signature_round_trip() {
        let now = chrono::Utc::now().timestamp();
        let body = b"{\"event\":\"x\"}";
        let secret = "test-secret";
        let sig = header(body, secret, now);
        assert!(verify_webhook_signature(body, &sig, secret, false).is_ok());
        assert_eq!(
            verify_webhook_signature(body, &sig, "other", false),
            Err(PaddleError::SignatureMismatch)
        );
        assert_eq!(
            verify_webhook_signature(b"tampered", &sig, secret, false),
            Err(PaddleError::SignatureMismatch)
        );
        let stale = header(body, secret, now - 3600);
        assert_eq!(
            verify_webhook_signature(body, &stale, secret, false),
            Err(PaddleError::SignatureOutsideTolerance)
        );
        assert!(verify_webhook_signature(body, "bogus", secret, false).is_err());
    }

    #[test]
    fn secret_and_config_parity() {
        assert_eq!(
            verify_webhook_signature(b"x", "ts=1;h1=x", "", false),
            Err(PaddleError::WebhookSecretMissing)
        );
        assert!(verify_webhook_signature(b"x", "bogus", "", true).is_ok());
        assert!(is_configured("key"));
        assert!(!is_configured(""));
    }
}
