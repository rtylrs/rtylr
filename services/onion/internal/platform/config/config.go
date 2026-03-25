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

const (
	R2AccountIDEnvVar       = "R2_ACCOUNT_ID"
	R2AccessKeyIDEnvVar     = "R2_ACCESS_KEY_ID"
	R2SecretAccessKeyEnvVar = "R2_SECRET_ACCESS_KEY"
	R2RegionEnvVar          = "R2_REGION"
	R2BucketEnvVar          = "R2_BUCKET"
	R2PublicURLEnvVar       = "R2_PUBLIC_URL"
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

func GetR2AccountID() string {
	return os.Getenv(R2AccountIDEnvVar)
}

func GetR2AccessKeyID() string {
	return os.Getenv(R2AccessKeyIDEnvVar)
}

func GetR2SecretAccessKey() string {
	return os.Getenv(R2SecretAccessKeyEnvVar)
}

func GetR2Region() string {
	region := os.Getenv(R2RegionEnvVar)
	if region == "" {
		return "auto"
	}

	return region
}

func GetR2Bucket() string {
	return os.Getenv(R2BucketEnvVar)
}

func GetR2PublicURL() string {
	return os.Getenv(R2PublicURLEnvVar)
}

func IsR2Configured() bool {
	return GetR2AccountID() != "" && GetR2AccessKeyID() != "" && GetR2SecretAccessKey() != "" && GetR2Bucket() != ""
}

const (
	ElevenLabsAPIKeyEnvVar  = "ELEVENLABS_API_KEY"
	ElevenLabsBaseURLEnvVar = "ELEVENLABS_BASE_URL"
)

func GetElevenLabsAPIKey() string {
	return os.Getenv(ElevenLabsAPIKeyEnvVar)
}

func GetElevenLabsBaseURL() string {
	baseURL := os.Getenv(ElevenLabsBaseURLEnvVar)
	if baseURL == "" {
		return "https://api.elevenlabs.io/v1"
	}

	return baseURL
}

func IsElevenLabsConfigured() bool {
	return GetElevenLabsAPIKey() != ""
}

func GetRunPodAPIKey() string {
	return os.Getenv("RUNPOD_API_KEY")
}

func GetRunPodEndpointID() string {
	return os.Getenv("RUNPOD_ENDPOINT_ID")
}

func IsRunPodConfigured() bool {
	return GetRunPodAPIKey() != ""
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
