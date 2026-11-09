//! Trino analytics query, mirroring the `TRINO_*` configuration and the
//! read-only AI query path (`internal/platform/config/trino.go`).

use serde::Deserialize;
use thiserror::Error;

#[derive(Debug, Error, PartialEq)]
pub enum TrinoError {
    #[error("trino is not configured")]
    NotConfigured,
    #[error("trino api returned {0}: {1}")]
    ApiError(String, String),
    #[error("transport: {0}")]
    Transport(String),
}

#[derive(Debug, Clone, Deserialize)]
pub struct TrinoStatement {
    #[serde(default, rename = "columns")]
    pub columns: Vec<TrinoColumn>,
    #[serde(default, rename = "data")]
    pub data: Vec<Vec<serde_json::Value>>,
    #[serde(default, rename = "nextUri")]
    pub next_uri: Option<String>,
    #[serde(default)]
    pub error: Option<TrinoErrorBody>,
}

#[derive(Debug, Clone, Deserialize)]
pub struct TrinoColumn {
    #[serde(default, rename = "name")]
    pub name: String,
    #[serde(default, rename = "type")]
    pub column_type: String,
}

#[derive(Debug, Clone, Deserialize)]
pub struct TrinoErrorBody {
    #[serde(default)]
    pub message: String,
    #[serde(default, rename = "errorType")]
    pub error_type: String,
}

pub fn is_configured(enabled: bool, host: &str, user: &str) -> bool {
    enabled && !host.is_empty() && !user.is_empty()
}

/// Bound on `nextUri` polls in `run_to_completion` so a hostile or stuck
/// Trino response cannot loop forever. (Go `trino/client.go` bounds the
/// query with `defaultQueryTimeout`; the Rust loop additionally caps
/// polls and each request has `REQUEST_TIMEOUT_SECS`.)
pub const MAX_POLLS: usize = 10_000;
pub const REQUEST_TIMEOUT_SECS: u64 = 120;

pub struct TrinoClient {
    base_url: String,
    user: String,
    password: String,
    access_token: String,
    catalog: String,
    schema: String,
    http: reqwest::Client,
}

#[derive(Debug, Clone)]
pub struct TrinoParams<'a> {
    pub host: &'a str,
    pub port: &'a str,
    pub user: &'a str,
    pub password: &'a str,
    pub access_token: &'a str,
    pub catalog: &'a str,
    pub schema: &'a str,
    pub use_tls: bool,
}

impl TrinoClient {
    pub fn new(params: TrinoParams<'_>) -> Self {
        let scheme = if params.use_tls { "https" } else { "http" };
        Self {
            base_url: format!("{scheme}://{}:{}", params.host, params.port),
            user: params.user.to_string(),
            password: params.password.to_string(),
            access_token: params.access_token.to_string(),
            catalog: params.catalog.to_string(),
            schema: params.schema.to_string(),
            http: reqwest::Client::builder()
                .timeout(std::time::Duration::from_secs(REQUEST_TIMEOUT_SECS))
                .build()
                .unwrap_or_else(|_| reqwest::Client::new()),
        }
    }

    /// Submit one statement (Trino REST `/v1/statement`).
    pub async fn submit(&self, sql: &str) -> Result<TrinoStatement, TrinoError> {
        let mut request = self
            .http
            .post(format!("{}/v1/statement", self.base_url))
            .body(sql.to_string());
        if !self.user.is_empty() {
            request = request.basic_auth(&self.user, Some(&self.password));
        }
        if !self.access_token.is_empty() {
            request = request.bearer_auth(&self.access_token);
        }
        if !self.catalog.is_empty() {
            request = request.header("X-Trino-Catalog", &self.catalog);
        }
        if !self.schema.is_empty() {
            request = request.header("X-Trino-Schema", &self.schema);
        }
        let response = request
            .send()
            .await
            .map_err(|error| TrinoError::Transport(error.to_string()))?;
        let status = response.status();
        let body = response
            .bytes()
            .await
            .map_err(|error| TrinoError::Transport(error.to_string()))?;
        let parsed: TrinoStatement = serde_json::from_slice(&body)
            .map_err(|error| TrinoError::Transport(error.to_string()))?;
        if let Some(error) = parsed.error {
            return Err(TrinoError::ApiError(status.to_string(), error.message));
        }
        Ok(parsed)
    }

    /// Follow Trino's `nextUri` pagination to completion, collecting rows.
    pub async fn run_to_completion(
        &self,
        first: TrinoStatement,
    ) -> Result<Vec<Vec<serde_json::Value>>, TrinoError> {
        let mut rows = Vec::new();
        let mut current = first;
        let mut polls = 0;
        loop {
            rows.extend(current.data.clone());
            let Some(next) = current.next_uri.clone() else {
                return Ok(rows);
            };
            polls += 1;
            if polls > MAX_POLLS {
                return Err(TrinoError::ApiError(
                    "poll limit".to_string(),
                    format!("exceeded {MAX_POLLS} nextUri polls"),
                ));
            }
            let mut request = self.http.get(&next);
            if !self.user.is_empty() {
                request = request.basic_auth(&self.user, Some(&self.password));
            }
            if !self.access_token.is_empty() {
                let token = self.access_token.clone();
                request = request.bearer_auth(token);
            }
            let response = request
                .send()
                .await
                .map_err(|error| TrinoError::Transport(error.to_string()))?;
            let status = response.status();
            let body = response
                .bytes()
                .await
                .map_err(|error| TrinoError::Transport(error.to_string()))?;
            let parsed: TrinoStatement = serde_json::from_slice(&body)
                .map_err(|error| TrinoError::Transport(error.to_string()))?;
            if let Some(error) = parsed.error {
                return Err(TrinoError::ApiError(status.to_string(), error.message));
            }
            current = parsed;
        }
    }
}

#[cfg(test)]
mod tests {
    use super::is_configured;

    #[test]
    fn configured_gate() {
        assert!(is_configured(true, "localhost", "user"));
        assert!(!is_configured(false, "localhost", "user"));
        assert!(!is_configured(true, "", "user"));
        assert!(!is_configured(true, "localhost", ""));
    }
}
