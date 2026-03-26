package config

import (
	"os"
)

const (
	OpenAIAPIKeyEnvVar  = "OPENAI_API_KEY"
	OpenAIOrgIDEnvVar   = "OPENAI_ORG_ID"
	OpenAIBaseURLEnvVar = "OPENAI_BASE_URL"
)

const (
	GeminiAPIKeyEnvVar = "GEMINI_API_KEY"
)

func GetOpenAIAPIKey() string {
	return os.Getenv(OpenAIAPIKeyEnvVar)
}

func GetOpenAIOrgID() string {
	return os.Getenv(OpenAIOrgIDEnvVar)
}

func GetOpenAIBaseURL() string {
	baseURL := os.Getenv(OpenAIBaseURLEnvVar)
	if baseURL == "" {
		return "https://api.openai.com/v1"
	}

	return baseURL
}

func IsOpenAIConfigured() bool {
	return GetOpenAIAPIKey() != ""
}

func GetGeminiAPIKey() string {
	return os.Getenv(GeminiAPIKeyEnvVar)
}

func IsGeminiConfigured() bool {
	return GetGeminiAPIKey() != ""
}

func IsAnyAIConfigured() bool {
	return IsOpenAIConfigured() || IsGeminiConfigured()
}
