package config

import (
	"os"
	"strings"
)

type RunMode string

const (
	APIRunMode RunMode = "default"
	WorkerMode RunMode = "worker"
)

const DefaultRunMode = APIRunMode

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

func GetRunMode() RunMode {
	mode := os.Getenv("runmode")
	if mode == "" {
		return DefaultRunMode
	}

	mode = strings.ToLower(strings.TrimSpace(mode))
	switch mode {
	case string(APIRunMode):
		return APIRunMode
	case string(WorkerMode):
		return WorkerMode
	default:
		return DefaultRunMode
	}
}

func IsValidRunMode(mode string) bool {
	mode = strings.ToLower(strings.TrimSpace(mode))

	return mode == string(APIRunMode) || mode == string(WorkerMode)
}

func GetRunModeString() string {
	return string(GetRunMode())
}

func IsDefaultMode() bool {
	return GetRunMode() == APIRunMode
}

func IsWorkerMode() bool {
	return GetRunMode() == WorkerMode
}
