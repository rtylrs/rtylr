package config

import (
	"bufio"
	"os"
	"strings"
)

// LoadDotEnv reads a .env file from the working directory and sets any key=value
// pairs as environment variables. Missing or unreadable .env files are silently
// ignored — in production, variables are injected by the container runtime.
// Existing environment variables are never overwritten.
func LoadDotEnv() {
	f, err := os.Open(".env")
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.Trim(strings.TrimSpace(parts[1]), `"'`)
		if key == "" || os.Getenv(key) != "" {
			continue
		}
		_ = os.Setenv(key, val)
	}
}

// ListenPort returns the HTTP listen port for this process. PORT from the
// environment wins when set; otherwise fallback is used (9999 for API and 8080
// for CDN upload, for example).
func ListenPort(fallback string) string {
	if v := strings.TrimSpace(os.Getenv("PORT")); v != "" {
		return v
	}
	return fallback
}
