//! Chat completions with the Go fallback order, mirroring the snips
//! OpenAI/Gemini clients and `config` provider selection:
//! OpenCode → Gemini → DeepSeek → OpenAI.
//! `AI_PRIMARY_PROVIDER` forces one provider when configured.

use serde::{Deserialize, Serialize};
use thiserror::Error;

pub const OPENCODE_BASE_URL: &str = "https://opencode.ai/zen/go/v1";
pub const OPENCODE_MODEL: &str = "deepseek-v4.1-flash";
pub const DEEPSEEK_BASE_URL: &str = "https://api.deepseek.com";
pub const DEEPSEEK_MODEL: &str = "deepseek-chat";
pub const OPENAI_BASE_URL: &str = "https://api.openai.com/v1";
pub const OPENAI_MODEL: &str = "gpt-4o-mini";
pub const GEMINI_BASE_URL: &str = "https://generativelanguage.googleapis.com/v1beta";
pub const GEMINI_MODEL: &str = "gemini-3.8-flash";

/// Gemini native API models, tried in order (Go `gemini.go`).
pub const GEMINI_MODELS: [&str; 2] = ["gemini-3.8-flash", "gemini-3.7-flash"];
/// Gemini request timeouts (Go `geminiRequestTimeout` / `geminiModelTimeout`).
pub const GEMINI_REQUEST_TIMEOUT_SECS: u64 = 35;

pub const PROVIDER_ORDER: [&str; 4] = ["opencode", "gemini", "deepseek", "openai"];

pub const AI_REQUEST_TIMEOUT_SECS: u64 = 90;
pub const AI_MAX_RETRIES: i32 = 2;

#[derive(Debug, Error, PartialEq)]
pub enum AiError {
    #[error("no AI provider configured")]
    NoProvider,
    #[error("provider {0} failed after {1} attempt(s): {2}")]
    ProviderFailed(String, i32, String),
    #[error("transport: {0}")]
    Transport(String),
    #[error("rate limited, retry after {0}ms")]
    RateLimited(u64),
    #[error("response too large")]
    TooLarge,
}

#[derive(Debug, Clone, Serialize)]
pub struct ChatMessage {
    pub role: String,
    pub content: String,
}

#[derive(Debug, Clone, Serialize)]
pub struct ChatRequest {
    pub model: String,
    pub messages: Vec<ChatMessage>,
    #[serde(default, skip_serializing_if = "is_zero_i32")]
    pub max_tokens: i32,
}

fn is_zero_i32(value: &i32) -> bool {
    *value == 0
}

#[derive(Debug, Clone, Deserialize)]
pub struct ChatResponse {
    #[serde(default)]
    pub choices: Vec<ChatChoice>,
}

#[derive(Debug, Clone, Deserialize)]
pub struct ChatChoice {
    #[serde(default)]
    pub message: ChatChoiceMessage,
}

#[derive(Debug, Clone, Default, Deserialize)]
pub struct ChatChoiceMessage {
    #[serde(default)]
    pub content: String,
}

#[derive(Debug, Clone, Deserialize)]
pub struct ModerationResponse {
    #[serde(default)]
    pub results: Vec<ModerationResult>,
}

#[derive(Debug, Clone, Deserialize)]
pub struct ModerationResult {
    #[serde(default)]
    pub flagged: bool,
}

#[derive(Debug, Clone)]
pub struct Provider {
    pub name: String,
    pub base_url: String,
    pub api_key: String,
    pub model: String,
}

impl Provider {
    pub fn is_configured(&self) -> bool {
        !self.api_key.is_empty()
    }
}

/// Fallback order as the clients walk it (Go default: opencode, gemini,
/// deepseek, openai). `AI_PRIMARY_PROVIDER` moves the named provider to
/// the front.
pub fn provider_order(primary: &str) -> Vec<&'static str> {
    let primary = primary.trim().to_ascii_lowercase();
    if primary.is_empty() || !PROVIDER_ORDER.contains(&primary.as_str()) {
        return PROVIDER_ORDER.to_vec();
    }
    let mut ordered = PROVIDER_ORDER.to_vec();
    let first = PROVIDER_ORDER
        .iter()
        .copied()
        .find(|name| *name == primary.as_str())
        .unwrap_or("opencode");
    ordered.retain(|name| *name != primary.as_str());
    ordered.insert(0, first);
    ordered
}

/// The request body for a chat completion (OpenAI-compatible shape; Gemini
/// is converted to the same by the caller when needed).
pub fn chat_request(model: &str, messages: &[ChatMessage], max_tokens: i32) -> ChatRequest {
    ChatRequest {
        model: model.to_string(),
        messages: messages.to_vec(),
        max_tokens,
    }
}

/// Try providers in order until one succeeds, retrying transient
/// failures within each provider.
pub async fn complete_with_fallback(
    providers: &[Provider],
    messages: &[ChatMessage],
    max_tokens: i32,
) -> Result<String, AiError> {
    let mut last_error = String::new();
    for provider in providers {
        if !provider.is_configured() {
            continue;
        }
        let mut attempt = 0;
        while attempt < AI_MAX_RETRIES + 1 {
            attempt += 1;
            if attempt > 1 {
                tokio::time::sleep(std::time::Duration::from_millis(
                    200 * 2u64.pow(attempt as u32 - 2),
                ))
                .await;
            }
            match complete(provider, messages, max_tokens).await {
                Ok(text) => return Ok(text),
                Err(AiError::RateLimited(_)) => {
                    tokio::time::sleep(std::time::Duration::from_secs(1)).await;
                    last_error = "rate limited".to_string();
                }
                Err(AiError::Transport(error)) => last_error = error,
                Err(_) => break,
            }
        }
    }
    if last_error.is_empty() {
        Err(AiError::NoProvider)
    } else {
        Err(AiError::ProviderFailed("all".to_string(), 0, last_error))
    }
}

pub async fn complete(
    provider: &Provider,
    messages: &[ChatMessage],
    max_tokens: i32,
) -> Result<String, AiError> {
    if provider.name == "gemini" {
        return complete_gemini(&provider.api_key, messages, max_tokens).await;
    }
    let client = reqwest::Client::builder()
        .timeout(std::time::Duration::from_secs(AI_REQUEST_TIMEOUT_SECS))
        .build()
        .map_err(|error| AiError::Transport(error.to_string()))?;
    let request = chat_request(&provider.model, messages, max_tokens);
    let response = client
        .post(format!(
            "{}/chat/completions",
            provider.base_url.trim_end_matches('/')
        ))
        .header("Authorization", format!("Bearer {}", provider.api_key))
        .json(&request)
        .send()
        .await
        .map_err(|error| AiError::Transport(error.to_string()))?;
    let status = response.status();
    let body = response
        .bytes()
        .await
        .map_err(|error| AiError::Transport(error.to_string()))?;
    if body.len() > (1 << 20) {
        return Err(AiError::TooLarge);
    }
    if status.as_u16() == 429 {
        return Err(AiError::RateLimited(1000));
    }
    if !status.is_success() {
        return Err(AiError::Transport(format!(
            "{status}: {}",
            String::from_utf8_lossy(&body)
        )));
    }
    let parsed: ChatResponse =
        serde_json::from_slice(&body).map_err(|error| AiError::Transport(error.to_string()))?;
    Ok(parsed
        .choices
        .first()
        .map(|choice| choice.message.content.clone())
        .unwrap_or_default())
}

#[derive(Debug, Clone, Deserialize)]
struct GeminiContentPart {
    #[serde(default)]
    text: Option<String>,
}

#[derive(Debug, Clone, Deserialize)]
struct GeminiContent {
    #[serde(default)]
    parts: Vec<GeminiContentPart>,
}

#[derive(Debug, Clone, Deserialize)]
struct GeminiCandidate {
    #[serde(default)]
    content: Option<GeminiContent>,
}

#[derive(Debug, Clone, Deserialize)]
struct GeminiResponse {
    #[serde(default)]
    candidates: Vec<GeminiCandidate>,
}

/// `POST {base}/models/{model}:generateContent` with `x-goog-api-key`.
/// Tries `GEMINI_MODELS` in order (Go `gemini.go`).
async fn complete_gemini(
    api_key: &str,
    messages: &[ChatMessage],
    max_tokens: i32,
) -> Result<String, AiError> {
    let client = reqwest::Client::builder()
        .timeout(std::time::Duration::from_secs(GEMINI_REQUEST_TIMEOUT_SECS))
        .build()
        .map_err(|error| AiError::Transport(error.to_string()))?;
    let mut last_error = String::new();
    for model in GEMINI_MODELS {
        let contents: Vec<serde_json::Value> = messages
            .iter()
            .map(|message| {
                serde_json::json!({
                    "role": if message.role == "assistant" { "model" } else { "user" },
                    "parts": [{ "text": message.content }],
                })
            })
            .collect();
        let mut body = serde_json::json!({ "contents": contents });
        if max_tokens > 0 {
            body["generationConfig"] = serde_json::json!({ "maxOutputTokens": max_tokens });
        }
        let response = client
            .post(format!("{GEMINI_BASE_URL}/models/{model}:generateContent"))
            .header("x-goog-api-key", api_key)
            .json(&body)
            .send()
            .await
            .map_err(|error| AiError::Transport(error.to_string()));
        let response = match response {
            Ok(response) => response,
            Err(error) => {
                last_error = error.to_string();
                continue;
            }
        };
        let status = response.status();
        let bytes = response
            .bytes()
            .await
            .map_err(|error| AiError::Transport(error.to_string()))?;
        if bytes.len() > (1 << 20) {
            return Err(AiError::TooLarge);
        }
        if status.as_u16() == 429 {
            return Err(AiError::RateLimited(1000));
        }
        if !status.is_success() {
            last_error = format!("{status}: {}", String::from_utf8_lossy(&bytes));
            continue;
        }
        let parsed: GeminiResponse = serde_json::from_slice(&bytes)
            .map_err(|error| AiError::Transport(error.to_string()))?;
        let text = parsed
            .candidates
            .first()
            .and_then(|candidate| candidate.content.as_ref())
            .map(|content| {
                content
                    .parts
                    .iter()
                    .filter_map(|part| part.text.clone())
                    .collect::<Vec<String>>()
                    .join("")
            })
            .unwrap_or_default();
        if !text.is_empty() {
            return Ok(text);
        }
        last_error = "empty gemini response".to_string();
    }
    Err(AiError::Transport(last_error))
}

pub async fn moderate(provider: &Provider, texts: &[String]) -> Result<Vec<bool>, AiError> {
    let client = reqwest::Client::builder()
        .timeout(std::time::Duration::from_secs(AI_REQUEST_TIMEOUT_SECS))
        .build()
        .map_err(|error| AiError::Transport(error.to_string()))?;
    let response = client
        .post(format!(
            "{}/moderations",
            provider.base_url.trim_end_matches('/')
        ))
        .header("Authorization", format!("Bearer {}", provider.api_key))
        .json(&serde_json::json!({"model": provider.model, "input": texts}))
        .send()
        .await
        .map_err(|error| AiError::Transport(error.to_string()))?;
    let body = response
        .bytes()
        .await
        .map_err(|error| AiError::Transport(error.to_string()))?;
    if body.len() > (1 << 20) {
        return Err(AiError::TooLarge);
    }
    let parsed: ModerationResponse =
        serde_json::from_slice(&body).map_err(|error| AiError::Transport(error.to_string()))?;
    Ok(parsed.results.iter().map(|result| result.flagged).collect())
}
#[cfg(test)]
mod tests {
    use super::{chat_request, provider_order, ChatMessage, Provider};

    #[test]
    fn orders_and_requests() {
        assert_eq!(
            provider_order(""),
            vec!["opencode", "gemini", "deepseek", "openai"]
        );
        assert_eq!(
            provider_order("deepseek"),
            vec!["deepseek", "opencode", "gemini", "openai"]
        );
        let request = chat_request(
            "gemini-2.0-flash",
            &[ChatMessage {
                role: "user".to_string(),
                content: "hi".to_string(),
            }],
            64,
        );
        assert_eq!(request.model, "gemini-2.0-flash");
        assert_eq!(request.messages.len(), 1);
        let provider = Provider {
            name: "x".to_string(),
            base_url: String::new(),
            api_key: String::new(),
            model: String::new(),
        };
        assert!(!provider.is_configured());
    }
}
