package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// envOrDefault returns the env var value, or fallback when it's unset/empty.
func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// envFirst returns the value of the first set/non-empty env var in keys, or "".
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

func GetOpenCodeAPIKey() string {
	if GlobalConfig != nil {
		return strings.TrimSpace(GlobalConfig.OpenCodeAPIKey)
	}
	return strings.TrimSpace(os.Getenv(OpenCodeAPIKeyEnvVar))
}

func GetOpenCodeBaseURL() string { return OpenCodeBaseURL }
func GetOpenCodeModel() string   { return OpenCodeModel }

func IsOpenCodeConfigured() bool {
	return GetOpenCodeAPIKey() != ""
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

// GetAIPrimaryProvider returns the operator's preferred primary provider
// or "" for the default DeepSeek, Gemini, OpenAI, OpenCode order.
func GetAIPrimaryProvider() string {
	if GlobalConfig != nil {
		return GlobalConfig.AIPrimaryProvider
	}
	return os.Getenv(AIPrimaryProviderEnvVar)
}

func IsAnyAIConfigured() bool {
	return IsDeepSeekConfigured() || IsOpenAIConfigured() || IsGeminiConfigured() || IsOpenCodeConfigured()
}
