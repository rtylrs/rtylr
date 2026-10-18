package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"
)

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

	rawRedisHosts := strings.TrimSpace(os.Getenv(RedisHostsEnvVar))
	redisHosts := normalizeRedisHosts(rawRedisHosts)
	if rawRedisHosts != "" && len(redisHosts) == 0 {
		return nil, fmt.Errorf("REDIS_HOSTS contains no usable host")
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
	googleRedirectURI := strings.TrimSpace(os.Getenv("GOOGLE_OAUTH_REDIRECT_URI"))
	if googleRedirectURI == "" {
		authURL := strings.TrimRight(envOrDefault("AUTH_URL", AuthAppURL), "/")
		googleRedirectURI = authURL + "/callback"
	}

	cfg := &AppConfig{
		RunMode:                 envOrDefault("RUN_MODE", "api"),
		DatabaseURL:             os.Getenv("DATABASE_URL"),
		DBHost:                  os.Getenv("DB_HOST"),
		DBPort:                  os.Getenv("DB_PORT"),
		DBUser:                  os.Getenv("DB_USER"),
		DBPass:                  os.Getenv("DB_PASSWORD"),
		DBName:                  os.Getenv("DB_NAME"),
		JWTSecret:               os.Getenv("JWT_SECRET"),
		JWTDuration:             jwtDuration,
		ResendKey:               os.Getenv("RESEND_API_KEY"),
		LeadNotifyEmails:        ParseEmailList(os.Getenv("LEAD_NOTIFY_EMAILS"), DefaultLeadNotifyEmails),
		NewsletterNotifyEmails:  ParseEmailList(os.Getenv("NEWSLETTER_NOTIFY_EMAILS"), DefaultNewsletterNotifyEmails),
		AppEnv:                  os.Getenv("APP_ENV"),
		AppRelease:              os.Getenv("APP_RELEASE"),
		Port:                    os.Getenv("PORT"),
		Hostname:                os.Getenv("HOSTNAME"),
		AuthAppURL:              envFirst("AUTH_URL", "AUTH_APP_URL"),
		GoogleOAuthClientID:     os.Getenv("GOOGLE_OAUTH_CLIENT_ID"),
		GoogleOAuthClientSecret: os.Getenv("GOOGLE_OAUTH_CLIENT_SECRET"),
		GoogleOAuthRedirectURI:  googleRedirectURI,

		DBMaxOpenConns:    dbMaxOpenConns,
		DBMaxIdleConns:    dbMaxIdleConns,
		DBConnMaxLifetime: dbConnMaxLifetime,
		DBConnMaxIdleTime: dbConnMaxIdleTime,

		// OpenAI, Gemini and OpenCode Zen
		OpenAIAPIKey:   os.Getenv("OPENAI_API_KEY"),
		OpenAIOrgID:    os.Getenv("OPENAI_ORG_ID"),
		OpenAIBaseURL:  os.Getenv("OPENAI_BASE_URL"),
		GeminiAPIKey:   os.Getenv("GEMINI_API_KEY"),
		OpenCodeAPIKey: strings.TrimSpace(os.Getenv(OpenCodeAPIKeyEnvVar)),

		// AI provider preferences
		DeepSeekAPIKey:    os.Getenv(DeepSeekAPIKeyEnvVar),
		AIPrimaryProvider: os.Getenv(AIPrimaryProviderEnvVar),
		SearXNGURL:        strings.TrimSpace(os.Getenv(SearXNGURLEnvVar)),
		BraveSearchAPIKey: strings.TrimSpace(os.Getenv(BraveSearchAPIKeyEnvVar)),

		// Caret document-to-form mapping specialist (optional)
		FormMappingCheckpointPath: strings.TrimSpace(os.Getenv("FORM_MAPPING_CHECKPOINT_PATH")),

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

		// CDN public media base. UPLOAD_URL remains the deployment override.
		CDNBaseURL:         envFirst("UPLOAD_URL", "CDN_BASE_URL"),
		PublicAPIURL:       strings.TrimRight(envOrDefault("PUBLIC_API_URL", DefaultAPIURL), "/"),
		CDNMaxUploadBytes:  int64(positiveEnvIntOrDefault("CDN_MAX_UPLOAD_MB", 100)) * 1024 * 1024,
		CDNMediaEnginePath: envOrDefault("MEDIA_ENGINE_PATH", "./media-engine/build/media-engine"),
		CDNAPIKey:          os.Getenv("CDN_API_KEY"),

		CDNModerationEnabled:    !strings.EqualFold(strings.TrimSpace(os.Getenv("CDN_MODERATION_ENABLED")), "false"),
		CDNModerationProvider:   strings.ToLower(envOrDefault("CDN_MODERATION_PROVIDER", "openai")),
		CDNModerationFailClosed: strings.EqualFold(strings.TrimSpace(os.Getenv("CDN_MODERATION_FAIL_CLOSED")), "true"),

		// Internal Auth
		InternalToken: os.Getenv("INTERNAL_TOKEN"),
		// Internal tooling (Partials) access. Trimmed so a stray newline from a
		// shell export or secret value can't silently break the header compare.
		PartialsAPIKey: strings.TrimSpace(os.Getenv("PARTIALS_API_KEY")),
		PartialsActorUserID: func() int64 {
			if raw := strings.TrimSpace(os.Getenv("PARTIALS_ACTOR_USER_ID")); raw != "" {
				if parsed, err := strconv.ParseInt(raw, 10, 64); err == nil {
					return parsed
				}
			}
			return 0
		}(),

		// Metrics Auth
		MetricsUser: os.Getenv("METRICS_USER"),
		MetricsPass: os.Getenv("METRICS_PASS"),
	}

	if cfg.AuthAppURL == "" {
		cfg.AuthAppURL = AuthAppURL
	}

	if cfg.CDNBaseURL == "" {
		cfg.CDNBaseURL = "https://api.voxire.com"
	}

	var missing []string

	// Every Onion service uses the business database and signing identity.
	if cfg.DatabaseURL == "" && (cfg.DBHost == "" || cfg.DBName == "" || cfg.DBUser == "") {
		missing = append(missing, "DATABASE_URL (or DB_HOST, DB_NAME, DB_USER)")
	}
	if cfg.JWTSecret == "" {
		missing = append(missing, "JWT_SECRET")
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("missing critical environment variable(s): %s", strings.Join(missing, ", "))
	}

	GlobalConfig = cfg
	assertAppEnvIsUnderstood(cfg.AppEnv)
	return cfg, nil
}

// assertAppEnvIsUnderstood reports the resolved environment at boot.
//
// It warns rather than returning an error on purpose. APP_ENV now defaults to
// production, so an unrecognised value is the safe direction and refusing to
// start would turn a security fix into an outage for any deployment whose
// label this build has not been taught. A loud, greppable line at boot is what
// makes the default visible; silence is what let this class of bug survive.
func assertAppEnvIsUnderstood(raw string) {
	env := NormalizeAppEnv(raw)

	switch {
	case env == "":
		slog.Warn("APP_ENV is not set; defaulting to production. " +
			"Development-only endpoints, CORS origins and unauthenticated webhook handling are disabled. " +
			"Set APP_ENV=development for a local machine.")
	case isUnrecognizedEnv(env):
		slog.Warn("APP_ENV is not a recognised environment; defaulting to production. "+
			"Development-only endpoints, CORS origins and unauthenticated webhook handling are disabled. "+
			"Known development values: dev, development, local, test.",
			slog.String("app_env", raw))
	case isNonProductionEnv(env):
		slog.Info("APP_ENV resolved to a development environment; "+
			"development-only endpoints and localhost CORS origins are enabled",
			slog.String("app_env", env))
	default:
		slog.Info("APP_ENV resolved to production", slog.String("app_env", env))
	}
}
