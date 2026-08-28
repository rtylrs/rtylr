package config

import (
	"os"
	"strings"
)

const (
	// #nosec G101 -- These are environment variable names, not secret values.
	PaddleAPIKeyEnvVar = "PADDLE_API_KEY"
	// #nosec G101 -- These are environment variable names, not secret values.
	PaddleWebhookSecretEnvVar = "PADDLE_WEBHOOK_SECRET"
	// #nosec G101 -- These are environment variable names, not secret values.
	PaddleClientTokenEnvVar = "PADDLE_CLIENT_TOKEN"
	PaddleSandboxEnvVar     = "PADDLE_SANDBOX"
)

func GetPaddleAPIKey() string {
	if GlobalConfig != nil {
		return GlobalConfig.PaddleAPIKey
	}
	return os.Getenv(PaddleAPIKeyEnvVar)
}

func GetPaddleWebhookSecret() string {
	if GlobalConfig != nil {
		return GlobalConfig.PaddleWebhookSecret
	}
	return os.Getenv(PaddleWebhookSecretEnvVar)
}

func GetPaddleClientToken() string {
	if GlobalConfig != nil {
		return GlobalConfig.PaddleClientToken
	}
	return os.Getenv(PaddleClientTokenEnvVar)
}

func IsPaddleSandbox() bool {
	if GlobalConfig != nil {
		return GlobalConfig.PaddleSandbox
	}
	return os.Getenv(PaddleSandboxEnvVar) == "true"
}

func IsPaddleConfigured() bool {
	return GetPaddleAPIKey() != ""
}

func IsProduction() bool {
	env := strings.ToLower(strings.TrimSpace(os.Getenv("APP_ENV")))
	if GlobalConfig != nil {
		env = strings.ToLower(strings.TrimSpace(GlobalConfig.AppEnv))
	}
	return env == "production" || env == "prod"
}

func GetPaddleBaseURL() string {
	if IsPaddleSandbox() {
		return "https://sandbox-api.paddle.com"
	}

	return "https://api.paddle.com"
}
