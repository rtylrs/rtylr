//! Google Sheets, mirroring `internal/platform/sheets/sheets.go`.
//!
//! Service accounts: a base64 JSON key is exchanged for a JWT grant,
//! the grant for an access token, then `append` POSTs race the row.
//! Sheets appends are best-effort (errors log, never propagate).

use base64::Engine as _;
use jsonwebtoken::{encode, Algorithm, EncodingKey, Header};
use serde_json::json;
use thiserror::Error;

pub const LEADS_SHEET_RANGE: &str = "Leads!A:J";
pub const SHEETS_SCOPE: &str = "https://www.googleapis.com/auth/spreadsheets";

#[derive(Debug, Error, PartialEq)]
pub enum SheetsError {
    #[error("GOOGLE_SHEETS_LEAD_SHEET_ID not set")]
    NotConfigured,
    #[error("failed to base64-decode GOOGLE_SHEETS_CREDENTIALS")]
    DecodeCredentials,
    #[error("failed to parse Google service account credentials")]
    ParseCredentials,
    #[error("transport: {0}")]
    Transport(String),
}

#[derive(Debug, Clone, serde::Serialize, serde::Deserialize)]
pub struct ServiceAccount {
    #[serde(default, rename = "client_email")]
    pub client_email: String,
    #[serde(default, rename = "private_key")]
    pub private_key: String,
    #[serde(default, rename = "token_uri")]
    pub token_uri: String,
}

pub fn decode_credentials_key(encoded: &str) -> Result<ServiceAccount, SheetsError> {
    let json_bytes = base64::engine::general_purpose::STANDARD
        .decode(encoded)
        .map_err(|_| SheetsError::DecodeCredentials)?;
    serde_json::from_slice(&json_bytes).map_err(|_| SheetsError::ParseCredentials)
}

pub fn service_account_jwt(account: &ServiceAccount) -> Result<String, SheetsError> {
    let now = chrono::Utc::now().timestamp();
    let claims = json!({
        "iss": account.client_email,
        "scope": SHEETS_SCOPE,
        "aud": account.token_uri,
        "exp": now + 3600,
        "iat": now,
    });
    let header = Header {
        alg: Algorithm::RS256,
        kid: None,
        typ: Some("JWT".to_string()),
        ..Default::default()
    };
    let key = EncodingKey::from_rsa_pem(account.private_key.as_bytes())
        .map_err(|_| SheetsError::ParseCredentials)?;
    encode(&header, &claims, &key).map_err(|_| SheetsError::ParseCredentials)
}

pub fn append_url(spreadsheet_id: &str) -> String {
    format!(
        "https://sheets.googleapis.com/v4/spreadsheets/{spreadsheet_id}/values/{range}:append",
        range = LEADS_SHEET_RANGE
    )
}

pub struct SheetsClient {
    access_token: String,
}

impl SheetsClient {
    pub fn new(access_token: &str) -> Self {
        Self {
            access_token: access_token.to_string(),
        }
    }

    /// Best-effort append; returns false instead of erroring, mirroring Go.
    pub async fn append_lead_row(
        &self,
        spreadsheet_id: &str,
        values: &[serde_json::Value],
    ) -> bool {
        let client = reqwest::Client::new();
        let body = json!({
            "values": [values],
        });
        let response = client
            .post(append_url(spreadsheet_id))
            .bearer_auth(&self.access_token)
            .query(&[
                ("valueInputOption", "USER_ENTERED"),
                ("insertDataOption", "INSERT_ROWS"),
            ])
            .json(&body)
            .send()
            .await;
        matches!(response, Ok(response) if response.status().is_success())
    }

    pub async fn access_token_from_credentials(encoded: &str) -> Result<String, SheetsError> {
        let account = decode_credentials_key(encoded)?;
        let jwt = service_account_jwt(&account)?;
        let client = reqwest::Client::new();
        let params = [
            ("grant_type", "urn:ietf:params:oauth:grant-type:jwt-bearer"),
            ("assertion", jwt.as_str()),
        ];
        let response = client
            .post(&account.token_uri)
            .form(&params)
            .send()
            .await
            .map_err(|error| SheetsError::Transport(error.to_string()))?;
        let body: serde_json::Value = response
            .json()
            .await
            .map_err(|error| SheetsError::Transport(error.to_string()))?;
        body.get("access_token")
            .and_then(|token| token.as_str())
            .map(str::to_string)
            .ok_or(SheetsError::Transport(
                "no access_token in oauth response".to_string(),
            ))
    }
}

#[cfg(test)]
mod tests {
    use super::{append_url, decode_credentials_key, SheetsError};

    #[test]
    fn credentials_and_urls() {
        use base64::Engine as _;
        let json = r#"{"client_email":"svc@example.com","private_key":"k","token_uri":"https://oauth2.googleapis.com/token"}"#;
        let encoded = base64::engine::general_purpose::STANDARD.encode(json);
        let account = decode_credentials_key(&encoded).unwrap();
        assert_eq!(account.client_email, "svc@example.com");
        assert_eq!(
            decode_credentials_key("!!!not-base64!!!").unwrap_err(),
            SheetsError::DecodeCredentials
        );
        assert!(append_url("sheet-1").contains("Leads!A:J"));
    }
}
