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
	if GlobalConfig != nil {
		return GlobalConfig.RedisURL
	}
	return os.Getenv(RedisURLEnvVar)
}

const (
	InMemoryCacheMaxSizeMbEnvVar = "INMEMORY_CACHE_MAX_SIZE_MB"
)

// Redis Configuration.
func GetRedisHosts() []string {
	if GlobalConfig != nil {
		return GlobalConfig.RedisHosts
	}
	hosts := os.Getenv(RedisHostsEnvVar)
	if hosts == "" {
		return []string{"localhost:6379"}
	}

	return strings.Split(hosts, ",")
}

func GetRedisUsername() string {
	if GlobalConfig != nil {
		return GlobalConfig.RedisUsername
	}
	return os.Getenv(RedisUsernameEnvVar)
}

func GetRedisPassword() string {
	if GlobalConfig != nil {
		return GlobalConfig.RedisPassword
	}
	return os.Getenv(RedisPasswordEnvVar)
}

func IsRedisEnabled() bool {
	if GlobalConfig != nil {
		return GlobalConfig.RedisEnabled
	}
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
	if GlobalConfig != nil {
		if GlobalConfig.InMemoryCacheMaxSizeMb > 0 {
			return GlobalConfig.InMemoryCacheMaxSizeMb
		}
		return 100
	}
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
