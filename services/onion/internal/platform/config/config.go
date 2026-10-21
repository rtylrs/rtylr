package config

import (
	"time"
)

const (
	// #nosec G101 -- These are environment variable names, not secret values.
	OpenAIAPIKeyEnvVar = "OPENAI_API_KEY"
	// #nosec G101 -- This is an environment variable name, not a secret value.
	OpenAIOrgIDEnvVar = "OPENAI_ORG_ID"
	// #nosec G101 -- This is an environment variable name, not a secret value.
	OpenAIBaseURLEnvVar = "OPENAI_BASE_URL"
)

const (
	// #nosec G101 -- This is an environment variable name, not a secret value.
	GeminiAPIKeyEnvVar = "GEMINI_API_KEY"
)

const (
	// OpenCode Go is an optional final fallback. The endpoint and model stay
	// product-controlled, like the other OpenAI-compatible providers.
	OpenCodeAPIKeyEnvVar = "OPENCODE_API_KEY"
	OpenCodeBaseURL      = "https://opencode.ai/zen/go/v1"
	OpenCodeModel        = "deepseek-v4.1-flash"
)

const (
	// DefaultJWTDuration is used when JWT_DURATION is not configured. Access
	// tokens must always have a bounded lifetime.
	DefaultJWTDuration = 24 * time.Hour
)

const (
	// Direct DeepSeek is an optional fallback provider. Only the API key is an env secret
	// (set DEEPSEEK_API_KEY on AWS / in .env); the base URL and model are fixed
	// here on purpose. The product chooses the model internally, users never do.
	// #nosec G101 -- This is an environment variable name, not a secret value.
	DeepSeekAPIKeyEnvVar = "DEEPSEEK_API_KEY"
	// DeepSeekBaseURL is DeepSeek's OpenAI-compatible endpoint, so we drive it
	// through the OpenAI SDK with this base URL.
	DeepSeekBaseURL = "https://api.deepseek.com"
	// DeepSeekModel is the chat model we run. deepseek-chat tracks DeepSeek's
	// latest general chat model.
	DeepSeekModel = "deepseek-chat"
)

const (
	// AIPrimaryProviderEnvVar optionally forces which configured provider is
	// primary ("deepseek", "gemini", "openai", "opencode"). Empty uses default order.
	AIPrimaryProviderEnvVar = "AI_PRIMARY_PROVIDER"
	SearXNGURLEnvVar        = "SEARXNG_URL"
	BraveSearchAPIKeyEnvVar = "BRAVE_SEARCH_API_KEY"
)

const (
	// AIRequestTimeout bounds any single LLM call so a hung provider can't pin
	// an SSE stream open forever.
	AIRequestTimeout = 90 * time.Second
	// AIMaxOutputTokens caps generation length (cost + latency guardrail).
	AIMaxOutputTokens = 2048
	// AIMaxRetries is how many times a transient provider error is retried with
	// exponential backoff before falling through to the next provider.
	AIMaxRetries = 2
	// AITokensPerCredit converts a plan's "AI credits" allowance into a real
	// token budget. One credit = this many tokens of combined prompt+completion.
	AITokensPerCredit = 1000
)

type AppConfig struct {
	RunMode                 string
	DatabaseURL             string
	DBHost                  string
	DBPort                  string
	DBUser                  string
	DBPass                  string
	DBName                  string
	JWTSecret               string
	JWTDuration             time.Duration
	ResendKey               string
	LeadNotifyEmails        []string
	NewsletterNotifyEmails  []string
	ResendAudienceID        string
	ResendAudienceKey       string
	AppEnv                  string
	AppRelease              string
	Port                    string
	Hostname                string
	AuthAppURL              string
	GoogleOAuthClientID     string
	GoogleOAuthClientSecret string
	GoogleOAuthRedirectURI  string

	// DB Connection Pool configuration
	DBMaxOpenConns    int
	DBMaxIdleConns    int
	DBConnMaxLifetime time.Duration
	DBConnMaxIdleTime time.Duration

	// OpenAI, Gemini and OpenCode Go (fallback providers)
	OpenAIAPIKey   string
	OpenAIOrgID    string
	OpenAIBaseURL  string
	GeminiAPIKey   string
	OpenCodeAPIKey string

	// AI provider secrets and preferred provider.
	DeepSeekAPIKey    string
	AIPrimaryProvider string
	SearXNGURL        string
	BraveSearchAPIKey string

	// FormMappingCheckpointPath points the Caret document-to-form mapping tool at
	// a form-mapping classifier checkpoint (directory, .safetensors, or .json).
	// The hosted image bundles one and sets this; empty uses the lexical matcher.
	FormMappingCheckpointPath string

	// Redis & Cache
	RedisURL               string
	RedisHosts             []string
	RedisUsername          string
	RedisPassword          string
	RedisEnabled           bool
	InMemoryCacheMaxSizeMb int

	// Trino analytics warehouse (optional)
	TrinoEnabled     bool
	TrinoHost        string
	TrinoPort        string
	TrinoUser        string
	TrinoPassword    string
	TrinoAccessToken string
	TrinoCatalog     string
	TrinoSchema      string
	TrinoTLS         bool
	TrinoSource      string

	// Paddle
	PaddleAPIKey        string
	PaddleWebhookSecret string
	PaddleClientToken   string
	PaddleSandbox       bool

	// Database Logging
	LoggingDatabaseURL string

	// Better Stack logging
	BetterStackSourceToken   string
	BetterStackIngestingHost string

	// Slack logging — routes error/request/suspicious logs to Slack channels.
	SlackLoggingEnabled     bool
	SlackBotToken           string
	SlackAlertsChannel      string
	SlackRequestLogsChannel string
	SlackSuspiciousChannel  string
	SlackDashboardChannel   string

	// Sentry
	SentryDSN string

	// Google Sheets
	GoogleSheetsLeadSheetID string
	GoogleSheetsCredentials string

	// AWS S3 & Floci Storage
	S3BucketName string
	S3Region     string
	// S3Endpoint is a custom S3-compatible endpoint (MinIO, LocalStack, floci,
	// a customer's own object store). Empty = real AWS S3. S3ForcePathStyle
	// turns on path-style addressing, which most non-AWS S3 services require.
	S3Endpoint       string
	S3ForcePathStyle bool
	FlociEnabled     bool
	FlociEndpoint    string
	FlociRegion      string
	FlociBucket      string

	// CDN runmode (cmd/upload) — media/CDN service settings.
	CDNBaseURL string // public base for generated file URLs
	// PublicAPIURL is this API's public origin, for links that leave the
	// app, such as the unsubscribe link in a marketing email.
	PublicAPIURL       string
	CDNMaxUploadBytes  int64  // hard ceiling on a single upload
	CDNMediaEnginePath string // path to the C++ media-engine binary
	CDNAPIKey          string // bearer key for write endpoints; empty disables auth

	// CDN image moderation (nudity filter on stored uploads). Enabled by
	// default (set CDN_MODERATION_ENABLED=false to opt out); provider "openai"
	// uses the free moderation endpoint (needs OPENAI_API_KEY), "local" uses
	// an in-process skin-tone heuristic.
	CDNModerationEnabled    bool
	CDNModerationProvider   string // "openai" | "local"
	CDNModerationFailClosed bool   // reject uploads when the moderation check errors

	// Internal Auth
	InternalToken string

	// PartialsAPIKey gates the internal X-API-KEY bypass used by the Partials
	// tooling to reach internal endpoints without the standard middleware
	// chain. Empty (the default) disables the bypass entirely.
	PartialsAPIKey string
	// PartialsActorUserID is the user the X-API-KEY bypass acts as. The bypass
	// always assumes the console_admin role so internal tooling can reach the
	// console endpoints; this id is what audit and license writes record.
	PartialsActorUserID int64

	// Metrics Auth
	MetricsUser string
	MetricsPass string
}

var GlobalConfig *AppConfig
