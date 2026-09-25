package config

import (
	"os"
	"strings"
)

const (
	TrinoEnabledEnvVar     = "TRINO_ENABLED"
	TrinoHostEnvVar        = "TRINO_HOST"
	TrinoPortEnvVar        = "TRINO_PORT"
	TrinoUserEnvVar        = "TRINO_USER"
	TrinoPasswordEnvVar    = "TRINO_PASSWORD"
	TrinoAccessTokenEnvVar = "TRINO_ACCESS_TOKEN"
	TrinoCatalogEnvVar     = "TRINO_CATALOG"
	TrinoSchemaEnvVar      = "TRINO_SCHEMA"
	TrinoTLSEnvVar         = "TRINO_TLS"
	TrinoSourceEnvVar      = "TRINO_SOURCE"
)

const (
	DefaultTrinoHost   = "localhost"
	DefaultTrinoPort   = "8080"
	DefaultTrinoSource = "onion"
)

func IsTrinoEnabled() bool {
	if GlobalConfig != nil {
		return GlobalConfig.TrinoEnabled
	}
	return strings.ToLower(strings.TrimSpace(os.Getenv(TrinoEnabledEnvVar))) == "true"
}

func GetTrinoHost() string {
	if GlobalConfig != nil && GlobalConfig.TrinoHost != "" {
		return GlobalConfig.TrinoHost
	}
	if v := strings.TrimSpace(os.Getenv(TrinoHostEnvVar)); v != "" {
		return v
	}
	return DefaultTrinoHost
}

func GetTrinoPort() string {
	if GlobalConfig != nil && GlobalConfig.TrinoPort != "" {
		return GlobalConfig.TrinoPort
	}
	if v := strings.TrimSpace(os.Getenv(TrinoPortEnvVar)); v != "" {
		return v
	}
	return DefaultTrinoPort
}

func GetTrinoUser() string {
	if GlobalConfig != nil {
		return GlobalConfig.TrinoUser
	}
	return strings.TrimSpace(os.Getenv(TrinoUserEnvVar))
}

func GetTrinoPassword() string {
	if GlobalConfig != nil {
		return GlobalConfig.TrinoPassword
	}
	return os.Getenv(TrinoPasswordEnvVar)
}

func GetTrinoAccessToken() string {
	if GlobalConfig != nil {
		return GlobalConfig.TrinoAccessToken
	}
	return os.Getenv(TrinoAccessTokenEnvVar)
}

func GetTrinoCatalog() string {
	if GlobalConfig != nil {
		return GlobalConfig.TrinoCatalog
	}
	return strings.TrimSpace(os.Getenv(TrinoCatalogEnvVar))
}

func GetTrinoSchema() string {
	if GlobalConfig != nil {
		return GlobalConfig.TrinoSchema
	}
	return strings.TrimSpace(os.Getenv(TrinoSchemaEnvVar))
}

func IsTrinoTLS() bool {
	if GlobalConfig != nil {
		return GlobalConfig.TrinoTLS
	}
	return strings.EqualFold(strings.TrimSpace(os.Getenv(TrinoTLSEnvVar)), "true")
}

func GetTrinoSource() string {
	if GlobalConfig != nil && GlobalConfig.TrinoSource != "" {
		return GlobalConfig.TrinoSource
	}
	if v := strings.TrimSpace(os.Getenv(TrinoSourceEnvVar)); v != "" {
		return v
	}
	return DefaultTrinoSource
}

// IsTrinoConfigured reports whether Trino is enabled and has the minimum
// settings needed to open a connection (host + user).
func IsTrinoConfigured() bool {
	if !IsTrinoEnabled() {
		return false
	}
	return GetTrinoHost() != "" && GetTrinoUser() != ""
}
