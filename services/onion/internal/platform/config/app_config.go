package config

import (
	"fmt"
	"os"
	"strings"
)

type AppConfig struct {
	DatabaseURL string
	DBHost      string
	DBPort      string
	DBUser      string
	DBPass      string
	DBName      string
	JWTSecret   string
	ResendKey   string
	AppEnv      string
	Port        string
}

var GlobalConfig *AppConfig

// LoadAndValidate reads the environment configuration at startup,
// validating the presence of critical variables.
func LoadAndValidate() (*AppConfig, error) {
	cfg := &AppConfig{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		DBHost:      os.Getenv("DB_HOST"),
		DBPort:      os.Getenv("DB_PORT"),
		DBUser:      os.Getenv("DB_USER"),
		DBPass:      os.Getenv("DB_PASSWORD"),
		DBName:      os.Getenv("DB_NAME"),
		JWTSecret:   os.Getenv("JWT_SECRET"),
		ResendKey:   os.Getenv("RESEND_API_KEY"),
		AppEnv:      os.Getenv("APP_ENV"),
		Port:        os.Getenv("PORT"),
	}

	if cfg.Port == "" {
		cfg.Port = "9999"
	}

	var missing []string

	// Validate critical environment variables
	if cfg.DatabaseURL == "" && (cfg.DBHost == "" || cfg.DBName == "" || cfg.DBUser == "") {
		missing = append(missing, "DATABASE_URL (or DB_HOST, DB_NAME, DB_USER)")
	}
	if cfg.JWTSecret == "" {
		missing = append(missing, "JWT_SECRET")
	}

	if len(missing) > 0 {
		return nil, fmt.Errorf("missing critical environment variable(s): %s", strings.Join(missing, ", "))
	}

	GlobalConfig = cfg
	return cfg, nil
}
