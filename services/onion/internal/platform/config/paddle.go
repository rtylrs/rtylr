package config

import "os"

const (
	PaddleAPIKeyEnvVar        = "PADDLE_API_KEY"
	PaddleWebhookSecretEnvVar = "PADDLE_WEBHOOK_SECRET"
	PaddleClientTokenEnvVar   = "PADDLE_CLIENT_TOKEN"
	PaddleSandboxEnvVar       = "PADDLE_SANDBOX"
)

func GetPaddleAPIKey() string {
	return os.Getenv(PaddleAPIKeyEnvVar)
}

func GetPaddleWebhookSecret() string {
	return os.Getenv(PaddleWebhookSecretEnvVar)
}

func GetPaddleClientToken() string {
	return os.Getenv(PaddleClientTokenEnvVar)
}

func IsPaddleSandbox() bool {
	return os.Getenv(PaddleSandboxEnvVar) == "true"
}

func IsPaddleConfigured() bool {
	return GetPaddleAPIKey() != ""
}

func GetPaddleBaseURL() string {
	if IsPaddleSandbox() {
		return "https://sandbox-api.paddle.com"
	}

	return "https://api.paddle.com"
}
