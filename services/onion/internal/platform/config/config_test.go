package config

import (
	"testing"
	"time"
)

func setConfigEnv(t *testing.T) {
	t.Helper()
	t.Setenv("RUN_MODE", "api")
	t.Setenv("DB_HOST", "127.0.0.1")
	t.Setenv("DB_PORT", "3306")
	t.Setenv("DB_USER", "test")
	t.Setenv("DB_PASSWORD", "test")
	t.Setenv("DB_NAME", "onion_test")
	t.Setenv("JWT_SECRET", "test-secret")
	t.Setenv("JWT_DURATION", "")
	t.Cleanup(func() { GlobalConfig = nil })
}

func TestOpenCodeConfiguration(t *testing.T) {
	setConfigEnv(t)
	GlobalConfig = nil
	for _, key := range []string{DeepSeekAPIKeyEnvVar, GeminiAPIKeyEnvVar, OpenAIAPIKeyEnvVar, OpenCodeAPIKeyEnvVar} {
		t.Setenv(key, "")
	}
	t.Setenv("NVIDIA_NIM_API_KEY", "retired-fixture-key")
	if IsAnyAIConfigured() {
		t.Fatal("AI must remain disabled without a provider key")
	}
	t.Setenv(OpenCodeAPIKeyEnvVar, " \t")
	if IsOpenCodeConfigured() {
		t.Fatal("whitespace is not a configured API key")
	}
	t.Setenv(OpenCodeAPIKeyEnvVar, " test-zen-key ")
	if !IsAnyAIConfigured() || GetOpenCodeAPIKey() != "test-zen-key" {
		t.Fatal("OpenCode should work as the only configured provider")
	}
	cfg, err := LoadAndValidate()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.OpenCodeAPIKey != "test-zen-key" || !IsOpenCodeConfigured() {
		t.Fatal("startup did not load the OpenCode key")
	}
	t.Setenv(OpenCodeAPIKeyEnvVar, "different-key")
	if GetOpenCodeAPIKey() != "test-zen-key" {
		t.Fatal("loaded configuration must take precedence over the environment")
	}
	GlobalConfig = &AppConfig{}
	if IsAnyAIConfigured() {
		t.Fatal("an empty loaded configuration must keep AI disabled")
	}
}

func TestLoadAndValidate_TrimsAndFiltersRedisHosts(t *testing.T) {
	setConfigEnv(t)
	t.Setenv("REDIS_HOSTS", " redis-a:6379, , redis-b:6380, ")

	cfg, err := LoadAndValidate()
	if err != nil {
		t.Fatalf("LoadAndValidate: %v", err)
	}
	want := []string{"redis-a:6379", "redis-b:6380"}
	if len(cfg.RedisHosts) != len(want) {
		t.Fatalf("RedisHosts = %v, want %v", cfg.RedisHosts, want)
	}
	for i := range want {
		if cfg.RedisHosts[i] != want[i] {
			t.Fatalf("RedisHosts = %v, want %v", cfg.RedisHosts, want)
		}
	}
}

func TestGetRedisHostsNormalizesBeforeConfigLoad(t *testing.T) {
	GlobalConfig = nil
	t.Cleanup(func() { GlobalConfig = nil })
	t.Setenv(RedisHostsEnvVar, " redis-a:6379, , redis-b:6380, ")

	got := GetRedisHosts()
	want := []string{"redis-a:6379", "redis-b:6380"}
	if len(got) != len(want) {
		t.Fatalf("GetRedisHosts() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("GetRedisHosts() = %v, want %v", got, want)
		}
	}
}

func TestTrustedProxyCIDRsRequirePeerNetwork(t *testing.T) {
	t.Setenv(TrustedProxyCIDRsEnvVar, "203.0.113.0/24,2001:db8::/32")
	if !IsTrustedProxyIP("203.0.113.7:443") {
		t.Fatal("configured IPv4 proxy was not trusted")
	}
	if !IsTrustedProxyIP("[2001:db8::7]:443") {
		t.Fatal("configured IPv6 proxy was not trusted")
	}
	if IsTrustedProxyIP("198.51.100.7:443") {
		t.Fatal("unconfigured peer was trusted")
	}
}

func TestLoadAndValidate_DefaultsJWTDuration(t *testing.T) {
	setConfigEnv(t)

	cfg, err := LoadAndValidate()
	if err != nil {
		t.Fatalf("LoadAndValidate: %v", err)
	}
	if cfg.JWTDuration != DefaultJWTDuration {
		t.Fatalf("JWTDuration = %v, want %v", cfg.JWTDuration, DefaultJWTDuration)
	}
}

func TestLoadAndValidate_RejectsMalformedJWTDuration(t *testing.T) {
	setConfigEnv(t)
	t.Setenv("JWT_DURATION", "not-a-duration")

	if _, err := LoadAndValidate(); err == nil {
		t.Fatal("expected malformed JWT_DURATION to fail validation")
	}
}

func TestLoadAndValidate_RejectsNonPositiveJWTDuration(t *testing.T) {
	setConfigEnv(t)
	t.Setenv("JWT_DURATION", "0s")

	if _, err := LoadAndValidate(); err == nil {
		t.Fatal("expected non-positive JWT_DURATION to fail validation")
	}
}

func TestLoadAndValidate_AcceptsShortJWTSecret(t *testing.T) {
	setConfigEnv(t)
	t.Setenv("JWT_SECRET", "short")
	t.Setenv("JWT_DURATION", (24 * time.Hour).String())

	if _, err := LoadAndValidate(); err != nil {
		t.Fatalf("short JWT_SECRET should remain accepted: %v", err)
	}
}

func TestPositiveEnvIntOrDefaultRejectsNonPositiveCaps(t *testing.T) {
	t.Setenv("CDN_MAX_UPLOAD_MB", "0")
	if got := positiveEnvIntOrDefault("CDN_MAX_UPLOAD_MB", 100); got != 100 {
		t.Fatalf("zero cap = %d, want fallback 100", got)
	}
	t.Setenv("CDN_MAX_UPLOAD_MB", "-5")
	if got := positiveEnvIntOrDefault("CDN_MAX_UPLOAD_MB", 100); got != 100 {
		t.Fatalf("negative cap = %d, want fallback 100", got)
	}
	t.Setenv("CDN_MAX_UPLOAD_MB", "25")
	if got := positiveEnvIntOrDefault("CDN_MAX_UPLOAD_MB", 100); got != 25 {
		t.Fatalf("positive cap = %d, want 25", got)
	}
}
