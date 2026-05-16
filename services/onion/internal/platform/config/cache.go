package config

import (
	"os"
	"strconv"
	"strings"
)

const (
	RedisURLEnvVar      = "REDIS_URL"
	RedisHostsEnvVar    = "REDIS_HOSTS"
	RedisUsernameEnvVar = "REDIS_USERNAME"
	RedisPasswordEnvVar = "REDIS_PASSWORD"
	RedisEnabledEnvVar  = "REDIS_ENABLED"
)

// GetRedisURL returns a single-string Redis connection URL (Upstash style:
// rediss://user:pass@host:port). Empty when not configured.
func GetRedisURL() string {
	return os.Getenv(RedisURLEnvVar)
}

const (
	InMemoryCacheMaxSizeMbEnvVar   = "INMEMORY_CACHE_MAX_SIZE_MB"
	CachedMarshallingEnabledEnvVar = "CACHED_MARSHALLING_ENABLED"
)

// Redis Configuration.
func GetRedisHosts() []string {
	hosts := os.Getenv(RedisHostsEnvVar)
	if hosts == "" {
		return []string{"localhost:6379"}
	}

	return strings.Split(hosts, ",")
}

func GetRedisUsername() string {
	return os.Getenv(RedisUsernameEnvVar)
}

func GetRedisPassword() string {
	return os.Getenv(RedisPasswordEnvVar)
}

func IsRedisEnabled() bool {
	return strings.ToLower(os.Getenv(RedisEnabledEnvVar)) == "true"
}

func IsRedisConfigured() bool {
	if !IsRedisEnabled() {
		return false
	}
	if GetRedisURL() != "" {
		return true
	}
	return len(GetRedisHosts()) > 0
}

// Cache Configuration.
func GetInMemoryCacheMaxSizeMb() int {
	sizeStr := os.Getenv(InMemoryCacheMaxSizeMbEnvVar)
	if sizeStr == "" {
		return 100 // Default 100MB
	}

	size, err := strconv.Atoi(sizeStr)
	if err != nil {
		return 100 // Default on error
	}

	return size
}

func IsCachedMarshallingEnabled() bool {
	return strings.ToLower(os.Getenv(CachedMarshallingEnabledEnvVar)) == "true"
}
