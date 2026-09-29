package config

import (
	"os"
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
	return !IsNonProduction()
}

func GetPaddleBaseURL() string {
	if IsPaddleSandbox() {
		return "https://sandbox-api.paddle.com"
	}

	return "https://api.paddle.com"
}
