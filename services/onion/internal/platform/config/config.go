package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
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
	// DefaultJWTDuration is used when JWT_DURATION is not configured. Access
	// tokens must always have a bounded lifetime.
	DefaultJWTDuration = 24 * time.Hour
	MinJWTSecretLength = 32
)

const (
	// DeepSeek is the primary LLM provider. Only the API key is an env secret
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
	// NVIDIA NIM is an OpenAI-compatible provider (hosted catalog or self-hosted).
	// Only the API key is an env secret (NVIDIA_NIM_API_KEY); base URL + model are
	// fixed here. Point NvidiaNIMBaseURL at a self-hosted NIM if/when we run one.
	// #nosec G101 -- This is an environment variable name, not a secret value.
	NvidiaNIMAPIKeyEnvVar = "NVIDIA_NIM_API_KEY"
	NvidiaNIMBaseURL      = "https://integrate.api.nvidia.com/v1"
	NvidiaNIMModel        = "meta/llama-3.1-8b-instruct"

	// AIPrimaryProviderEnvVar optionally forces which configured provider is
	// primary ("deepseek" | "nvidia"). Empty = use the default priority order.
	AIPrimaryProviderEnvVar = "AI_PRIMARY_PROVIDER"
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
	RunMode             string
	DatabaseURL         string
	DBHost              string
	DBPort              string
	DBUser              string
	DBPass              string
	DBName              string
	JWTSecret           string
	JWTDuration         time.Duration
	ResendKey           string
	AppEnv              string
	AppRelease          string
	DeploymentMode      string
	Port                string
	Hostname            string
	AuthAppURL          string
	GoogleOAuthClientID string

	// WorkOS
	WorkOSAPIKey      string
	WorkOSClientID    string
	WorkOSRedirectURI string

	// DB Connection Pool configuration
	DBMaxOpenConns    int
	DBMaxIdleConns    int
	DBConnMaxLifetime time.Duration
	DBConnMaxIdleTime time.Duration

	// OpenAI & Gemini
	OpenAIAPIKey  string
	OpenAIOrgID   string
	OpenAIBaseURL string
	GeminiAPIKey  string

	// DeepSeek + NVIDIA NIM (primary providers) — only the keys are configurable.
	DeepSeekAPIKey    string
	NvidiaNIMAPIKey   string
	AIPrimaryProvider string

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
	CDNBaseURL         string // public base for generated file URLs
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

	// License runmode / self-hosting controls.
	LicenseKey               string
	LicenseFile              string
	LicensePublicKey         string
	LicenseSigningPrivateKey string
	LicenseIssuer            string
	LicenseGracePeriod       time.Duration
	LicenseClockStateFile    string
	// LicenseAuthorityURL is the hosted license authority a self-host backend
	// validates RTYLR_LICENSE_KEY against. Defaults to the Voxire authority and
	// is intentionally NOT a documented customer knob (override only for our own
	// testing via LICENSE_AUTHORITY_URL).
	LicenseAuthorityURL string

	// Internal Auth
	InternalToken string

	// Metrics Auth
	MetricsUser string
	MetricsPass string
}

var GlobalConfig *AppConfig

// LoadAndValidate reads the environment configuration at startup,
// validating the presence of critical variables.
func LoadAndValidate() (*AppConfig, error) {
	jwtDuration := DefaultJWTDuration
	if rawDuration := os.Getenv("JWT_DURATION"); rawDuration != "" {
		parsed, err := time.ParseDuration(rawDuration)
		if err != nil || parsed <= 0 {
			if err != nil {
				return nil, fmt.Errorf("invalid JWT_DURATION %q: %w", rawDuration, err)
			}
			return nil, fmt.Errorf("JWT_DURATION must be greater than zero, got %q", rawDuration)
		}
		jwtDuration = parsed
	}

	hosts := os.Getenv("REDIS_HOSTS")
	var redisHosts []string
	if hosts != "" {
		redisHosts = strings.Split(hosts, ",")
	} else {
		redisHosts = []string{"localhost:6379"}
	}

	inMemoryCacheMaxSizeMb := 0
	if sizeStr := os.Getenv("INMEMORY_CACHE_MAX_SIZE_MB"); sizeStr != "" {
		if parsed, err := strconv.Atoi(sizeStr); err == nil {
			inMemoryCacheMaxSizeMb = parsed
		}
	}

	var dbMaxOpenConns int
	if maxOpenStr := os.Getenv("DB_MAX_OPEN_CONNS"); maxOpenStr != "" {
		if parsed, err := strconv.Atoi(maxOpenStr); err == nil && parsed > 0 {
			dbMaxOpenConns = parsed
		}
	}

	var dbMaxIdleConns int
	if maxIdleStr := os.Getenv("DB_MAX_IDLE_CONNS"); maxIdleStr != "" {
		if parsed, err := strconv.Atoi(maxIdleStr); err == nil && parsed > 0 {
			dbMaxIdleConns = parsed
		}
	}

	var dbConnMaxLifetime time.Duration
	if lifetimeStr := os.Getenv("DB_CONN_MAX_LIFETIME"); lifetimeStr != "" {
		if parsed, err := time.ParseDuration(lifetimeStr); err == nil && parsed > 0 {
			dbConnMaxLifetime = parsed
		}
	}

	var dbConnMaxIdleTime time.Duration
	if idleTimeStr := os.Getenv("DB_CONN_MAX_IDLE_TIME"); idleTimeStr != "" {
		if parsed, err := time.ParseDuration(idleTimeStr); err == nil && parsed > 0 {
			dbConnMaxIdleTime = parsed
		}
	}

	cfg := &AppConfig{
		RunMode:             envOrDefault("RUN_MODE", "api"),
		DatabaseURL:         os.Getenv("DATABASE_URL"),
		DBHost:              os.Getenv("DB_HOST"),
		DBPort:              os.Getenv("DB_PORT"),
		DBUser:              os.Getenv("DB_USER"),
		DBPass:              os.Getenv("DB_PASSWORD"),
		DBName:              os.Getenv("DB_NAME"),
		JWTSecret:           os.Getenv("JWT_SECRET"),
		JWTDuration:         jwtDuration,
		ResendKey:           os.Getenv("RESEND_API_KEY"),
		AppEnv:              os.Getenv("APP_ENV"),
		AppRelease:          os.Getenv("APP_RELEASE"),
		DeploymentMode:      envOrDefault("DEPLOYMENT_MODE", "hosted"),
		Port:                os.Getenv("PORT"),
		Hostname:            os.Getenv("HOSTNAME"),
		AuthAppURL:          envFirst("AUTH_URL", "AUTH_APP_URL"),
		GoogleOAuthClientID: os.Getenv("GOOGLE_OAUTH_CLIENT_ID"),

		// WorkOS
		WorkOSAPIKey:      os.Getenv("WORKOS_API_KEY"),
		WorkOSClientID:    os.Getenv("WORKOS_CLIENT_ID"),
		WorkOSRedirectURI: os.Getenv("WORKOS_REDIRECT_URI"),

		DBMaxOpenConns:    dbMaxOpenConns,
		DBMaxIdleConns:    dbMaxIdleConns,
		DBConnMaxLifetime: dbConnMaxLifetime,
		DBConnMaxIdleTime: dbConnMaxIdleTime,

		// OpenAI & Gemini
		OpenAIAPIKey:  os.Getenv("OPENAI_API_KEY"),
		OpenAIOrgID:   os.Getenv("OPENAI_ORG_ID"),
		OpenAIBaseURL: os.Getenv("OPENAI_BASE_URL"),
		GeminiAPIKey:  os.Getenv("GEMINI_API_KEY"),

		// DeepSeek + NVIDIA NIM
		DeepSeekAPIKey:    os.Getenv(DeepSeekAPIKeyEnvVar),
		NvidiaNIMAPIKey:   os.Getenv(NvidiaNIMAPIKeyEnvVar),
		AIPrimaryProvider: os.Getenv(AIPrimaryProviderEnvVar),

		// Redis & Cache
		RedisURL:               os.Getenv("REDIS_URL"),
		RedisHosts:             redisHosts,
		RedisUsername:          os.Getenv("REDIS_USERNAME"),
		RedisPassword:          os.Getenv("REDIS_PASSWORD"),
		RedisEnabled:           strings.ToLower(os.Getenv("REDIS_ENABLED")) == "true",
		InMemoryCacheMaxSizeMb: inMemoryCacheMaxSizeMb,

		// Trino analytics warehouse (optional)
		TrinoEnabled:     strings.ToLower(strings.TrimSpace(os.Getenv("TRINO_ENABLED"))) == "true",
		TrinoHost:        strings.TrimSpace(os.Getenv("TRINO_HOST")),
		TrinoPort:        strings.TrimSpace(os.Getenv("TRINO_PORT")),
		TrinoUser:        strings.TrimSpace(os.Getenv("TRINO_USER")),
		TrinoPassword:    os.Getenv("TRINO_PASSWORD"),
		TrinoAccessToken: os.Getenv("TRINO_ACCESS_TOKEN"),
		TrinoCatalog:     strings.TrimSpace(os.Getenv("TRINO_CATALOG")),
		TrinoSchema:      strings.TrimSpace(os.Getenv("TRINO_SCHEMA")),
		TrinoTLS:         strings.EqualFold(strings.TrimSpace(os.Getenv("TRINO_TLS")), "true"),
		TrinoSource:      strings.TrimSpace(os.Getenv("TRINO_SOURCE")),

		// Paddle
		PaddleAPIKey:        os.Getenv("PADDLE_API_KEY"),
		PaddleWebhookSecret: os.Getenv("PADDLE_WEBHOOK_SECRET"),
		PaddleClientToken:   os.Getenv("PADDLE_CLIENT_TOKEN"),
		PaddleSandbox:       os.Getenv("PADDLE_SANDBOX") == "true",

		// Database Logging
		LoggingDatabaseURL: os.Getenv("LOGGING_DATABASE_URL"),

		// Better Stack logging
		BetterStackSourceToken:   os.Getenv("BETTERSTACK_SOURCE_TOKEN"),
		BetterStackIngestingHost: strings.TrimSpace(os.Getenv("BETTERSTACK_INGESTING_HOST")),

		SlackLoggingEnabled:     strings.ToLower(strings.TrimSpace(os.Getenv("SLACK_LOGGING_ENABLED"))) == "true",
		SlackBotToken:           os.Getenv("SLACK_BOT_TOKEN"),
		SlackAlertsChannel:      os.Getenv("SLACK_ALERTS_CHANNEL_ID"),
		SlackRequestLogsChannel: os.Getenv("SLACK_REQUEST_LOGS_CHANNEL"),
		SlackSuspiciousChannel:  os.Getenv("SLACK_SUSPICIOUS_REQUESTS_CHANNEL"),
		SlackDashboardChannel:   os.Getenv("SLACK_DASHBOARD_CHANNEL_ID"),

		// Sentry
		SentryDSN: os.Getenv("SENTRY_DSN"),

		// Google Sheets
		GoogleSheetsLeadSheetID: os.Getenv("GOOGLE_SHEETS_LEAD_SHEET_ID"),
		GoogleSheetsCredentials: os.Getenv("GOOGLE_SHEETS_CREDENTIALS"),

		// AWS S3 & Floci Storage
		S3BucketName:     os.Getenv("S3_BUCKET_NAME"),
		S3Region:         os.Getenv("S3_REGION"),
		S3Endpoint:       os.Getenv("S3_ENDPOINT"),
		S3ForcePathStyle: strings.EqualFold(strings.TrimSpace(os.Getenv("S3_FORCE_PATH_STYLE")), "true"),
		FlociEnabled:     strings.ToLower(strings.TrimSpace(os.Getenv("FLOCI_ENABLED"))) == "true",
		FlociEndpoint:    os.Getenv("FLOCI_ENDPOINT"),
		FlociRegion:      os.Getenv("FLOCI_REGION"),
		FlociBucket:      os.Getenv("FLOCI_BUCKET"),

		// CDN runmode (cmd/upload). In self-host the public media base is the
		// upload service (UPLOAD_URL); hosted falls back to CDN_BASE_URL / the
		// cdn.voxire.com default (applied below).
		CDNBaseURL:         envFirst("UPLOAD_URL", "CDN_BASE_URL"),
		CDNMaxUploadBytes:  int64(positiveEnvIntOrDefault("CDN_MAX_UPLOAD_MB", 100)) * 1024 * 1024,
		CDNMediaEnginePath: envOrDefault("MEDIA_ENGINE_PATH", "./media-engine/build/media-engine"),
		CDNAPIKey:          os.Getenv("CDN_API_KEY"),

		CDNModerationEnabled:    !strings.EqualFold(strings.TrimSpace(os.Getenv("CDN_MODERATION_ENABLED")), "false"),
		CDNModerationProvider:   strings.ToLower(envOrDefault("CDN_MODERATION_PROVIDER", "openai")),
		CDNModerationFailClosed: strings.EqualFold(strings.TrimSpace(os.Getenv("CDN_MODERATION_FAIL_CLOSED")), "true"),

		// License runmode / self-hosting controls. Self-host customers set the
		// documented RTYLR_LICENSE_KEY; LICENSE_KEY stays as an internal alias.
		LicenseKey:  envFirst("RTYLR_LICENSE_KEY", "LICENSE_KEY"),
		LicenseFile: os.Getenv("LICENSE_FILE"),
		// LicensePublicKey is retained for console compatibility. Verification
		// uses the vendor key embedded in the released binary instead.
		LicensePublicKey:         "",
		LicenseSigningPrivateKey: os.Getenv("LICENSE_SIGNING_PRIVATE_KEY"),
		LicenseIssuer:            envOrDefault("LICENSE_ISSUER", "voxire"),
		LicenseGracePeriod:       envDurationOrDefault("LICENSE_GRACE_PERIOD", 7*24*time.Hour),
		LicenseClockStateFile:    envOrDefault("LICENSE_CLOCK_STATE_FILE", "/var/lib/rtylr/license-clock.json"),
		LicenseAuthorityURL:      DefaultLicenseAuthorityURL,

		// Internal Auth
		InternalToken: os.Getenv("INTERNAL_TOKEN"),

		// Metrics Auth
		MetricsUser: os.Getenv("METRICS_USER"),
		MetricsPass: os.Getenv("METRICS_PASS"),
	}

	if cfg.AuthAppURL == "" {
		cfg.AuthAppURL = AuthAppURL
	}

	if cfg.CDNBaseURL == "" {
		cfg.CDNBaseURL = "https://cdn.voxire.com"
	}

	var missing []string

	// Validate critical environment variables. The license runmode is the
	// only DB-less Onion service; every business-data runmode stays strict.
	if cfg.RunMode != "license" {
		if cfg.DatabaseURL == "" && (cfg.DBHost == "" || cfg.DBName == "" || cfg.DBUser == "") {
			missing = append(missing, "DATABASE_URL (or DB_HOST, DB_NAME, DB_USER)")
		}
		if cfg.JWTSecret == "" {
			missing = append(missing, "JWT_SECRET")
		}
	}
	if cfg.JWTSecret != "" && len(cfg.JWTSecret) < MinJWTSecretLength {
		return nil, fmt.Errorf("JWT_SECRET must be at least %d characters", MinJWTSecretLength)
	}

	if len(missing) > 0 {
		return nil, fmt.Errorf("missing critical environment variable(s): %s", strings.Join(missing, ", "))
	}

	GlobalConfig = cfg
	return cfg, nil
}

// envOrDefault returns the env var value, or fallback when it's unset/empty.
func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// envFirst returns the value of the first set/non-empty env var in keys, or "".
// Used to accept a customer-facing name and an internal alias (e.g. the
// documented RTYLR_LICENSE_KEY plus the legacy LICENSE_KEY).
func envFirst(keys ...string) string {
	for _, k := range keys {
		if v := strings.TrimSpace(os.Getenv(k)); v != "" {
			return v
		}
	}
	return ""
}

// envIntOrDefault parses an int env var, returning fallback when unset or
// unparseable.
func envIntOrDefault(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

func positiveEnvIntOrDefault(key string, fallback int) int {
	n := envIntOrDefault(key, fallback)
	if n <= 0 {
		return fallback
	}
	return n
}

func envDurationOrDefault(key string, fallback time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return fallback
	}
	return d
}

func GetOpenAIAPIKey() string {
	if GlobalConfig != nil {
		return GlobalConfig.OpenAIAPIKey
	}
	return os.Getenv(OpenAIAPIKeyEnvVar)
}

func GetOpenAIOrgID() string {
	if GlobalConfig != nil {
		return GlobalConfig.OpenAIOrgID
	}
	return os.Getenv(OpenAIOrgIDEnvVar)
}

func GetOpenAIBaseURL() string {
	var baseURL string
	if GlobalConfig != nil {
		baseURL = GlobalConfig.OpenAIBaseURL
	} else {
		baseURL = os.Getenv(OpenAIBaseURLEnvVar)
	}

	if baseURL == "" {
		return "https://api.openai.com/v1"
	}

	return baseURL
}

func IsOpenAIConfigured() bool {
	return GetOpenAIAPIKey() != ""
}

func GetGeminiAPIKey() string {
	if GlobalConfig != nil {
		return GlobalConfig.GeminiAPIKey
	}
	return os.Getenv(GeminiAPIKeyEnvVar)
}

func IsGeminiConfigured() bool {
	return GetGeminiAPIKey() != ""
}

func GetDeepSeekAPIKey() string {
	if GlobalConfig != nil {
		return GlobalConfig.DeepSeekAPIKey
	}
	return os.Getenv(DeepSeekAPIKeyEnvVar)
}

// GetDeepSeekBaseURL / GetDeepSeekModel return the hardcoded values — they are
// intentionally not env-configurable.
func GetDeepSeekBaseURL() string { return DeepSeekBaseURL }
func GetDeepSeekModel() string   { return DeepSeekModel }

func IsDeepSeekConfigured() bool {
	return GetDeepSeekAPIKey() != ""
}

func GetNvidiaNIMAPIKey() string {
	if GlobalConfig != nil {
		return GlobalConfig.NvidiaNIMAPIKey
	}
	return os.Getenv(NvidiaNIMAPIKeyEnvVar)
}

// GetNvidiaNIMBaseURL / GetNvidiaNIMModel return the hardcoded values.
func GetNvidiaNIMBaseURL() string { return NvidiaNIMBaseURL }
func GetNvidiaNIMModel() string   { return NvidiaNIMModel }

func IsNvidiaNIMConfigured() bool {
	return GetNvidiaNIMAPIKey() != ""
}

// GetAIPrimaryProvider returns the operator's preferred primary provider
// ("deepseek" | "nvidia"), or "" for the default priority order.
func GetAIPrimaryProvider() string {
	if GlobalConfig != nil {
		return GlobalConfig.AIPrimaryProvider
	}
	return os.Getenv(AIPrimaryProviderEnvVar)
}

func IsAnyAIConfigured() bool {
	return IsDeepSeekConfigured() || IsNvidiaNIMConfigured() || IsOpenAIConfigured() || IsGeminiConfigured()
}

// GetLicenseSigningPrivateKey returns the private key used to sign self-host
// license keys (license authority runmode).
func GetLicenseSigningPrivateKey() string {
	if GlobalConfig != nil {
		return GlobalConfig.LicenseSigningPrivateKey
	}
	return os.Getenv("LICENSE_SIGNING_PRIVATE_KEY")
}

// GetLicensePublicKey returns the public key used to verify license keys on
// self-host deployments.
func GetLicensePublicKey() string {
	if GlobalConfig != nil {
		return GlobalConfig.LicensePublicKey
	}
	return os.Getenv("LICENSE_PUBLIC_KEY")
}

// GetLicenseIssuer returns the issuer name embedded in generated license keys,
// defaulting to "voxire" when unset.
func GetLicenseIssuer() string {
	if GlobalConfig != nil && GlobalConfig.LicenseIssuer != "" {
		return GlobalConfig.LicenseIssuer
	}
	if v := strings.TrimSpace(os.Getenv("LICENSE_ISSUER")); v != "" {
		return v
	}
	return "voxire"
}
