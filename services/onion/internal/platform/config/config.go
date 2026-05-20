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

func IsAnyAIConfigured() bool {
	return IsOpenAIConfigured() || IsGeminiConfigured()
}
